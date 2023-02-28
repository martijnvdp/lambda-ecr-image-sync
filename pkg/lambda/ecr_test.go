package lambda

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type mockECRClient struct {
	ecriface.ECRAPI
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// mock repositories
func (m *mockECRClient) DescribeRepositories(input *ecr.DescribeRepositoriesInput) (*ecr.DescribeRepositoriesOutput, error) {
	// Mock the DescribeRepositories method to return test data
	output := &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{
			{
				RepositoryName: aws.String("base/infra/datadog/datadog-operator"),
				RepositoryArn:  aws.String("arn:aws:ecr:us-east-1:123456789012:repository/base/infra/datadog/datadog-operator"),
				CreatedAt:      aws.Time(parseTime("2022-02-17T13:00:00Z")),
				ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
					ScanOnPush: aws.Bool(true),
				},
			},
			{
				RepositoryName: aws.String("base/infra/gatekeeper/gatekeeper"),
				RepositoryArn:  aws.String("arn:aws:ecr:us-east-1:123456789012:repository/base/infra/gatekeeper/gatekeeper"),
				CreatedAt:      aws.Time(parseTime("2022-02-17T14:00:00Z")),
			},
		},
	}
	return output, nil
}

// mock ListImages method to return test data
func (m *mockECRClient) ListImages(input *ecr.ListImagesInput) (*ecr.ListImagesOutput, error) {
	output := &ecr.ListImagesOutput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageDigest: aws.String("sha256:1234567890123456789012345678901234567890123456789012345678901234"),
				ImageTag:    aws.String("v1.1.1"),
			},
			{
				ImageDigest: aws.String("sha256:1234567890123456789012345678901234567890123456789012345678901234"),
				ImageTag:    aws.String("v1.1.2"),
			},
			{
				ImageDigest: aws.String("sha256:1234567890123456789012345678901234567890123456789012345678901234"),
				ImageTag:    aws.String("v1.1.3"),
			},
		},
	}
	return output, nil
}

// mock tags for the repository
func (m *mockECRClient) ListTagsForResource(*ecr.ListTagsForResourceInput) (*ecr.ListTagsForResourceOutput, error) {
	output := &ecr.ListTagsForResourceOutput{
		Tags: []*ecr.Tag{
			{
				Key:   aws.String("ecr_sync_opt"),
				Value: aws.String("in"),
			},
			{
				Key:   aws.String("ecr_sync_max_results"),
				Value: aws.String("10"),
			},
			{
				Key:   aws.String("ecr_sync_constraint"),
				Value: aws.String("-gt v1.1.1"),
			},
			{
				Key:   aws.String("ecr_sync_source"),
				Value: aws.String("docker.io/datadog/operator"),
			},
		},
	}
	return output, nil
}

func Test_parseinputImageFromTags(t *testing.T) {
	type args struct {
		repo string
		tags map[string]string
	}
	tests := []struct {
		name      string
		args      args
		wantImage inputImage
		wantErr   bool
	}{
		{
			name: "test parseinputImageFromTags",
			args: args{
				repo: "base/infra/datadog/datadog-operator",
				tags: map[string]string{
					"ecr_sync_opt":         "in",
					"ecr_sync_max_results": "10",
					"ecr_sync_constraint":  ">= v1.1.1",
					"ecr_sync_source":      "docker.io/datadog/datadog-operator",
				},
			},
			wantImage: inputImage{
				Constraint:    ">= v1.1.1",
				MaxResults:    10,
				ReleaseOnly:   false,
				EcrRepoPrefix: "base/infra",
				ImageName:     "docker.io/datadog/datadog-operator",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImage := parseInputImageFromTags(tt.args.repo, tt.args.tags)
			if !reflect.DeepEqual(gotImage, tt.wantImage) {
				t.Errorf("parseinputImageFromTags() = %v, want %v", gotImage, tt.wantImage)
			}
		})
	}
}

// test the getECRRepositories function
// func TestGetECRRepositories(t *testing.T) {
// 	// Initialize the ecrClient with a mock implementation of the ECRAPI interface
// 	client := &ecrClient{&mockECRClient{}}

// 	// Call the function being tested
// 	repositories, err := client.getECRRepositories("*")

// 	// Verify the results
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// 	if len(repositories) != 2 {
// 		t.Errorf("Expected 2 repositories, but got %d", len(repositories))
// 	}
// 	if repositories[0].name != "base/infra/datadog/datadog-operator" {
// 		t.Errorf("Expected first repository name to be 'repo1', but got %s", repositories[0].name)
// 	}
// 	if repositories[1].arn != "arn:aws:ecr:us-east-1:123456789012:repository/base/infra/gatekeeper/gatekeeper" {
// 		t.Errorf("Expected second repository ARN to be 'arn:aws:ecr:us-east-1:123456789012:repository/base/infra/gatekeeper/gatekeeper', but got %s", repositories[1].arn)
// 	}
// }

