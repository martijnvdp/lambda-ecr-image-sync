package handlers

import (
	"reflect"
	"testing"
)

func Test_buildCSVFile(t *testing.T) {
	type args struct {
		imageName             string
		ecrImageName          string
		ecrRepoPrefix         string
		resultsFromPublicRepo *[]string
		awsAccount            string
		awsRegion             string
	}
	tests := []struct {
		name           string
		args           args
		wantCsvContent []csvFormat
		wantErr        bool
	}{
		{
			name: "TestbuildCSV",
			args: args{
				imageName:             "gcr.io/datadoghq/agent",
				ecrImageName:          "datadoghq/agent",
				ecrRepoPrefix:         "base/infra",
				resultsFromPublicRepo: &[]string{"v7.32.0", "v7.31.0", "v7.28.0"},
				awsAccount:            "123321",
				awsRegion:             "eu-west-2",
			},
			wantErr:        false,
			wantCsvContent: []csvFormat{{"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.32.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.31.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.28.0"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCsvContent, err := buildCSVFile(tt.args.imageName, tt.args.ecrImageName, tt.args.ecrRepoPrefix, tt.args.resultsFromPublicRepo, tt.args.awsAccount, tt.args.awsRegion)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCSVFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCsvContent, tt.wantCsvContent) {
				t.Errorf("buildCSVFile() = %v, want %v", gotCsvContent, tt.wantCsvContent)
			}
		})
	}
}

func Test_writeCSVFile(t *testing.T) {
	type args struct {
		csvContent  *[]csvFormat
		csvFileName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeCSVFile(tt.args.csvContent, tt.args.csvFileName); (err != nil) != tt.wantErr {
				t.Errorf("writeCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
