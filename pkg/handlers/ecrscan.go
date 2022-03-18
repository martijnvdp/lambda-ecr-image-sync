package handlers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type ecrResults struct {
	name string
	tag  string
	hash string
}

func (i *inputImage) getImagesFromECR(ecrImageName, ecrRepoPrefix, region string) (results map[string]ecrResults, err error) {
	var ecrResult *ecr.ListImagesOutput
	results = make(map[string]ecrResults)
	session, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	ecrRepoPrefix = tryString(ecrRepoPrefix, i.EcrRepoPrefix)

	if err != nil {
		fmt.Println("failed to create session,", err)
		return nil, err
	}
	svc := ecr.New(session)

	input := &ecr.ListImagesInput{
		RepositoryName: aws.String((ecrRepoPrefix + "/" + ecrImageName)),
		Filter: &ecr.ListImagesFilter{
			TagStatus: aws.String("TAGGED"),
		},
	}
	ecrResult, err = svc.ListImages(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case ecr.ErrCodeServerException:
				fmt.Println(ecr.ErrCodeServerException, awsErr.Error())
			case ecr.ErrCodeInvalidParameterException:
				fmt.Println(ecr.ErrCodeInvalidParameterException, awsErr.Error())
			case ecr.ErrCodeRepositoryNotFoundException:
				fmt.Println(ecr.ErrCodeRepositoryNotFoundException, awsErr.Error())
			default:
				fmt.Println(awsErr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	for _, id := range ecrResult.ImageIds {
		results[i.ImageName+":"+*id.ImageTag] = ecrResults{
			name: ecrImageName,
			tag:  *id.ImageTag,
			hash: *id.ImageDigest,
		}
	}

	return results, err
}
