package handlers

import (
	"os"
	"reflect"
	"testing"
)

func Test_inputImage_getImagesFromECR(t *testing.T) {
	type args struct {
		ecrImageName  string
		ecrRepoPrefix string
		region        string
	}
	tests := []struct {
		name        string
		i           *inputImage
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
			},
			i: &inputImage{
				EcrRepoPrefix: "base/infra",
				ImageName:     "docker.io/datadog/agent",
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
			break
		}
		t.Run(tt.name, func(t *testing.T) {
			gotResults, err := tt.i.getImagesFromECR(tt.args.ecrImageName, tt.args.ecrRepoPrefix, tt.args.region)
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
