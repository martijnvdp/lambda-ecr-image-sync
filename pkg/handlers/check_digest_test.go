package handlers

import (
	"reflect"
	"testing"
)

func Test_checkDigest(t *testing.T) {
	type args struct {
		imageName            string
		resultPublicRepoTags []string
		resultsFromEcr       map[string]ecrResults
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			args: args{
				imageName:            "docker.io/datadog/agent",
				resultPublicRepoTags: []string{"7.28.1-jmx"},
				resultsFromEcr: map[string]ecrResults{
					"docker.io/datadog/agent:7.28.1-jmx": {
						name: "docker.io/datadog/agent",
						tag:  "7.28.1-jmx",
						hash: "sha256:e5d5e7359053957f76535b2c963af07b213570c7c4e1ce477b0cb20ec071b28c",
					},
				},
			},
			name:       "checkDockerhubDigestMultipleManifests",
			wantResult: nil,
			wantErr:    false,
		},
		{
			args: args{
				imageName:            "docker.io/datadog/agent",
				resultPublicRepoTags: []string{"7.32.1-jmx"},
				resultsFromEcr: map[string]ecrResults{
					"docker.io/datadog/agent:7.32.1-jmx": {
						name: "docker.io/datadog/agent",
						tag:  "7.32.1-jmx",
						hash: "sha256:80c2998cee82600b0a38d31abab7e483b09ef2b25ba1600f99002dece5e8d5a9",
					},
				},
			},
			name:       "checkDockerhubDigestMultipleManifests2",
			wantResult: nil,
			wantErr:    false,
		}, /*
			{
				args: args{
					imageName:            "docker.io/martijnvdp/ecr-image-sync",
					resultPublicRepoTags: []string{"v0.0.6"},
					resultsFromEcr: map[string]ecrResults{
						"docker.io/martijnvdp/ecr-image-sync:v0.0.6": {
							name: "martijnvdp/ecr-image-sync",
							tag:  "v0.0.6",
							hash: "sha256:1d7be9f0713e72dcdc886f68aca7ebfd6f8099cac42d34d8e9cfd61d812ef559",
						},
					},
				},
				name:       "checkDockerhubDigestSingleManifest",
				wantResult: nil,
				wantErr:    false,
			},*/{
			args: args{
				imageName:            "quay.io/cilium/cilium",
				resultPublicRepoTags: []string{"v1.4.2"},
				resultsFromEcr: map[string]ecrResults{
					"quay.io/cilium/cilium:v1.4.2": {
						name: "cilium/cilium",
						tag:  "v1.4.2",
						hash: "sha256:b8bd97cf24605f674b462474f71190a405b1ee9b3df767a2a8995c31fc76c8c9",
					},
				},
			},
			name:       "checkHashQuay",
			wantResult: nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := checkDigest(tt.args.imageName, &tt.args.resultPublicRepoTags, &tt.args.resultsFromEcr)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkDigest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("checkDigest() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_checkNoDigest(t *testing.T) {
	type args struct {
		imageName            string
		resultPublicRepoTags *[]string
		resultsFromEcr       *map[string]ecrResults
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			args: args{
				imageName:            "docker.io/martijnvdp/ecr-image-sync",
				resultPublicRepoTags: &[]string{"v0.0.6"},
				resultsFromEcr: &map[string]ecrResults{
					"docker.io/martijnvdp/ecr-image-sync:v0.0.6": {
						name: "martijnvdp/ecr-image-sync",
						tag:  "v0.0.6",
						hash: "sha256:1d7be9f0713e72dcdc886f68aca7ebfd6f8099cac42d34d8e9cfd61d812ef559",
					},
				},
			},
			name:       "checkDigestDisabledTagAlreadyOnEcr",
			wantResult: nil,
			wantErr:    false,
		},
		{
			args: args{
				imageName:            "docker.io/martijnvdp/ecr-image-sync",
				resultPublicRepoTags: &[]string{"v0.0.6"},
				resultsFromEcr: &map[string]ecrResults{
					"docker.io/martijnvdp/ecr-image-sync:v0.0.7": {
						name: "martijnvdp/ecr-image-sync",
						tag:  "v0.0.7",
						hash: "sha256:1d7be9f0713e72dcdc886f68aca7ebfd6f8099cac42d34d8e9cfd61d812ef559",
					},
				},
			},
			name:       "checkDigestDisabledTagNotOnEcr",
			wantResult: []string{"v0.0.6"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := checkNoDigest(tt.args.imageName, tt.args.resultPublicRepoTags, tt.args.resultsFromEcr)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkNoDigest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("checkNoDigest() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
