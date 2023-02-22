package main

import (
	"context"
	"os"

	ecrImageSync "github.com/martijnvdp/lambda-ecr-image-sync/pkg/lambda"
)

func main() {
	os.Setenv("AWS_REGION", "eu-west-1")
	os.Setenv("AWS_ACCOUNT_ID", "1234")
	InputImage1 := ecrImageSync.InputImage{
		Constraint:    ">= 7.0.4",
		EcrRepoPrefix: "base/infra",
		ImageName:     "docker.io/redis",
		ReleaseOnly:   true,
		IncludeRLS:    []string{"alpine"},
	}
	InputImage2 := ecrImageSync.InputImage{
		Constraint:    ">= 0.66.0",
		EcrRepoPrefix: "base/infra",
		ImageName:     "docker.io/otel/opentelemetry-collector-contrib",
	}
	lambdaEvent := ecrImageSync.LambdaEvent{
		CheckDigest: true,
		Images:      []ecrImageSync.InputImage{InputImage1, InputImage2},
		MaxResults:  5,
	}

	ecrImageSync.Start(context.Background(), lambdaEvent)
}
