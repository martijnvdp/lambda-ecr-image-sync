[![goreleaser](https://github.com/martijnvdp/lambda-ecr-image-sync/actions/workflows/go.yml/badge.svg)](https://github.com/martijnvdp/lambda-ecr-image-sync/actions/workflows/go.yml)
# ecr-image-sync

Golang Lambda function to compare images between ECR and Public Repositories as dockerHub, quay.io, gcr.io 
and creates a CSV in an S3 bucket with the missing images/tags to be synced.

The function will compare the given images and tags between ECR and the public registry and places the missing images in a CSV file on in an S3 bucket for CodePipeline to pick up and synchronize the missing images mentioned in the CSV.
compatible with most container registry's, see for more info the container lib https://github.com/containers/image

## Docker images

- `docker pull martijnvdp/ecr-image-sync:latest`
- `docker pull martijnvdp/ecr-image-sync:v0.1.7`

## usage

Create a lambda function using the container image (pushed to ecr) and set runtime at go1.x.`\
Set environment variables in the lambda configuration section. \
https://github.com/martijnvdp/terraform-ecr-image-sync

Image names format:
(registry hostname)/imagename/name

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
SLACK_OAUTH_TOKEN='Slack oath token for notifications'
```

Lambda event data:

```hcl
{
"ecr_repo_prefix":"base/images" \\optional global_ecr_repo_prefix
"images": [
      {
        "constraint": "~>2.0" 
        "exclude_rls": ["beta","rc"] \\ excluded pre-releases matches the text after eg: 1.1-beta beta
        "exclude_tags": [],
        "image_name": "docker.io/hashicorp/tfc-agent",
        "include_tags": [],
        "include_rls": ["linux","debian","cee"] \\ included pre-releases matches the text after eg: 1.1-beta beta
        "max_results": 10
        "repo_prefix": "ecr/cm" 
      },  
      {
        "exclude_tags": [],
        "image_name": "docker.io/datadog/agent",
        "include_tags": ["latest","6.27.0-rc.6"],
        "repo_prefix": "ecr/cm"
      }
    ]
"check_digest": true // check digest of existing tags on ecr and only add tags if the digest is not the same
"max_results": 5
"slack_channel_id":"CDDF324"
"slack_errors_only": true // only return errors to slack
"slack_msg_err_subject":"The following error has occurred:"
"slack_msg_header":"The Lambda ECR-IMAGE-SYNC has completed"
"slack_msg_subject":"The following images are now synced to ECR:"
  }
```

## Versions 

use constraint for version constraints 

examples:
```hcl
"constraint": "~> v3.0"
"constraint": "=> v3.0, < v5.0"
"constraint": "= v3.0"
```

use include_rls to include certain keywords/pre-releases:

Prerelease info is everything after the -

Example for v1.2-beta-10 it is beta and 10
to include beta pre-releases: 

```hcl
"include_rls": ["beta"]
```
to exclude beta pre-releases: 

```hcl
"exclude_rls": ["beta"]
```

to include debian builds but exclude release candidates,alpha or beta 

v1.2.3-debian-1-rc

```hcl
"include_rls": ["debian"]
"exclude_rls": ["rc","beta","alpha"]
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

## before a pr

before submitting a pr execute:
```
make pre-pr
```

## test

Before testing aws functions make sure the aws environment vars are set.
```
make test
```

## test in visual studio code
prerequisites

install delve:
```
go get -u github.com/go-delve/delve
go get -u github.com/go-delve/delve/cmd/dlv

```

## Create Test for functions
With the gotests tool you can auto generate go tests for new functions:
https://github.com/cweill/gotests

```
make tests
```

Set breakpoints and run the launch tests debug

## Check code
```
make test
make quality
make cyclo
make pre-pr
```

## references
* https://github.com/containers/skopeo
* https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html
* https://github.com/pgarbe/cdk-ecr-sync
* https://garbe.io/blog/2020/04/22/cdk-ecr-sync/
* https://github.com/cweill/gotests 
* https://github.com/docker-slim/docker-slim

### used modules
* https://github.com/containers/image
* https://github.com/nikoksr/notify

## cloned modules
* https://github.com/hashicorp/go-version@v1.3.0