package main

import (
	"context"
	"os"

	ecrImageSync "github.com/martijnvdp/lambda-ecr-image-sync/pkg/lambda"
)

func main() {
	os.Setenv("AWS_REGION", "eu-west-1")
	os.Setenv("AWS_ACCOUNT_ID", "1234")

	lambdaEvent := ecrImageSync.LambdaEvent{
		Concurrent:  2,
		CheckDigest: true,
		MaxResults:  5,
	}

	ecrImageSync.Start(context.Background(), lambdaEvent)
}
