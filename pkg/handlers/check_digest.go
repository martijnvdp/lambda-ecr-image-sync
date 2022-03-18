package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

type manifest struct {
	Digest string `json:"digest"`
}

type manifestsBlob struct {
	Manifests []manifest `json:"manifests"`
}

func checkManifestDigests(ecrDigest string, manifestDigests manifestsBlob) bool {
	for _, manifest := range manifestDigests.Manifests {
		if manifest.Digest == ecrDigest {
			return true
		}
	}
	return false
}

func parseImageSource(ctx context.Context, name string) (types.ImageSource, error) {
	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return nil, err
	}
	sys := newSystemContext()

	if err != nil {
		return nil, err
	}
	return ref.NewImageSource(ctx, sys)
}

func checkNoDigest(imageName string, resultPublicRepoTags *[]string, resultsFromEcr *map[string]ecrResults) (result []string, err error) {
	for _, tag := range *resultPublicRepoTags {
		if (*resultsFromEcr)[imageName+":"+tag].hash == "" {
			result = append(result, tag)
		}
	}

	return result, err
}

// function to compare the digest of the ecr results with the public repo digests, and also check the manifest digests if its a multiplatform manifest.
func checkDigest(imageName string, resultPublicRepoTags *[]string, resultsFromEcr *map[string]ecrResults) (result []string, err error) {
	var manifestDigests manifestsBlob
	ctx := context.TODO()
	systemContext := newSystemContext()

	for _, tag := range *resultPublicRepoTags {

		if (*resultsFromEcr)[imageName+":"+tag].hash == "" {
			result = append(result, tag)
		} else {
			src, err := parseImageSource(ctx, "docker://"+imageName+":"+tag)

			if err != nil {
				return result, err
			}
			img, err := image.FromUnparsedImage(ctx, systemContext, image.UnparsedInstance(src, nil))

			if err != nil {
				return result, err
			}
			digest, err := docker.GetDigest(ctx, systemContext, img.Reference())

			if err != nil {
				return result, err
			}
			match := (*resultsFromEcr)[imageName+":"+tag].hash == fmt.Sprint(digest)

			if !match {
				manifestBlob, _, err := img.Manifest(ctx)

				if err != nil {
					return result, err
				}
				json.Unmarshal([]byte(manifestBlob), &manifestDigests)

				if !checkManifestDigests((*resultsFromEcr)[imageName+":"+tag].hash, manifestDigests) {
					result = append(result, tag)
				}
			}
		}
	}

	return result, err
}
