package lambda

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type mockECRClient struct {
	ecriface.ECRAPI
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

func Test_parseinputRepositoryFromTags(t *testing.T) {
	type args struct {
		repo string
		tags map[string]string
	}
	tests := []struct {
		name      string
		args      args
		wantImage inputRepository
		wantErr   bool
	}{
		{
			name: "test parseinputRepositoryFromTags",
			args: args{
				repo: "dev/test/datadog/datadog-operator",
				tags: map[string]string{
					"ecr_sync_opt":         "in",
					"ecr_sync_max_results": "10",
					"ecr_sync_constraint":  ">= v1.1.1",
					"ecr_sync_source":      "docker.io/datadog/datadog-operator",
				},
			},
			wantImage: inputRepository{
				constraint:   ">= v1.1.1",
				ecrImageName: "dev/test/datadog/datadog-operator",
				maxResults:   10,
				releaseOnly:  false,
				source:       "docker.io/datadog/datadog-operator",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImage := parseinputRepositoryFromTags(tt.args.repo, tt.args.tags)
			if !reflect.DeepEqual(gotImage, tt.wantImage) {
				t.Errorf("parseinputRepositoryFromTags() = %v, want %v", gotImage, tt.wantImage)
			}
		})
	}
}

func Test_inputRepository_getImagesFromECR(t *testing.T) {
	type args struct {
		ecrsource       string
		ecrRepoPrefix   string
		region          string
		inputRepository *inputRepository
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
				ecrsource:     "datadog/agent",
				ecrRepoPrefix: "",
				region:        "eu-west-1",
				inputRepository: &inputRepository{
					source: "docker.io/datadog/agent",
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
				ecrsource:     "datadog/agent",
				ecrRepoPrefix: "base/infra",
				region:        "eu-west-1",
				inputRepository: &inputRepository{
					source: "docker.io/datadog/agent",
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

			gotResults, err := svc.getImagesFromECR(tt.args.ecrsource, tt.args.region, tt.args.inputRepository)
			if (err != nil) != tt.wantErr {
				t.Errorf("inputRepository.getImagesFromECR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("inputRepository.getImagesFromECR() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}
