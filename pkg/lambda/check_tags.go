package lambda

import (
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/martijnvdp/lambda-ecr-image-sync/external/go-version"
)

const noConstraint string = "> 0, < 0"

func checkRelease(v *version.Version, c *version.Constraints) bool {
	return v.Prerelease() == "" && c.Check(v)
}

func checkPreRelease(v *version.Version, c *version.Constraints) bool {
	if c.String() != noConstraint {
		return v.Prerelease() != "" && c.Check(v.Core())
	}
	return true
}

func compareIncExclTags(tag *string, tags *[]string) bool {
	for _, t := range *tags {
		if t == *tag {
			return true
		}
	}
	return false
}

func comparePreReleases(v *version.Version, releases *[]string) bool {
	split := strings.Split(v.Prerelease(), "-")
	for _, r := range *releases {
		for i := 0; !(i >= len(split)); i++ {
			switch {
			case strings.HasPrefix(split[i], r):
				return true
			}
		}
	}

	return false
}

func (i *InputImage) checkExcConstraints(v *version.Version, c *version.Constraints) bool {
	return comparePreReleases(v, &i.ExcludeRLS)
}

func (i *InputImage) checkExcTags(t string) bool {
	return compareIncExclTags(&t, &i.ExcludeTags)
}

func (i *InputImage) checkFilter() bool {
	return len(i.IncludeTags) == 0 && len(i.ExcludeTags) == 0 && !(len(i.IncludeRLS) > 0) && !(len(i.ExcludeRLS) > 0) && i.Constraint == ""
}

func (i *InputImage) checkIncConstraints(v *version.Version, c *version.Constraints) bool {
	return comparePreReleases(v, &i.IncludeRLS)
}

func (i *InputImage) checkIncTags(t string) bool {
	return compareIncExclTags(&t, &i.IncludeTags)
}

func (i *InputImage) checkNonVersionTags(tag string) bool {
	switch {
	case len(i.IncludeTags) > 0 && i.checkIncTags(tag):
		return true
	case len(i.ExcludeTags) > 0 && !i.checkExcTags(tag):
		return true
	}
	return false
}

func (i *InputImage) checkVersionTags(v *version.Version, c *version.Constraints) bool {
	switch {
	case len(i.IncludeTags) > 0 && i.checkIncTags(v.Original()):
		return true
	case len(i.ExcludeTags) > 0:
		return !i.checkExcTags(v.Original())
	case checkRelease(v, c) && !(i.ReleaseOnly):
		return true
	case i.checkExcConstraints(v, c):
		return false
	case i.checkIncConstraints(v, c):
		return checkPreRelease(v, c)
	}

	return false
}

func (i *InputImage) createConstraint() (constraints version.Constraints, err error) {
	if i.Constraint != "" {
		return version.NewConstraint(i.Constraint)
	}
	return version.NewConstraint(noConstraint)
}

func (i *InputImage) maxResults(globalMaxResults int) (maxResults int) {
	maxResults = maxInt(globalMaxResults, i.MaxResults)

	if !(maxResults > 0) {
		maxResults = -1
	}

	if len(i.IncludeTags) > 0 && i.Constraint == "" {
		return len(i.IncludeTags)
	}
	return maxResults
}

func parseVersions(tags *[]string) (versionTags []string, nonVersionTags []string) {
	matchVersionPrefix, _ := regexp.Compile(`(^v\d{1,5})|(^\d{1,5}\.)`)
	matchRegExp, _ := regexp.Compile(version.VersionRegexpRaw)

	for _, t := range *tags {
		if matchRegExp.MatchString(t) && matchVersionPrefix.MatchString(t) {
			versionTags = append(versionTags, t)
		} else {
			nonVersionTags = append(nonVersionTags, t)
		}
	}
	sort.Strings(nonVersionTags)
	return versionTags, nonVersionTags
}

func sortVersions(rawTags *[]string) (sortedTags []*version.Version, err error) {
	for _, t := range *rawTags {
		v, err := version.NewVersion(t)

		if err != nil {
			if strings.Contains(err.Error(), "Malformed version:") {
				log.Println("Received malformed error:", err)
				err = nil
			} else {
				return sortedTags, err
			}
		} else {
			sortedTags = append(sortedTags, v)
		}
	}
	sort.Sort(version.Collection(sortedTags))
	return sortedTags, err
}

func (i *InputImage) checkTagsFromPublicRepo(inputTags *[]string, maxResults int) (result []string, err error) {
	maxResults = i.maxResults(maxResults)
	noFilter := i.checkFilter()
	versionTags, nonVersionTags := parseVersions(inputTags)
	sortedTags, err := sortVersions(&versionTags)

	if err != nil {
		return result, err
	}
	versionConstraint, err := i.createConstraint()

	if err != nil {
		return result, err
	}

	// go through non version tags like latest/current/stable
	if len(nonVersionTags) > 0 {
		for _, t := range nonVersionTags {
			if maxResults == 0 {
				break
			}
			switch {
			case noFilter && maxResults != 0:
				result = append(result, t)
				maxResults--
			case i.checkNonVersionTags(t):
				result = append(result, t)
				maxResults--
			}
		}
	}

	// go through correct versioned tags
	for x := len(sortedTags) - 1; x != -1; {
		if maxResults == 0 {
			break
		}
		switch {
		case noFilter && maxResults != 0:
			result = append(result, (sortedTags)[x].Original())
			maxResults--
		case i.checkVersionTags((sortedTags)[x], &versionConstraint) && maxResults != 0:
			result = append(result, (sortedTags)[x].Original())
			maxResults--
		}
		x--
	}
	return result, err
}