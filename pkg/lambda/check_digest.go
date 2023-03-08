package lambda

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func getDigest(source string) (string, error) {
	ref, err := name.ParseReference(source)
	if err != nil {
		panic(err)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil && strings.Contains(err.Error(), "unsupported MediaType: \"application/vnd.docker.distribution.manifest.v1") {
		return "", nil
	}
	if err != nil {
		panic(err)
	}

	digest, err := img.Digest()
	// if err not is null and the error message does not contain "unsupported MediaType: "application/vnd.docker.distribution.manifest.v1" retun error else continue
	if err != nil && strings.Contains(err.Error(), "unsupported MediaType: \"application/vnd.docker.distribution.manifest.v1") {
		return "", nil
	}
	if err != nil {
		panic(err)
	}
	return digest.String(), err
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
			match := (*resultsFromEcr)[imageName+":"+tag].hash == digest

			if !match {
				fmt.Printf("no match %s", imageName)
				result = append(result, tag)
			}
		}
	}

	return result, err
}
