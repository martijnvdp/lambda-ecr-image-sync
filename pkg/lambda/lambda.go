package lambda

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// LambdaEvent lambda input event data, fields have to be exported
type LambdaEvent struct {
	Action             string   `json:"action"` // s3 or sync
	CheckDigest        bool     `json:"check_digest"`
	Concurrent         int      `json:"concurrent"` // number of concurrent syncs
	Repositories       []string `json:"repositories"`
	MaxResults         int      `json:"max_results"`
	SlackChannelID     string   `json:"slack_channel_id"`
	SlackErrorsOnly    bool     `json:"slack_errors_only"`
	SlackMSGErrSubject string   `json:"slack_msg_err_subject"`
	SlackMSGHeader     string   `json:"slack_msg_header"`
	SlackMSGSubject    string   `json:"slack_msg_subject"`
}

type inputRepository struct {
	constraint   string
	ecrImageName string
	excludeRLS   []string
	excludeTags  []string
	source       string
	includeRLS   []string
	includeTags  []string
	maxResults   int
	releaseOnly  bool
}

type process struct {
	wg  *sync.WaitGroup
	mu  *sync.Mutex
	svc *ecrClient
}

type response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

type environmentVars struct {
	awsAccount      string
	awsBucket       string
	awsRegion       string
	slackOAuthToken string
}

var tmpDir string

// getEnvironmentVars returns environment variables
func getEnvironmentVars() (vars environmentVars, err error) {
	vars = environmentVars{
		awsRegion:       os.Getenv("AWS_REGION"),
		awsBucket:       os.Getenv("BUCKET_NAME"),
		awsAccount:      os.Getenv("AWS_ACCOUNT_ID"),
		slackOAuthToken: os.Getenv("SLACK_OAUTH_TOKEN"),
	}

	if vars.awsRegion == "" || vars.awsAccount == "" {
		return vars, fmt.Errorf("error no environment variables set")
	}
	return vars, err
}

// init creates a temp directory for the lambda function
func init() {
	var err error
	tmpDir, err = os.MkdirTemp("", "")

	if err != nil {
		log.Fatal(err)
	}
}

// returnErr returns an error and sends a slack notification
func returnErr(err error, slackOAuthTokenm, slackChannelID, errSubject, errText string) (response, error) {
	sendSlackNotification(slackOAuthTokenm, slackChannelID, errSubject, fmt.Sprint(err))
	errMessage := fmt.Sprintf(errText+" %f", err)
	log.Print(errSubject)
	log.Print(errMessage)
	return response{
		Message: errMessage,
		Ok:      false,
	}, err
}

// ecrRepoNamesFromAWSARNs returns a slice of repository names from a slice of AWS ARNs
func ecrRepoNamesFromAWSARNs(arns []string, region, account string) []string {
	var names []string
	for _, arn := range arns {
		names = append(names, strings.Replace(arn, "arn:aws:ecr:"+region+":"+account+":repository/", "", -1))
	}
	return names
}

func (proc *process) processRepositories(repositories []inputRepository, max, maxResults int, checkDigest bool, environmentVars environmentVars) (allTagsToSync []syncOptions) {
	totalItems := len(repositories)

	for i := 0; i < totalItems; i += max {
		limit := max
		if i+max > totalItems {
			limit = totalItems - i
		}

		proc.wg.Add(limit)
		for j := 0; j < limit; j++ {
			repo := repositories[i+j]
			log.Printf("Processing repository: %s", repo.source)
			go func(j int) {
				defer proc.wg.Done()
				tagsToSync, err := proc.svc.getTagsToSync(&repo, repo.ecrImageName, maxResults, checkDigest, environmentVars)
				if err != nil {
					log.Fatal(err)
				}
				proc.mu.Lock()
				if len(tagsToSync.tags) > 0 {
					allTagsToSync = append(allTagsToSync, tagsToSync)
				}
				proc.mu.Unlock()
			}(j)
		}
		proc.wg.Wait()
	}
	return allTagsToSync
}

