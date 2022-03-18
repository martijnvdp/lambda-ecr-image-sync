package handlers

import (
	"reflect"
	"testing"

	"github.com/containers/image/v5/types"
)

func Test_newSystemContext(t *testing.T) {
	tests := []struct {
		name string
		want *types.SystemContext
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSystemContext(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSystemContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inputImage_getTagsFromPublicRepo(t *testing.T) {
	tests := []struct {
		name     string
		i        *inputImage
		wantTags []string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTags, err := tt.i.getTagsFromPublicRepo()
			if (err != nil) != tt.wantErr {
				t.Errorf("inputImage.getTagsFromPublicRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("inputImage.getTagsFromPublicRepo() = %v, want %v", gotTags, tt.wantTags)
			}
		})
	}
}
