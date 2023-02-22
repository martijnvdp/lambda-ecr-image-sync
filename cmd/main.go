package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	ecrImageSync "github.com/martijnvdp/lambda-ecr-image-sync/pkg/lambda"
)

func main() {
	lambda.Start(ecrImageSync.Start)
}