func (proc *process) processTags(allTagsToSync []syncOptions, max int, environmentVars environmentVars) (total int, syncErrors []error) {
	totalItems := len(allTagsToSync)

	for i := 0; i < totalItems; i += max {
		limit := max
		if i+max > totalItems {
			limit = totalItems - i
		}

		proc.wg.Add(limit)
		for j := 0; j < limit; j++ {
			tags := allTagsToSync[i+j]
			log.Printf("Syncing image: %s", tags.source)
			go func(j int) {
				defer proc.wg.Done()
				err := proc.svc.syncImages(tags, environmentVars)
				if err != nil {
					proc.mu.Lock()
					syncErrors = append(syncErrors, err)
					proc.mu.Unlock()
					log.Fatal(err)
				}
			}(j)
			proc.mu.Lock()
			total = total + len(tags.tags)
			proc.mu.Unlock()
		}
		proc.wg.Wait()
	}
	return total, syncErrors
}

// Start Lambda Function for syncing ecr images with public repositories, outputs csv with needed images to S3 bucket.
func Start(ctx context.Context, event LambdaEvent) (response, error) {
	const (
		imagesZipFile = "images.zip"
		imagesCSVFile = "images.csv"
	)

	var (
		csvContent   []csvFormat
		syncErrors   []error
		dockerConfig = filepath.Dir(tmpDir)
		errSubject   = tryString(event.SlackMSGErrSubject, "The following error has occurred during the lambda ecr-image-sync:")
		repositories []inputRepository
		total        int
	)
	environmentVars, err := getEnvironmentVars()

	zipFile := filepath.Join(tmpDir, imagesZipFile)
	csvFile := filepath.Join(tmpDir, imagesCSVFile)
	os.Setenv("DOCKER_CONFIG", dockerConfig)

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
			"Error reading environment variables , or not set:")
	}
	svc, err := newEcrClient(environmentVars.awsRegion)

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
			"Error creating ECR client:")
	}

	if os.Getenv("DOCKER_USERNAME") != "" && os.Getenv("DOCKER_PASSWORD") != "" {
		err = login(loginOptions{
			serverAddress: "docker.io",
			user:          os.Getenv("DOCKER_USERNAME"),
			password:      os.Getenv("DOCKER_PASSWORD"),
		})
		if err != nil {
			log.Println("error logging in to docker.io: ", err)
		}
	}
	names := ecrRepoNamesFromAWSARNs(event.Repositories, environmentVars.awsRegion, environmentVars.awsAccount)
	repositories, err = svc.getinputRepositorysFromTags(names)

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
			"Error getting input images from tags")
	}

	log.Printf("Starting lambda for %s repositories", strconv.Itoa(len(repositories)))
	maxConcurrent := maxInt(event.Concurrent, 1)

	proc := process{
		wg:  &sync.WaitGroup{},
		mu:  &sync.Mutex{},
		svc: svc,
	}
	allTagsToSync := proc.processRepositories(repositories, maxConcurrent, event.MaxResults, event.CheckDigest, environmentVars)

	switch {
	case event.Action == "s3":
		csvContent, total, err = buildCSVFile(allTagsToSync, environmentVars)
		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
				"Error building csv output:")
		}
	default:
		err = proc.svc.authToECR(environmentVars)
		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
				"Error authenticating to ECR:")
		}
		total, syncErrors = proc.processTags(allTagsToSync, maxConcurrent, environmentVars)
		for _, err := range syncErrors {
			return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
				"Error syncing repositories:")
		}

	}

	resultMessage := fmt.Sprintf("Successfully synced %s images to the ecr", strconv.Itoa(total))

	if csvContent != nil && event.Action == "s3" && environmentVars.awsBucket != "" {

		if err := outputToS3Bucket(csvContent, csvFile, zipFile, environmentVars.awsRegion, environmentVars.awsBucket); err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
				"Error while writing zip file to the S3 Bucket with error:")
		}

		if !event.SlackErrorsOnly {
			sendResultsToSlack(event.SlackMSGHeader, event.SlackMSGSubject, &csvContent, environmentVars.slackOAuthToken, event.SlackChannelID)
		}
		resultMessage = fmt.Sprintf("Successfully added %s images to the csv", strconv.Itoa(total))
	}

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, event.SlackChannelID, errSubject,
			"Lambda function resulted with error:")
	}
	log.Print(resultMessage)

	return response{
		Message: resultMessage,
		Ok:      true,
	}, nil
}
