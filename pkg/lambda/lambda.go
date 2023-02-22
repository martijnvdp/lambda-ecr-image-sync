package lambda

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LambdaEvent lambda input event data, fields have to be exported
type LambdaEvent struct {
	Action             string       `json:"action"` // s3 or sync
	CheckDigest        bool         `json:"check_digest"`
	EcrRepoPrefix      string       `json:"ecr_repo_prefix"`
	Images             []InputImage `json:"images"`
	MaxResults         int          `json:"max_results"`
	SlackChannelID     string       `json:"slack_channel_id"`
	SlackErrorsOnly    bool         `json:"slack_errors_only"`
	SlackMSGErrSubject string       `json:"slack_msg_err_subject"`
	SlackMSGHeader     string       `json:"slack_msg_header"`
	SlackMSGSubject    string       `json:"slack_msg_subject"`
}

type InputImage struct {
	Constraint    string   `json:"constraint"`
	EcrRepoPrefix string   `json:"repo_prefix"`
	ExcludeRLS    []string `json:"exclude_rls"`
	ExcludeTags   []string `json:"exclude_tags"`
	ImageName     string   `json:"image_name"`
	IncludeRLS    []string `json:"include_rls"`
	IncludeTags   []string `json:"include_tags"`
	MaxResults    int      `json:"max_results"`
	ReleaseOnly   bool     `json:"release_only"`
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

func getEcrImageName(imageName string) string {
	split := strings.Split(imageName, "/")
	switch {
	case len(split) == 3:
		return (split[1] + "/" + split[2])
	case strings.Contains(split[0], "."):
		return split[1]
	default:
		return (split[0] + "/" + split[1])
	}
}

func init() {
	var err error
	tmpDir, err = ioutil.TempDir("", "")

	if err != nil {
		log.Fatal(err)
	}
}

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

// Lambda Function for syncing ecr images with public repositories, outputs csv with needed images to S3 bucket.
func Start(ctx context.Context, lambdaEvent LambdaEvent) (response, error) {
	count := 0
	var csvContent []csvFormat
	zipFile := filepath.Join(tmpDir, "images.zip")
	csvFile := filepath.Join(tmpDir, "images.csv")
	dockercfg := filepath.Join(tmpDir, "config.json")
	environmentVars, err := getEnvironmentVars()
	errSubject := tryString(lambdaEvent.SlackMSGErrSubject, "The following error has occurred during the lambda ecr-image-sync:")
	os.Setenv("DOCKER_CONFIG", dockercfg)

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Error reading environment variables , or not set:")
	}
	svc, err := newEcrClient(environmentVars.awsRegion)

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Error creating ECR client:")
	}

	// get input images from ECR Repositories which are configured to be synced using tags
	images, err := svc.getInputImagesFromTags()

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Error getting input images from tags:")
	}
	totalImages := append(images, lambdaEvent.Images...)

	for _, i := range totalImages {
		ecrImageName := getEcrImageName(i.ImageName)
		resultsFromEcr, err := svc.getImagesFromECR(ecrImageName, lambdaEvent.EcrRepoPrefix, environmentVars.awsRegion, &i)

		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error searching image: "+i.ImageName+" on private ECR repository:")
		}
		tagsFromPublicRepo, err := i.getTagsFromPublicRepo()

		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error getting tags of image: "+i.ImageName+" from public repo:")
		}
		resultsFromPublicRepo, err := i.checkTagsFromPublicRepo(&tagsFromPublicRepo, lambdaEvent.MaxResults)

		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error while checking tags and constraints of image: "+i.ImageName+" :")
		}

		if lambdaEvent.CheckDigest {
			resultsFromPublicRepo, err = checkDigest(i.ImageName, &resultsFromPublicRepo, &resultsFromEcr)
		} else {
			resultsFromPublicRepo, err = checkNoDigest(i.ImageName, &resultsFromPublicRepo, &resultsFromEcr)
		}

		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error while comparing digest of image: "+i.ImageName+" with digest from ecr:")
		}

		if lambdaEvent.Action != "s3" {
			count, err = syncImages(i.ImageName, ecrImageName, i.EcrRepoPrefix, &resultsFromPublicRepo, environmentVars.awsAccount, environmentVars.awsRegion)
			if err != nil {
				return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
					"Error syncing images:")
			}
		} else {
			csvOutput, err := buildCSVFile(i.ImageName, ecrImageName, i.EcrRepoPrefix, &resultsFromPublicRepo, environmentVars.awsAccount, environmentVars.awsRegion)
			if err != nil {
				return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
					"Error building csv output:")
			}
			csvContent = append(csvContent, csvOutput...)
		}
		count = count + len(resultsFromPublicRepo)
	}

	if csvContent != nil && lambdaEvent.Action == "s3" && environmentVars.awsBucket != "" {

		if err := outputToS3Bucket(csvContent, csvFile, zipFile, environmentVars.awsRegion, environmentVars.awsBucket); err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error writing output to s3 bucket")
		}
	}

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Lambda function resulted with error:")
	}
	resultMessage := fmt.Sprintf("Successfully synced %q images to the ecr", strconv.Itoa(count))
	log.Print(resultMessage)

	return response{
		Message: resultMessage,
		Ok:      true,
	}, nil
}