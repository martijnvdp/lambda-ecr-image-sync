package handlers

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
	CheckDigest        bool         `json:"check_digest"`
	EcrRepoPrefix      string       `json:"ecr_repo_prefix"`
	Images             []inputImage `json:"images"`
	MaxResults         int          `json:"max_results"`
	SlackChannelID     string       `json:"slack_channel_id"`
	SlackErrorsOnly    bool         `json:"slack_errors_only"`
	SlackMSGErrSubject string       `json:"slack_msg_err_subject"`
	SlackMSGHeader     string       `json:"slack_msg_header"`
	SlackMSGSubject    string       `json:"slack_msg_subject"`
}

type inputImage struct {
	Constraint    string   `json:"constraint"`
	EcrRepoPrefix string   `json:"repo_prefix"`
	ExcludeRLS    []string `json:"exclude_rls"`
	ExcludeTags   []string `json:"exclude_tags"`
	ImageName     string   `json:"image_name"`
	IncludeRLS    []string `json:"include_rls"`
	IncludeTags   []string `json:"include_tags"`
	MaxResults    int      `json:"max_results"`
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

	if vars.awsRegion == "" || vars.awsBucket == "" || vars.awsAccount == "" {
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

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
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
	return response{
		Message: fmt.Sprintf(errText+" %f", err),
		Ok:      false,
	}, err
}

func tryString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

// Lambda Function for syncing ecr images with public repositories, outputs csv with needed images to S3 bucket.
func Lambda(ctx context.Context, lambdaEvent LambdaEvent) (response, error) {
	var csvContent []csvFormat
	zipFile := filepath.Join(tmpDir, "images.zip")
	csvFile := filepath.Join(tmpDir, "images.csv")
	environmentVars, err := getEnvironmentVars()
	errSubject := tryString(lambdaEvent.SlackMSGErrSubject, "The following error has occurred during the lambda ecr-image-sync:")

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Error reading environment variables , or not set:")
	}

	for _, i := range lambdaEvent.Images {
		ecrImageName := getEcrImageName(i.ImageName)
		resultsFromEcr, err := i.getImagesFromECR(ecrImageName, lambdaEvent.EcrRepoPrefix, environmentVars.awsRegion)

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
		csvOutput, err := buildCSVFile(i.ImageName, ecrImageName, i.EcrRepoPrefix, &resultsFromPublicRepo, environmentVars.awsAccount, environmentVars.awsRegion)

		if err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error building csv output:")

		}
		csvContent = append(csvContent, csvOutput...)
	}

	if csvContent != nil {

		if err := writeCSVFile(&csvContent, csvFile); err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error writing to local storrage:")
		}

		if err := createZipFile(csvFile, zipFile); err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error while compesssing csv to zip file with error:")
		}

		if err := addFileToS3Bucket(zipFile, environmentVars.awsRegion, environmentVars.awsBucket); err != nil {
			return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
				"Error while writing zip file to the S3 Bucket with error:")
		}

		if !lambdaEvent.SlackErrorsOnly {
			sendResultsToSlack(lambdaEvent.SlackMSGHeader, lambdaEvent.SlackMSGSubject, &csvContent, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID)
		}
		defer os.Remove(csvFile)
		defer os.Remove(zipFile)
	}

	if err != nil {
		return returnErr(err, environmentVars.slackOAuthToken, lambdaEvent.SlackChannelID, errSubject,
			"Lambda function resulted with error:")
	}

	return response{
		Message: fmt.Sprintf("Successfully added %q images to csv", strconv.Itoa(len(csvContent))),
		Ok:      true,
	}, nil
}
