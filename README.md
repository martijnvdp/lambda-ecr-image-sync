# Lambda-ecr-image-sync
![docker build](https://github.com/martijnvdp/lambda-ecr-image-sync/actions/workflows/release-docker-slim.yml/badge.svg)

This is a Golang Lambda function that compares images between ECR and public repositories such as DockerHub, Quay.io, and GCR.io and synces/copies the missing images to the ECR. It has the capability to sync the images directly to the target ECR on AWS or output a zipped CSV file with the missing images/tags to an S3 bucket.

The function compares the provided images and tags between ECR and the public registry using the Crane library to login and copy the missing images to the ECR on AWS. 
This function is compatible with most container registries. For more information, please refer to the container lib at https://github.com/containers/image.

## Docker images

- `docker pull ghcr.io/martijnvdp/lambda-ecr-image-sync:v1.0.3`

## usage

Create a lambda function using the container image (pushed to ecr) and set runtime at go1.x.`\
Set environment variables in the lambda configuration section. \
https://github.com/martijnvdp/terraform-ecr-image-sync

Image names format:
(registry hostname)/Source/name

```hcl
docker.io/datadog/agent
gcr.io/datadoghq/agent
quay.io/cilium/cilium
```

Environment variables:

```hcl
AWS_ACCOUNT_ID='12345'
AWS_REGION='eu-west-1'
BUCKET_NAME='bucket_name'
DOCKER_USERNAME='optional Username for docker hub'
DOCKER_PASSWORD='optional Password for docker hub'
SLACK_OAUTH_TOKEN='Slack oath token for notifications'
```

Lambda event data:

```hcl
{
"repositories": [ // optional if not specified it wil syn call repos that are configured with tags
  "arn:aws:ecr:us-east-1:123456789012:repository/dev/datadog/datadog-operator","arn:aws:ecr:us-east-1:123456789012:repository/dev/datadog/datadog"]
"check_digest": true // check digest of existing tags on ecr and only add tags if the digest is not the same
"concurrent": 2 // max number of concurrent jobs
"max_results": 5
"slack_channel_id":"CDDF324"
"slack_errors_only": true // only return errors to slack
"slack_msg_err_subject":"The following error has occurred:"
"slack_msg_header":"The Lambda ECR-IMAGE-SYNC has completed"
"slack_msg_subject":"The following images are now synced to ECR:"
  }
```

## configure ECR Sync with tags on the internal ECR Repository
Repository tags:
```
ecr_sync_constraint = "-ge v1.1.1" // equivalent of >= v1.1.1 other operators ( -gt -le -lt) because >= chars is not allowed in aws tags
ecr_sync_source = "docker.io/owner/image"
ecr_sync_include_rls = "ubuntu rc" // releases to include v.1.2-ubuntu v1.2-RC-1
ecr_sync_release_only = "true" // only release version exclude normal tags
ecr_sync_max_results = "10"
ecr_sync_exclude_rls = "RC UBUNTU" // exclude certain releases 
ecr_sync_exclude_tags = "1.1.1 2.2.2" // exclude specific tags
ecr_sync_include_tags = "1.1.1 2.2.2" // exclude specific tags
```
## Versions 

use constraint for version constraints 

examples:
```hcl
"constraint": "-ge v3.0"
"constraint": "-gt v3.0"
"constraint": "-le v3.0"
"constraint": "-lt v3.0"

```

use include_rls to include certain keywords/pre-releases:

Prerelease info is everything after the -

Example for v1.2-beta-10 it is beta and 10
to include beta pre-releases: 

```hcl
"include_rls": "beta"
```
to exclude beta pre-releases: 

```hcl
"exclude_rls": "beta"
```

to include debian builds but exclude release candidates,alpha or beta 

v1.2.3-debian-1-rc

```hcl
"include_rls": "debian"
"exclude_rls": "rc beta alpha"
```

See for more info:
https://github.com/hashicorp/go-version

## Slack notifications

The function can send notifications to a slack channel:

preparation:
* Create a new Slack App
* Give your bot/app the following OAuth permission scopes: chat:write, chat:write.public
* Copy your Bot User OAuth Access Token for the environment variable in the lambda function
* Copy the Channel ID of the channel you want to post a message to. You can grab the Channel ID by right clicking a channel and selecting * copy link. Your Channel ID will be in that link.

Now you can use the fields in the Lambda event payload to set the channel id , message header and subject.

```hcl
"slack_channel_id":"CDDF324"
"slack_errors_only": true // only return errors to slack
"slack_err_msg_subject": "subject error messages"
"slack_msg_header":"ECR-IMAGE-SYNC has completed"
"slack_msg_subject":"The following images are now synced to ECR:"
```

The Token needs to be set as an environment variable in the lambda function configuration
```hcl
SLACK_OAUTH_TOKEN = "OAuth token"
```
you can use go test in slack_test.go to test with a test message

used module https://github.com/nikoksr/notify/blob/main/service/slack/usage.md

## Build requirements

To install the gcc and other dependencys execute:

```
make init

```

## Create Test for functions
With the gotests tool you can auto generate go tests for new functions:
https://github.com/cweill/gotests


## references
* https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html
* https://github.com/pgarbe/cdk-ecr-sync
* https://garbe.io/blog/2020/04/22/cdk-ecr-sync/
* https://github.com/cweill/gotests 
* https://github.com/docker-slim/docker-slim

### used modules
* https://github.com/google/go-containerregistry
* https://github.com/nikoksr/notify
* https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane.md

## cloned modules
* https://github.com/hashicorp/go-version@v1.3.0
