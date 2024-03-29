package lambda

import (
	"reflect"
	"testing"
)

func Test_inputRepository_maxResults(t *testing.T) {
	type args struct {
		globalMaxResults int
	}
	tests := []struct {
		name           string
		i              *inputRepository
		args           args
		wantMaxResults int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMaxResults := tt.i.getMaxResults(tt.args.globalMaxResults); gotMaxResults != tt.wantMaxResults {
				t.Errorf("inputRepository.maxResults() = %v, want %v", gotMaxResults, tt.wantMaxResults)
			}
		})
	}
}

func Test_inputRepository_checkTagsFromPublicRepo(t *testing.T) {
	type args struct {
		inputTags  []string
		maxResults int
	}
	tests := []struct {
		name       string
		i          *inputRepository
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "TestExcludeFormedVersions",
			args: args{
				inputTags: []string{"7.32.1-rc.3-jmx", "7.32.1-rc.3-server_core", "7.32.1-servercore-jmx", "7.32.1-servercore", "latest", "latest-jmx", "latest-py2", "latest-py2-jmx", "latest-servercore", "latest-servercore-jmx", "7.3.x-exemplars"},
			},
			i: &inputRepository{
				constraint: ">= 7.30.0",
				includeRLS: []string{"jmx"},
				excludeRLS: []string{"rc"},
			},
			wantResult: []string{"7.32.1-servercore-jmx"},
			wantErr:    false,
		},
		{
			name: "TestMallFormedVersions",
			args: args{
				inputTags: []string{"latest-py2", "latest-py2-jmx", "7.3.x-exemplars"},
			},
			i:          &inputRepository{},
			wantResult: []string{"latest-py2", "latest-py2-jmx"},
			wantErr:    false,
		},
		{
			name: "TestMallFormedVersionsGrafana",
			args: args{
				inputTags: []string{"9.3.6", "9.3.4", "7.3.x-exemplars", "7.4.x-exemplars", "7.5.x-exemplars"},
			},
			i:          &inputRepository{},
			wantResult: []string{"9.3.6", "9.3.4"},
			wantErr:    false,
		},
		{
			name: "TestMatchExclTagsNonVersionsTagLatestCurrent",
			args: args{
				inputTags: []string{"v0.0.3", "current", "v0.0.6", "latest", "v0.0.7"},
			},
			i: &inputRepository{
				excludeTags: []string{"v0.0.3", "v0.0.6"},
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"current", "latest", "v0.0.7"},
			wantErr:    false,
		},
		{
			name: "TestMatchInclTagsNonVersionsTagLatestCurrentAndConstraint",
			args: args{
				inputTags: []string{"v0.0.3", "current", "v0.0.6", "latest", "v0.0.7"},
			},
			i: &inputRepository{
				includeTags: []string{"latest", "current"},
				constraint:  ">= v0.0.6",
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"current", "latest", "v0.0.7", "v0.0.6"},
			wantErr:    false,
		},
		{
			name: "TestMatchIncltagsNonVersionsTagLatestCurrent",
			args: args{
				inputTags: []string{"v0.0.3", "current", "v0.0.6", "latest", "v0.0.7"},
			},
			i: &inputRepository{
				includeTags: []string{"latest", "current"},
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"current", "latest"},
			wantErr:    false,
		},
		{
			name: "TestMatchInclTagsOnly",
			args: args{
				inputTags: []string{"v0.0.3", "current", "v0.0.6", "latest", "v0.0.7"},
			},
			i: &inputRepository{
				includeTags: []string{"v0.0.6", "v0.0.7"},
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"v0.0.7", "v0.0.6"},
			wantErr:    false,
		},
		{
			name: "TestMatchInclRLSWithConstraint",
			args: args{
				inputTags: []string{"latest", "v0.0.3", "v0.0.4-debian-r20", "v0.0.5-beta", "v0.0.6-win", "v0.0.7", "v0.0.8-debian", "v0.0.9", "v0.1.0", "v0.1.1"},
			},
			i: &inputRepository{
				includeRLS: []string{"debian", "win"},
				constraint: "< v0.0.8",
				source:     "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"v0.0.7", "v0.0.6-win", "v0.0.4-debian-r20", "v0.0.3"},
			wantErr:    false,
		},
		{
			name: "TestNoPrereleases",
			args: args{
				inputTags: []string{"latest", "v0.0.3", "v0.0.4", "v0.0.5-beta", "v0.0.6", "v0.0.7", "v0.0.8-rc", "v0.0.9", "v0.1.0", "v0.1.1"},
			},
			i: &inputRepository{
				includeTags: []string{"v0.0.4", "v0.1.1"},
				constraint:  "< v0.0.9",
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"v0.1.1", "v0.0.7", "v0.0.6", "v0.0.4", "v0.0.3"},
			wantErr:    false,
		},
		{
			name: "TestCheckTagLiveDataIncludeReleaseMaxresults",
			args: args{
				inputTags: []string{"latest", "v0.0.3", "v0.0.4", "v0.0.5-beta", "v0.0.6", "v0.0.7", "v0.0.8", "v0.0.9", "v0.1.0", "v0.1.1", "v0.1.6-rc"},
			},
			i: &inputRepository{
				source:     "ghcr.io/martijnvdp/ecr-image-sync",
				includeRLS: []string{"rc"},
				maxResults: 6,
			},
			wantResult: []string{"v0.1.6-rc"},
			wantErr:    false,
		},
		{
			name: "TestIncludeTagTogetherWithConstraint",
			args: args{
				inputTags: []string{"latest", "v0.0.3", "v0.0.4", "v0.0.5", "v0.0.6", "v0.0.7", "v0.0.8", "v0.0.9", "v0.1.0", "v0.1.1"},
			},
			i: &inputRepository{
				includeTags: []string{"v0.0.4", "v0.1.1"},
				constraint:  "< v0.0.7",
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"v0.1.1", "v0.0.6", "v0.0.5", "v0.0.4", "v0.0.3"},
			wantErr:    false,
		},
		{
			name: "TestMinMaxConstraint",
			args: args{
				inputTags:  []string{"latest", "v0.0.3", "v0.0.4", "v0.0.5", "v0.0.6", "v0.0.7", "v0.0.8", "v0.0.9", "v0.1.0", "v0.1.1"},
				maxResults: 2,
			},
			i: &inputRepository{
				constraint: ">= v0.0.8, <= v0.0.9",
				source:     "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"v0.0.9", "v0.0.8"},
			wantErr:    false,
		},
		{
			name: "TestVersionSortingMaxResult",
			args: args{
				inputTags:  []string{"latest", "v0.0.3", "v0.0.4", "v0.0.7", "v0.0.8", "v0.0.9", "v0.1.0", "v0.1.1", "v0.0.5", "v0.0.6"},
				maxResults: 7,
			},
			i: &inputRepository{
				source: "ghcr.io/martijnvdp/ecr-image-sync",
			},
			wantResult: []string{"latest", "v0.1.1", "v0.1.0", "v0.0.9", "v0.0.8", "v0.0.7", "v0.0.6"},
			wantErr:    false,
		},
		{
			name: "TestPreReleaseDataFilter",
			args: args{
				inputTags: []string{"latest", "0.3.0-debian-10-r32", "0.5.0-debian-10-r59", "0.9.0-debian-10-r32", "1.5.0-debian-12"},
			},
			i: &inputRepository{
				includeRLS: []string{"debian", "test"},
				constraint: ">= v0.6.0",
				source:     "docker.io/bitnami/metrics-server",
				maxResults: -1,
			},
			wantResult: []string{"1.5.0-debian-12", "0.9.0-debian-10-r32"},
			wantErr:    false,
		},
		{
			name: "TestPreReleaseOnlyDataFilter",
			args: args{
				inputTags: []string{"latest", "v0.7.0", "0.3.0-debian-10-r32", "0.5.0-debian-10-r59", "0.9.0-debian-10-r32", "1.5.0-debian-12"},
			},
			i: &inputRepository{
				includeRLS:  []string{"debian", "test"},
				constraint:  ">= v0.6.0",
				source:      "docker.io/bitnami/metrics-server",
				maxResults:  -1,
				releaseOnly: true,
			},
			wantResult: []string{"1.5.0-debian-12", "0.9.0-debian-10-r32"},
			wantErr:    false,
		}, {
			name: "TestPreReleaseDataFilterWithExclude",
			args: args{
				inputTags: []string{"latest", "0.3.0-debian-10-r32", "0.5.0-debian-10-r59", "0.9.0-debian-10-r32", "0.9.0-debian-10-beta", "1.5.0-debian-12", "1.5.0-debian-12-rc"},
			},
			i: &inputRepository{
				includeRLS: []string{"debian", "test"},
				excludeRLS: []string{"beta", "rc"},
				constraint: ">= v0.6.0",
				source:     "docker.io/bitnami/metrics-server",
				maxResults: -1,
			},
			wantResult: []string{"1.5.0-debian-12", "0.9.0-debian-10-r32"},
			wantErr:    false,
		},
		{
			name: "TestexcludeTags",
			args: args{
				inputTags: []string{"latest", "v1.4.1", "v1.4.5", "v1.4.6", "v1.4.7", "v1.4.8", "v1.4.9"},
			},
			i: &inputRepository{
				excludeTags: []string{"v1.4.8", "v1.4.5"},
				source:      "quay.io/cilium/cilium",
				maxResults:  6,
			},
			wantResult: []string{"latest", "v1.4.9", "v1.4.7", "v1.4.6", "v1.4.1"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.i.checkTagsFromPublicRepo(&tt.args.inputTags, tt.args.maxResults)
			if (err != nil) != tt.wantErr {
				t.Errorf("inputRepository.checkTagsFromPublicRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("inputRepository.checkTagsFromPublicRepo() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_inputRepository_checkTagsFromResultsPublicRepo(t *testing.T) {
	type args struct {
		maxResults int
	}
	tests := []struct {
		name       string
		i          *inputRepository
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "TestCheckTagLiveDataIncludetags2",
			i: &inputRepository{
				source:      "ghcr.io/martijnvdp/ecr-image-sync",
				includeTags: []string{"v0.1.1", "v0.0.4"},
			},
			wantResult: []string{"v0.1.1", "v0.0.4"},
			wantErr:    false,
		},
		{
			name: "TestCheckTagLiveDataIncludetags",
			i: &inputRepository{
				source:      "docker.io/openpolicyagent/gatekeeper",
				includeTags: []string{"v3.4.0", "v3.5.1"},
			},
			wantResult: []string{"v3.5.1", "v3.4.0"},
			wantErr:    false,
		},
		{
			name: "TestCheckTagLiveData",
			i: &inputRepository{
				source:     "docker.io/openpolicyagent/gatekeeper",
				constraint: ">= 3.5.0, < 3.7.0",
			},
			wantResult: []string{"v3.6.0", "v3.5.2", "v3.5.1", "v3.5.0"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputTags, err := tt.i.getTagsFromPublicRepo()
			if (err == nil) != tt.wantErr {
				gotResult, err := tt.i.checkTagsFromPublicRepo(&inputTags, tt.args.maxResults)
				if (err != nil) != tt.wantErr {
					t.Errorf("inputRepository.checkTagsFromPublicRepo() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(gotResult, tt.wantResult) {
					t.Errorf("inputRepository.checkTagsFromPublicRepo() = %v, want %v", gotResult, tt.wantResult)
				}
			}
		})
	}
}
