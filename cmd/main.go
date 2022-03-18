package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/martijnvdp/lambda-ecr-image-sync/pkg/handlers"
)

func main() {
	lambda.Start(handlers.Lambda)
}
