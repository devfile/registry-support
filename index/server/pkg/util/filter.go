//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"regexp"
	"strings"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	versionpkg "github.com/hashicorp/go-version"
	"github.com/mohae/deepcopy"
)

const (
	/* Array empty strategies */

	// Filters out entries with empty field
	ARRAY_FILTER_IF_EMPTY = iota
	// Omits entries from filtering with empty field (leaves entries in result)
	ARRAY_SKIP_IF_EMPTY = iota

	/* Parameter Names */

	// Parameter 'name'
	PARAM_NAME = "name"
	// Parameter 'displayName'
	PARAM_DISPLAY_NAME = "displayName"
	// Parameter 'description'
	PARAM_DESCRIPTION = "description"
	// Parameter 'icon'
	PARAM_ICON = "icon"
	// Parameter 'projectType'
	PARAM_PROJECT_TYPE = "projectType"
	// Parameter 'language'
	PARAM_LANGUAGE = "language"
	// Parameter 'version'
	PARAM_VERSION = "version"
	// Parameter 'schemaVersion'
	PARAM_SCHEMA_VERSION = "schemaVersion"
	// Parameter 'default'
	PARAM_DEFAULT = "default"
	// Parameter 'git.url'
	PARAM_GIT_URL = "git.url"
	// Parameter 'git.remoteName'
	PARAM_GIT_REMOTE_NAME = "git.remoteName"
	// Parameter 'git.subDir'
	PARAM_GIT_SUBDIR = "git.subDir"
	// Parameter 'git.revision'
	PARAM_GIT_REVISION = "git.revision"
	// Parameter 'provider'
	PARAM_PROVIDER = "provider"
	// Parameter 'supportUrl'
	PARAM_SUPPORT_URL = "supportUrl"

	/* Array Parameter Names */

	// Parameter 'attributes'
	ARRAY_PARAM_ATTRIBUTE_NAMES = "attributes"
	// Parameter 'tags'
	ARRAY_PARAM_TAGS = "tags"
	// Parameter 'architectures'
	ARRAY_PARAM_ARCHITECTURES = "architectures"
	// Parameter 'resources'
	ARRAY_PARAM_RESOURCES = "resources"
	// Parameter 'starterProjects'
	ARRAY_PARAM_STARTER_PROJECTS = "starterProjects"
	// Parameter 'links'
	ARRAY_PARAM_LINKS = "links"
	// Parameter 'commandGroups'
	ARRAY_PARAM_COMMAND_GROUPS = "commandGroups"
	// Parameter 'git.remotes'
	ARRAY_PARAM_GIT_REMOTES = "git.remotes"
)

// filterOut filters out element at i in a given referenced array,
// if element does not exist the given referenced array does not
// change.
func filterOut[T any](arr *[]T, i *int) {
	if len(*arr) <= *i {
		return
	}

	// if the requested value is not present, filter it out
	*arr = append((*arr)[:*i], (*arr)[*i+1:]...)

	// decrement counter, since we shifted the array
	*i--
}

func trimExtraSpace(s string) string {
	re := regexp.MustCompile(`\s+`)
	splitStr := re.Split(strings.TrimSpace(s), -1)
	for i := 0; i < len(splitStr); i++ {
		if splitStr[i] == "" {
			filterOut(&splitStr, &i)
		}
	}
	return strings.Join(splitStr, " ")
}

func trimPunc(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// fuzzyMatch
func fuzzyMatch(a string, b string) bool {
	aLower, bLower := strings.ToLower(a), strings.ToLower(b)
	aTrim, bTrim := trimPunc(trimExtraSpace(aLower)), trimPunc(trimExtraSpace(bLower))
	return strings.Contains(aTrim, bTrim)
}

// filterDevfileDescription filters devfiles based on the description field
func filterDevfileFieldFuzzy(index []indexSchema.Schema, requestedValue string, getIndexValue func(*indexSchema.Schema) (string, error),
	getVersionValue func(*indexSchema.Version) (string, error), v1Index bool) ([]indexSchema.Schema, error) {

	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)

	if getIndexValue != nil || getVersionValue != nil {
		for i := 0; i < len(filteredIndex); i++ {
			toFilterOutIndex := false

			if getIndexValue != nil {
				indexValue, err := getIndexValue(&filteredIndex[i])
				if err != nil {
					return []indexSchema.Schema{}, err
				} else if !fuzzyMatch(indexValue, requestedValue) {
					toFilterOutIndex = true
				}
			} else {
				toFilterOutIndex = true
			}

			if !v1Index && getVersionValue != nil {
				filteredVersions := deepcopy.Copy(filteredIndex[i].Versions).([]indexSchema.Version)
				for versionIndex := 0; versionIndex < len(filteredVersions); versionIndex++ {
					versionValue, err := getVersionValue(&filteredVersions[versionIndex])
					if err != nil {
						return []indexSchema.Schema{}, err
					} else if !fuzzyMatch(versionValue, requestedValue) {
						filterOut(&filteredVersions, &versionIndex)
					}
				}

				if len(filteredVersions) != 0 {
					filteredIndex[i].Versions = filteredVersions
					toFilterOutIndex = false
				}
			}

			if toFilterOutIndex {
				filterOut(&filteredIndex, &i)
			}
		}
	}

	return filteredIndex, nil
}

