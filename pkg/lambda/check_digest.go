package lambda

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func getDigest(source string) (string, error) {

	params := crane.Options{
		Platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	}
	opts := []crane.Option{crane.WithPlatform(params.Platform)}

	manifest, err := crane.Manifest(source, opts...)
	if err != nil && strings.Contains(err.Error(), "unsupported MediaType: \"application/vnd.docker.distribution.manifest.v1") {
		return "", nil
	}
	if err != nil && strings.Contains(err.Error(), "You have reached your pull rate limit.") {
		log.Printf("Pull rate limit exceeded for %s", source)
		return "", nil
	}

	hash := sha256.New()
	hash.Write(manifest)
	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	return digest, err
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

	for _, tag := range *resultPublicRepoTags {

		if (*resultsFromEcr)[imageName+":"+tag].hash == "" {
			result = append(result, tag)
		} else {
			digest, err := getDigest(imageName + ":" + tag)
			if err != nil {
				return result, err
			}
			match := (*resultsFromEcr)[imageName+":"+tag].hash == digest || digest == ""

			if !match {
				result = append(result, tag)
			}
		}
	}

	return result, err
}
