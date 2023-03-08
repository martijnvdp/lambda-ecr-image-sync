package lambda

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func (i *inputRepository) getTagsFromPublicRepo() (tags []string, err error) {
	params := crane.Options{
		Platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	}

	opts := []crane.Option{crane.WithPlatform(params.Platform)}

	tags, err = crane.ListTags(i.source, opts...)
	if err != nil {
		return tags, fmt.Errorf("reading tags for %s: %w", i.source, err)
	}

	return tags, err
}