// filterDevfileTags filters devfiles based on tags
func filterDevfileTags(index []indexSchema.Schema, tags []string, v1Index bool) []indexSchema.Schema {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)
	for i := 0; i < len(filteredIndex); i++ {
		if len(tags) != 0 && len(filteredIndex[i].Tags) == 0 {
			// If tags are requested and a stack has no tags mentioned, then filter out entry
			filterOut(&filteredIndex, &i)
			continue
		}

		filterIn := true
		tagsInIndex := StrArrayToSetMap(filteredIndex[i].Tags)

		for _, requestedTag := range tags {
			if !tagsInIndex.Has(requestedTag) {
				filterIn = false
				break
			}
		}

		if !filterIn {
			filterOut(&filteredIndex, &i)
			continue
		}

		// go through each version's tags if multi-version stack is supported
		if !v1Index {
			for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
				versionTags := filteredIndex[i].Versions[versionIndex].Tags
				if len(tags) != 0 && len(versionTags) == 0 {
					// If tags are requested and a stack has no tags mentioned, then filter out entry
					filterOut(&filteredIndex[i].Versions, &versionIndex)
					continue
				}
				filterVersion := true
				tagsInVersion := StrArrayToSetMap(filteredIndex[i].Versions[versionIndex].Tags)

				for _, requestedTag := range tags {
					if !tagsInVersion.Has(requestedTag) {
						filterVersion = false
						break
					}
				}

				if !filterVersion {
					filterOut(&filteredIndex[i].Versions, &versionIndex)
				}

			}
		}
	}

	return filteredIndex
}

// filterDevfileArchitectures filters devfiles based on architectures
func filterDevfileArchitectures(index []indexSchema.Schema, archs []string, v1Index bool) []indexSchema.Schema {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)
	for i := 0; i < len(filteredIndex); i++ {
		if len(filteredIndex[i].Architectures) == 0 {
			// If a stack has no architectures mentioned, then it supports all architectures
			continue
		}

		filterIn := true

		for _, requestedArch := range archs {
			isArchPresent := false
			for _, devfileArch := range filteredIndex[i].Architectures {
				if requestedArch == devfileArch {
					isArchPresent = true
					break
				}
			}

			if !isArchPresent {
				// if one of the arch requested is not present, no need to search for the others
				filterIn = false
				break
			}
		}

		if !filterIn {
			filterOut(&filteredIndex, &i)
			continue
		}

		// go through each version's architecture if multi-version stack is supported
		if !v1Index {
			for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
				versionArchs := filteredIndex[i].Versions[versionIndex].Architectures
				if len(versionArchs) == 0 {
					// If a devfile has no architectures mentioned, then it supports all architectures
					continue
				}
				archInVersion := true
				for _, requestedArch := range archs {
					archPresentInVersion := false
					for _, versionArch := range versionArchs {
						if requestedArch == versionArch {
							archPresentInVersion = true
							break
						}
					}
					if !archPresentInVersion {
						// if one of the arch requested is not present, no need to search for the others
						archInVersion = false
						break
					}
				}

				if !archInVersion {
					filterOut(&filteredIndex[i].Versions, &versionIndex)
				}

			}
		}
	}

	return filteredIndex
}