// test the getTagsFromECRRepositories function
// func TestGetTagsFromECRRepositories(t *testing.T) {

// 	// Initialize the ecrClient with a mock implementation of the ECRAPI interface
// 	client := &ecrClient{&mockECRClient{}}

// 	// Call the function being tested
// 	repositories, _ := client.getECRRepositories("*")

// 	// Call the function being tested
// 	tags, err := client.getTagsFromECRRepositories(
// 		&repositories,
// 	)

// 	// Verify the results
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// 	// verify the tags
// 	if len(tags) != 2 {
// 		t.Errorf("Expected 2 tags, but got %d", len(tags))
// 	}
// 	tag1 := false
// 	for _, tag := range tags["base/infra/datadog/datadog-operator"].tags {
// 		if *tag.Key == "ecr_sync_opt" && *tag.Value == "in" {
// 			tag1 = true
// 		}
// 	}
// 	if !tag1 {
// 		t.Errorf("Expected tag1 to be true, but got %t", tag1)
// 	}
// }

// test parseinputImageFromTags function
// func TestParseinputImageFromTags(t *testing.T) {

// 	// Initialize the ecrClient with a mock implementation of the ECRAPI interface
// 	client := &ecrClient{&mockECRClient{}}

// 	// Call the function being tested
// 	repositories, _ := client.getECRRepositories("*")

// 	// Call the function being tested
// 	tags, _ := client.getTagsFromECRRepositories(
// 		&repositories,
// 	)
// 	var err error
// 	images := make(map[string]inputImage)
// 	// Call the function being tested
// 	for repo, tags := range tags {
// 		image, err := parseinputImageFromTags(repo, parseTags(tags.tags))
// 		if err != nil {
// 			log.Printf("Error: %s", err)
// 		}
// 		images[repo] = image
// 	}
// 	// Verify the results
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// 	// verify the inputImage
// 	if len(images) != 2 {
// 		t.Errorf("Expected 2 inputImage, but got %d", len(images))
// 	}
// 	if images["base/infra/datadog/datadog-operator"].ImageName != "docker.io/datadog/operator" {
// 		t.Errorf("Expected repository to be 'docker.io/datadog/operator', but got %s", images["base/infra/datadog/datadog-operator"].ImageName)
// 	}
// 	if images["base/infra/datadog/datadog-operator"].Constraint != "> v1.1.1" {
// 		t.Errorf("Expected constraint to be '> v1.1.1', but got %s", images["base/infra/datadog/datadog-operator"].Constraint)
// 	}
// }

func Test_inputImage_getImagesFromECR(t *testing.T) {
	type args struct {
		ecrImageName  string
		ecrRepoPrefix string
		region        string
		inputImage    *inputImage
	}
	tests := []struct {
		name        string
		args        args
		wantResults map[string]ecrResults
		wantErr     bool
	}{
		{
			name: "TestGetDatadogEcrImages",
			args: args{
				ecrImageName:  "datadog/agent",
				ecrRepoPrefix: "",
				region:        "eu-west-1",
				inputImage: &inputImage{
					EcrRepoPrefix: "base/infra",
					ImageName:     "docker.io/datadog/agent",
				},
			},
			wantErr: false,
			wantResults: map[string]ecrResults{
				"docker.io/datadog/agent:v1.1.1": {
					"datadog/agent", "v1.1.1", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
				"docker.io/datadog/agent:v1.1.2": {
					"datadog/agent", "v1.1.2", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
				"docker.io/datadog/agent:v1.1.3": {
					"datadog/agent", "v1.1.3", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
			},
		},
		{
			name: "TestGetDatadogEcrImagesWithPrefix",
			args: args{
				ecrImageName:  "datadog/agent",
				ecrRepoPrefix: "base/infra",
				region:        "eu-west-1",
				inputImage: &inputImage{
					EcrRepoPrefix: "base/infra",
					ImageName:     "docker.io/datadog/agent",
				},
			},
			wantErr: false,
			wantResults: map[string]ecrResults{
				"docker.io/datadog/agent:v1.1.1": {
					"datadog/agent", "v1.1.1", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
				"docker.io/datadog/agent:v1.1.2": {
					"datadog/agent", "v1.1.2", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
				"docker.io/datadog/agent:v1.1.3": {
					"datadog/agent", "v1.1.3", "sha256:1234567890123456789012345678901234567890123456789012345678901234",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := ecrClient{&mockECRClient{}}

			gotResults, err := svc.getImagesFromECR(tt.args.ecrImageName, tt.args.ecrRepoPrefix, tt.args.region, tt.args.inputImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("inputImage.getImagesFromECR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("inputImage.getImagesFromECR() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}
