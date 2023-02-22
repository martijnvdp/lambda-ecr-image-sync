package lambda

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

func login(opts loginOptions) error {
	if opts.passwordStdin {
		contents, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		opts.password = strings.TrimSuffix(string(contents), "\n")
		opts.password = strings.TrimSuffix(opts.password, "\r")
	}
	if opts.user == "" && opts.password == "" {
		return errors.New("username and password required")
	}
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}
	creds := cf.GetCredentialsStore(opts.serverAddress)
	if opts.serverAddress == name.DefaultRegistry {
		opts.serverAddress = authn.DefaultAuthKey
	}
	if err := creds.Store(types.AuthConfig{
		ServerAddress: opts.serverAddress,
		Username:      opts.user,
		Password:      opts.password,
	}); err != nil {
		return err
	}

	if err := cf.Save(); err != nil {
		return err
	}
	log.Printf("logged in via %s", cf.Filename)
	return nil
}

func (svc *ecrClient) copyImageWithCrane(imageName, tag, awsPrefix, ecrRepoPrefix, ecrImageName string) (err error) {
	params := crane.Options{
		Platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	}

	opts := []crane.Option{crane.WithPlatform(params.Platform)}

	if err := crane.Copy((imageName + ":" + tag), (awsPrefix + "/" + ecrRepoPrefix + "/" + ecrImageName + ":" + tag), opts...); err != nil {
		log.Printf("error copying image: %v", err)
		return err
	}

	return nil
}

func syncImages(imageName, ecrImageName, ecrRepoPrefix string, resultsFromPublicRepo *[]string, awsAccount, awsRegion string) (count int, err error) {
	svc, err := newEcrClient(awsRegion)

	if err != nil {
		log.Println("error logging in to ecr: ", err)
		return 0, err
	}

	awsPrefix := awsAccount + ".dkr.ecr." + awsRegion + ".amazonaws.com"
	log.Printf("add login for %v", awsPrefix)
	awsAuthData, err := svc.getECRAuthData()

	if err != nil {
		log.Println("error getting authdata: ", err)
		return 0, err
	}

	err = login(loginOptions{
		serverAddress: awsPrefix,
		user:          awsAuthData.username,
		password:      awsAuthData.password,
	})

	if err != nil {
		log.Println("error authentication to ecr: ", err)
		return 0, err
	}

	if os.Getenv("DOCKER_USERNAME") != "" && os.Getenv("DOCKER_PASSWORD") != "" {
		err = login(loginOptions{
			serverAddress: "docker.io",
			user:          os.Getenv("DOCKER_USERNAME"),
			password:      os.Getenv("DOCKER_PASSWORD"),
		})
		if err != nil {
			log.Println("error logging in to docker.io: ", err)
			return 0, err
		}
	}

	for _, tag := range *resultsFromPublicRepo {
		log.Printf("copying %s:%s to %s/%s/%s:%s", imageName, tag, awsPrefix, ecrRepoPrefix, ecrImageName, tag)
		err := svc.copyImageWithCrane(imageName, tag, awsPrefix, ecrRepoPrefix, ecrImageName)

		if err != nil {
			log.Println("error copying image: ", err)
			return 0, err
		}
		count++
	}
	return count, nil
}