// FilterDevfileSchemaVersion filters devfiles based on schema version
func FilterDevfileSchemaVersion(index []indexSchema.Schema, minSchemaVersion string, maxSchemaVersion string) ([]indexSchema.Schema, error) {
	for i := 0; i < len(index); i++ {
		for versionIndex := 0; versionIndex < len(index[i].Versions); versionIndex++ {
			currectSchemaVersion := index[i].Versions[versionIndex].SchemaVersion
			schemaVersionWithoutServiceVersion := currectSchemaVersion[:strings.LastIndex(currectSchemaVersion, ".")]
			curVersion, err := versionpkg.NewVersion(schemaVersionWithoutServiceVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to parse schemaVersion %s for stack: %s, version %s. Error: %v", currectSchemaVersion, index[i].Name, index[i].Versions[versionIndex].Version, err)
			}

			versionInRange := true
			if minSchemaVersion != "" {
				minVersion, err := versionpkg.NewVersion(minSchemaVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse minSchemaVersion %s. Error: %v", minSchemaVersion, err)
				}
				if minVersion.GreaterThan(curVersion) {
					versionInRange = false
				}
			}
			if versionInRange && maxSchemaVersion != "" {
				maxVersion, err := versionpkg.NewVersion(maxSchemaVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse maxSchemaVersion %s. Error: %v", maxSchemaVersion, err)
				}
				if maxVersion.LessThan(curVersion) {
					versionInRange = false
				}
			}
			if !versionInRange {
				// if schemaVersion is not in requested range, filter it out
				index[i].Versions = append(index[i].Versions[:versionIndex], index[i].Versions[versionIndex+1:]...)

				// decrement counter, since we shifted the array
				versionIndex--
			}
		}
		if len(index[i].Versions) == 0 {
			// if versions list is empty after filter, remove this index
			index = append(index[:i], index[i+1:]...)

			// decrement counter, since we shifted the array
			i--
		}
	}

	return index, nil
}

// FilterDevfileStrField
func FilterDevfileStrField(index []indexSchema.Schema, paramName string, requestedValue string, v1Index bool) ([]indexSchema.Schema, error) {
	var getIndexValue func(*indexSchema.Schema) (string, error)
	var getVersionValue func(*indexSchema.Version) (string, error)
	switch paramName {
	case PARAM_NAME:
		getIndexValue = indexSchema.GetName
		getVersionValue = nil
	case PARAM_DISPLAY_NAME:
		getIndexValue = indexSchema.GetDisplayName
		getVersionValue = nil
	case PARAM_DESCRIPTION:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			description, err := indexSchema.GetDescription(s)
			if err != nil {
				return "", err
			}

			return description, nil
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			description, err := indexSchema.GetDescription(v)
			if err != nil {
				return "", err
			}

			return description, nil
		}
	case PARAM_ICON:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			icon, err := indexSchema.GetIcon(s)
			if err != nil {
				return "", err
			}

			return icon, nil
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			icon, err := indexSchema.GetIcon(v)
			if err != nil {
				return "", err
			}

			return icon, nil
		}
	case PARAM_PROJECT_TYPE:
		getIndexValue = indexSchema.GetProjectType
		getVersionValue = nil
	case PARAM_LANGUAGE:
		getIndexValue = indexSchema.GetLanguage
		getVersionValue = nil
	case PARAM_VERSION:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			return indexSchema.GetVersion(s)
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			return indexSchema.GetVersion(v)
		}
	case PARAM_SCHEMA_VERSION:
		getIndexValue = nil
		getVersionValue = indexSchema.GetSchemaVersion
	case PARAM_GIT_URL:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			return indexSchema.GetGitUrl(s)
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			return indexSchema.GetGitUrl(v)
		}
	case PARAM_GIT_REMOTE_NAME:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			return indexSchema.GetGitRemoteName(s)
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			return indexSchema.GetGitRemoteName(v)
		}
	case PARAM_GIT_SUBDIR:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			return indexSchema.GetGitSubDir(s)
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			return indexSchema.GetGitSubDir(v)
		}
	case PARAM_GIT_REVISION:
		getIndexValue = func(s *indexSchema.Schema) (string, error) {
			return indexSchema.GetGitRevision(s)
		}
		getVersionValue = func(v *indexSchema.Version) (string, error) {
			return indexSchema.GetGitRevision(v)
		}
	case PARAM_PROVIDER:
		getIndexValue = indexSchema.GetProvider
		getVersionValue = nil
	case PARAM_SUPPORT_URL:
		getIndexValue = indexSchema.GetSupportUrl
		getVersionValue = nil
	default:
		return index, nil
	}

	return filterDevfileFieldFuzzy(index, requestedValue, getIndexValue, getVersionValue, v1Index)
}

func AndFilter(results [][]indexSchema.Schema) ([]indexSchema.Schema, error) {
	panic("not implemented")
}

// FilterDevfileStrArrayField filters devfiles based on an array field
func FilterDevfileStrArrayField(index []indexSchema.Schema, paramName string, requestedValues []string, v1Index bool) []indexSchema.Schema {
	switch paramName {
	case ARRAY_PARAM_ARCHITECTURES:
		return filterDevfileArchitectures(index, requestedValues, v1Index)
	case ARRAY_PARAM_TAGS:
		return filterDevfileTags(index, requestedValues, v1Index)
	default:
		return index
	}
}
