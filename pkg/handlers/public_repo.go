package handlers

import (
	"context"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

func newSystemContext() *types.SystemContext {
	ctx := &types.SystemContext{
		ArchitectureChoice:      "amd64",
		OSChoice:                "linux",
		DockerRegistryUserAgent: "ecr-image-sync/v0.1.1",
	}
	return ctx
}

func (i *inputImage) getTagsFromPublicRepo() (tags []string, err error) {
	ctx := context.TODO()
	systemContext := newSystemContext()
	imageref, err := alltransports.ParseImageName("docker://" + i.ImageName)

	if err != nil {
		return tags, err
	}
	tags, err = docker.GetRepositoryTags(ctx, systemContext, imageref)

	if err != nil {
		return tags, err
	}

	return tags, err
}
