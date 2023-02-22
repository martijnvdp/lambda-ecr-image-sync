package lambda

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type ecrClient struct {
	ecriface.ECRAPI
}

type ecrResults struct {
	name string
	tag  string
	hash string
}

type repository struct {
	name string
	arn  string
}

type authData struct {
	username string
	password string
}

type repoTags struct {
	tags []*ecr.Tag
	repo repository
}

func checkRepoTag(key, value string, tags []*ecr.Tag) bool {

	for _, tag := range tags {
		if *tag.Key == key && *tag.Value == value {
			return true
		}
	}

	return false
}

// starts a new session and returns an ecrClient
func newEcrClient(region string) (*ecrClient, error) {
	mySession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return nil, err
	}

	svc := ecr.New(mySession)

	return &ecrClient{svc}, nil
}

func parseInputImageFromTags(repo string, tags map[string]string) (InputImage, error) {
	var imageName string
	image := InputImage{}

	if tags["ecr_sync_source"] == "" {
		log.Printf("ecr source not set of %v", repo)
		return image, nil
	}

	parts := strings.Split(tags["ecr_sync_source"], "/")
	if len(parts) < 2 {
		fmt.Println("Invalid repository format")
		return InputImage{}, nil
	}

	for _, part := range parts {
		if strings.Contains(part, ".") {
			continue
		}
		imageName = imageName + "/" + part
	}

	imageName = strings.TrimLeft(imageName, "/")

	if tags["ecr_sync_release_only"] == "true" {
		image.ReleaseOnly = true
	}

	if tags["ecr_sync_max_results"] != "" {
		image.MaxResults, _ = strconv.Atoi(tags["ecr_sync_max_results"])
	}
	image.Constraint = tags["ecr_sync_constraint"]

	image.ExcludeRLS = stringToSlice(tags["ecr_sync_exclude_rls"])
	image.ExcludeTags = stringToSlice(tags["ecr_sync_exclude_tags"])
	image.ImageName = tags["ecr_sync_source"]
	image.EcrRepoPrefix = strings.TrimSuffix(repo, "/"+imageName)
	image.IncludeRLS = stringToSlice(tags["ecr_sync_include_rls"])
	image.IncludeTags = stringToSlice(tags["ecr_sync_include_tags"])

	return image, nil
}

func parseTags(tags []*ecr.Tag) map[string]string {

	operators := map[string]string{
		"-gt": ">",
		"-ge": ">=",
		"-lt": "<",
		"-le": "=<",
	}
	tagMap := make(map[string]string)

	// Loop through tags
	for _, tag := range tags {

		// Skip if not ecr_sync tag
		if !strings.HasPrefix(*tag.Key, "ecr_sync") {
			continue
		}
		newVal := *tag.Value

		// Check for ecr_sync_constraint tag and replace with correct operator for constraint
		if *tag.Key == "ecr_sync_constraint" {
			prefix := strings.SplitN(*tag.Value, " ", 3)[0]
			op := operators[prefix]
			newVal = op + strings.TrimPrefix(*tag.Value, prefix)
		}

		tagMap[*tag.Key] = newVal
	}

	return tagMap
}

func (svc *ecrClient) getECRAuthData() (authData, error) {
	base64token, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return authData{}, fmt.Errorf("failed to retrieve ecr token: %w", err)
	}

	// #2
	if len(base64token.AuthorizationData) == 0 {
		return authData{}, fmt.Errorf("ecr token is empty")
	}
	if len(base64token.AuthorizationData) > 1 {
		return authData{}, fmt.Errorf("multiple ecr tokens: length: %d", len(base64token.AuthorizationData))
	}
	if base64token.AuthorizationData[0].AuthorizationToken == nil {
		return authData{}, fmt.Errorf("ecr token is nil")
	}

	token, err := base64.URLEncoding.DecodeString(*base64token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return authData{}, fmt.Errorf("failed to decode ecr token: %w", err)
	}

	return authData{
		username: strings.Split(string(token), ":")[0],
		password: strings.Split(string(token), ":")[1],
	}, err
}

func (svc *ecrClient) getECRRepositories() (repositories []repository, err error) {
	input := &ecr.DescribeRepositoriesInput{}
	repos, err := svc.DescribeRepositories(input)

	if err != nil {
		log.Printf("Error: %s", err)
	}

	for _, repo := range repos.Repositories {
		repositories = append(repositories, repository{
			name: *repo.RepositoryName,
			arn:  *repo.RepositoryArn,
		})
	}

	return repositories, err
}

func (svc *ecrClient) getImagesFromECR(ecrImageName, ecrRepoPrefix, region string, i *InputImage) (results map[string]ecrResults, err error) {
	var ecrResult *ecr.ListImagesOutput
	results = make(map[string]ecrResults)
	ecrRepoPrefix = tryString(ecrRepoPrefix, i.EcrRepoPrefix)
	repositoryName := (ecrRepoPrefix + "/" + ecrImageName)

	input := &ecr.ListImagesInput{
		RepositoryName: aws.String(repositoryName),
		Filter: &ecr.ListImagesFilter{
			TagStatus: aws.String("TAGGED"),
		},
	}
	ecrResult, err = svc.ListImages(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case ecr.ErrCodeServerException:
				fmt.Println(ecr.ErrCodeServerException, awsErr.Error())
			case ecr.ErrCodeInvalidParameterException:
				fmt.Println(ecr.ErrCodeInvalidParameterException, awsErr.Error())
			case ecr.ErrCodeRepositoryNotFoundException:
				fmt.Println(ecr.ErrCodeRepositoryNotFoundException, awsErr.Error())
			default:
				fmt.Println(awsErr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	for _, id := range ecrResult.ImageIds {
		results[i.ImageName+":"+*id.ImageTag] = ecrResults{
			name: ecrImageName,
			tag:  *id.ImageTag,
			hash: *id.ImageDigest,
		}
	}

	return results, err
}

func (svc *ecrClient) getTagsFromECRRepositories(repositories *[]repository) (tags map[string]repoTags, err error) {
	// Create map to hold tags
	tags = make(map[string]repoTags)

	for _, repo := range *repositories {
		// Get tags for repo
		repository_tags, err := svc.ListTagsForResource(&ecr.ListTagsForResourceInput{ResourceArn: aws.String(*&repo.arn)})

		if err != nil {
			log.Printf("Error: %s", err)
		}

		// Skip if no tags
		if len(repository_tags.Tags) == 0 {
			continue
		}

		// Skip if no ecr_sync_opt tag or not set to "in"
		if checkRepoTag("ecr_sync_opt", "in", repository_tags.Tags) {
			tags[repo.name] = repoTags{
				tags: repository_tags.Tags,
				repo: repo,
			}
		}
	}
	return tags, err
}

func (svc *ecrClient) getInputImagesFromTags() (images []InputImage, err error) {
	repositories, err := svc.getECRRepositories()
	if err != nil {
		return nil, err
	}

	tags, err := svc.getTagsFromECRRepositories(&repositories)
	if err != nil {
		return nil, err
	}

	for repo, tag := range tags {
		image, err := parseInputImageFromTags(repo, parseTags(tag.tags))
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, err
}