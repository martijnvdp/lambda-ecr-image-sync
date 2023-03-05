package lambda

import (
	"reflect"
	"testing"
)

func Test_createZipFile(t *testing.T) {
	type args struct {
		file   string
		target string
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
			if err := createZipFile(tt.args.file, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("createZipFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_buildCSVFile(t *testing.T) {
	type args struct {
		source  string
		options syncOptions
		env     environmentVars
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
				source: "gcr.io/datadoghq/agent",
				options: syncOptions{
					ecrImageName: "dev/datadoghq/agent",
					tags:         []string{"v7.32.0", "v7.31.0", "v7.28.0"},
				},
				env: environmentVars{
					awsAccount: "123321",
					awsRegion:  "eu-west-2",
				},
			},
			wantErr:        false,
			wantCsvContent: []csvFormat{{"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/dev/datadoghq/agent", "v7.32.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/dev/datadoghq/agent", "v7.31.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/dev/datadoghq/agent", "v7.28.0"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCsvContent, err := buildCSVFile(tt.args.source, tt.args.options, tt.args.env)
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
