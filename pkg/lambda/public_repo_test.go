package lambda

import (
	"reflect"
	"testing"
)

func Test_inputRepository_getTagsFromPublicRepo(t *testing.T) {
	tests := []struct {
		name     string
		i        *inputRepository
		wantTags []string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTags, err := tt.i.getTagsFromPublicRepo()
			if (err != nil) != tt.wantErr {
				t.Errorf("inputRepository.getTagsFromPublicRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("inputRepository.getTagsFromPublicRepo() = %v, want %v", gotTags, tt.wantTags)
			}
		})
	}
}
