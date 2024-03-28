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
	sets "github.com/hashicorp/go-set"
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
	PARAM_ICON = "iconUri"
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
	PARAM_GIT_URL = "gitUrl"
	// Parameter 'git.remoteName'
	PARAM_GIT_REMOTE_NAME = "gitRemoteName"
	// Parameter 'git.subDir'
	PARAM_GIT_SUBDIR = "gitSubDir"
	// Parameter 'git.revision'
	PARAM_GIT_REVISION = "gitRevision"
	// Parameter 'provider'
	PARAM_PROVIDER = "provider"
	// Parameter 'supportUrl'
	PARAM_SUPPORT_URL = "supportUrl"

	/* Array Parameter Names */

	// Parameter 'attributeNames'
	ARRAY_PARAM_ATTRIBUTE_NAMES = "attributeNames"
	// Parameter 'tags'
	ARRAY_PARAM_TAGS = "tags"
	// Parameter 'architectures'
	ARRAY_PARAM_ARCHITECTURES = "arch"
	// Parameter 'resources'
	ARRAY_PARAM_RESOURCES = "resources"
	// Parameter 'starterProjects'
	ARRAY_PARAM_STARTER_PROJECTS = "starterProjects"
	// Parameter 'links'
	ARRAY_PARAM_LINKS = "links"
	// Parameter 'commandGroups'
	ARRAY_PARAM_COMMAND_GROUPS = "commandGroups"
	// Parameter 'gitRemoteNames'
	ARRAY_PARAM_GIT_REMOTE_NAMES = "gitRemoteNames"
	// Parameter 'gitRemotes'
	ARRAY_PARAM_GIT_REMOTES = "gitRemotes"
)

// FilterResult result entity of filtering the index schema
type FilterResult struct {
	// Name of filter
	Name string
	// Index schema result
	Index []indexSchema.Schema
	// First error returned in result
	Error error
}

// FilterOptions provides filtering options to filters operations
type FilterOptions[T any] struct {
	GetFromIndexField   func(*indexSchema.Schema) T
	GetFromVersionField func(*indexSchema.Version) T
	FilterOutEmpty      bool
	V1Index             bool
}

// indexFieldEmptyHandler handles what to do with empty index array fields
func indexFieldEmptyHandler(fieldValues []string, requestedValues []string, options FilterOptions[[]string]) bool {
	// If filtering out empty, assume that if a stack has no field values mentioned and field value are request
	// Else assume that if a stack has no array field values mentioned, then assume all possible are valid
	if options.FilterOutEmpty {
		return len(requestedValues) != 0 && len(fieldValues) == 0
	} else {
		return len(fieldValues) == 0
	}
}

// versionFieldEmptyHandler handles what to do with empty version array fields
func versionFieldEmptyHandler(fieldValues []string, requestedValues []string, options FilterOptions[[]string]) bool {
	// If filtering out empty, assume that if a stack has no field values mentioned and field value are request
	// Else assume that if a stack has no array field values mentioned, then assume all possible are valid
	if options.FilterOutEmpty {
		return len(requestedValues) != 0 && len(fieldValues) == 0
	} else {
		return len(fieldValues) == 0
	}
}

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

// trimExtraSpace Trims extra whitespace from string
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

// preProcessString pre-process give string to perform fuzzy matching
func preProcessString(s string) string {
	sLower := strings.ToLower(s)
	return trimExtraSpace(sLower)
}

// fuzzyMatch fuzzy compare function
func fuzzyMatch(a, b string) bool {
	return strings.Contains(preProcessString(a), preProcessString(b))
}

// filterDevfileFieldFuzzy filters devfiles based on fuzzy filtering of string fields
func filterDevfileFieldFuzzy(index []indexSchema.Schema, requestedValue string, options FilterOptions[string]) []indexSchema.Schema {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)

	if options.GetFromIndexField != nil || options.GetFromVersionField != nil {
		for i := 0; i < len(filteredIndex); i++ {
			toFilterOutIndex := false

			if options.GetFromIndexField != nil {
				indexValue := options.GetFromIndexField(&filteredIndex[i])
				if !fuzzyMatch(indexValue, requestedValue) {
					toFilterOutIndex = true
				}
			} else {
				toFilterOutIndex = true
			}

			if !options.V1Index && options.GetFromVersionField != nil {
				filteredVersions := deepcopy.Copy(filteredIndex[i].Versions).([]indexSchema.Version)
				for versionIndex := 0; versionIndex < len(filteredVersions); versionIndex++ {
					versionValue := options.GetFromVersionField(&filteredVersions[versionIndex])
					if !fuzzyMatch(versionValue, requestedValue) {
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

	return filteredIndex
}

// filterDevfileFieldFuzzy filters devfiles based on fuzzy filtering of string array fields
func filterDevfileArrayFuzzy(index []indexSchema.Schema, requestedValues []string, options FilterOptions[[]string]) []indexSchema.Schema {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)

	if options.GetFromIndexField != nil || options.GetFromVersionField != nil {
		for i := 0; i < len(filteredIndex); i++ {
			toFilterOutIndex := false

			if options.GetFromIndexField != nil {
				fieldValues := options.GetFromIndexField(&filteredIndex[i])

				// If index schema field is not empty perform fuzzy filtering
				// else if filtering out based on empty fields is set, set index schema to be filtered out
				// (after version filtering if applicable)
				if !indexFieldEmptyHandler(fieldValues, requestedValues, options) {
					matchAll := true
					for _, requestedValue := range requestedValues {
						matchFound := false

						for _, fieldValue := range fieldValues {
							if fuzzyMatch(fieldValue, requestedValue) {
								matchFound = true
								break
							}
						}

						if !matchFound {
							matchAll = false
							break
						}
					}
					if !matchAll {
						toFilterOutIndex = true
					}
				} else if options.FilterOutEmpty {
					toFilterOutIndex = true
				}
			}

			// go through each version's tags if multi-version stack is supported
			if !options.V1Index && options.GetFromVersionField != nil {
				filteredVersions := deepcopy.Copy(filteredIndex[i].Versions).([]indexSchema.Version)
				for versionIndex := 0; versionIndex < len(filteredVersions); versionIndex++ {
					fieldValues := options.GetFromVersionField(&filteredVersions[versionIndex])

					// If version schema field is not empty perform fuzzy filtering
					// else if filtering out based on empty fields is set, filter out version schema
					if !versionFieldEmptyHandler(fieldValues, requestedValues, options) {
						matchAll := true
						for _, requestedValue := range requestedValues {
							matchFound := false

							for _, fieldValue := range fieldValues {
								if fuzzyMatch(fieldValue, requestedValue) {
									matchFound = true
									break
								}
							}

							if !matchFound {
								matchAll = false
								break
							}
						}
						if !matchAll {
							filterOut(&filteredVersions, &versionIndex)
						}
					} else if options.FilterOutEmpty {
						filterOut(&filteredVersions, &versionIndex)
					}
				}

				// If the filtered versions is not empty, set to the filtered index result
				// Else if empty and index field filtering was not performed ensure index entry is filtered out
				if len(filteredVersions) != 0 {
					filteredIndex[i].Versions = filteredVersions
					toFilterOutIndex = false
				} else if options.GetFromIndexField == nil {
					toFilterOutIndex = true
				}
			}

			if toFilterOutIndex {
				filterOut(&filteredIndex, &i)
			}
		}
	}

	return filteredIndex
}

func IsFieldParameter(name string) bool {
	parameterNames := sets.From([]string{
		PARAM_NAME,
		PARAM_DISPLAY_NAME,
		PARAM_DESCRIPTION,
		PARAM_ICON,
		PARAM_PROJECT_TYPE,
		PARAM_LANGUAGE,
		PARAM_VERSION,
		PARAM_SCHEMA_VERSION,
		PARAM_DEFAULT,
		PARAM_GIT_URL,
		PARAM_GIT_REMOTE_NAME,
		PARAM_GIT_SUBDIR,
		PARAM_GIT_REVISION,
		PARAM_PROVIDER,
		PARAM_SUPPORT_URL,
	})

	return parameterNames.Contains(name)
}

func IsArrayParameter(name string) bool {
	parameterNames := sets.From([]string{
		ARRAY_PARAM_ATTRIBUTE_NAMES,
		ARRAY_PARAM_ARCHITECTURES,
		ARRAY_PARAM_TAGS,
		ARRAY_PARAM_RESOURCES,
		ARRAY_PARAM_STARTER_PROJECTS,
		ARRAY_PARAM_LINKS,
		ARRAY_PARAM_COMMAND_GROUPS,
		ARRAY_PARAM_GIT_REMOTE_NAMES,
		ARRAY_PARAM_GIT_REMOTES,
	})

	return parameterNames.Contains(name)
}

// FilterDevfileSchemaVersion filters devfiles based on schema version
func FilterDevfileSchemaVersion(index []indexSchema.Schema, minSchemaVersion, maxSchemaVersion *string) ([]indexSchema.Schema, error) {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)
	for i := 0; i < len(filteredIndex); i++ {
		for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
			currentSchemaVersion := filteredIndex[i].Versions[versionIndex].SchemaVersion
			curVersion, err := versionpkg.NewVersion(currentSchemaVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to parse schemaVersion %s for stack: %s, version %s. Error: %v", currentSchemaVersion, filteredIndex[i].Name, index[i].Versions[versionIndex].Version, err)
			}

			versionInRange := true
			if StrPtrIsSet(minSchemaVersion) {
				minVersion, err := versionpkg.NewVersion(*minSchemaVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse minSchemaVersion %s. Error: %v", *minSchemaVersion, err)
				}
				if minVersion.GreaterThan(curVersion) {
					versionInRange = false
				}
			}
			if versionInRange && StrPtrIsSet(maxSchemaVersion) {
				maxVersion, err := versionpkg.NewVersion(*maxSchemaVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse maxSchemaVersion %s. Error: %v", *maxSchemaVersion, err)
				}
				if maxVersion.LessThan(curVersion) {
					versionInRange = false
				}
			}
			if !versionInRange {
				// if schemaVersion is not in requested range, filter it out
				filterOut(&filteredIndex[i].Versions, &versionIndex)
			}
		}
		if len(filteredIndex[i].Versions) == 0 {
			// if versions list is empty after filter, remove this index
			filterOut(&filteredIndex, &i)
		}
	}

	return filteredIndex, nil
}

// FilterDevfileVersion filters devfiles based on stack version
func FilterDevfileVersion(index []indexSchema.Schema, minVersion, maxVersion *string) ([]indexSchema.Schema, error) {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)
	for i := 0; i < len(filteredIndex); i++ {
		for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
			currentVersion := filteredIndex[i].Versions[versionIndex].Version
			curVersion, err := versionpkg.NewVersion(currentVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to parse version %s for stack: %s. Error: %v", currentVersion, filteredIndex[i].Name, err)
			}

			versionInRange := true
			if StrPtrIsSet(minVersion) {
				minVersion, err := versionpkg.NewVersion(*minVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse minVersion %s. Error: %v", minVersion, err)
				}
				if minVersion.GreaterThan(curVersion) {
					versionInRange = false
				}
			}
			if versionInRange && StrPtrIsSet(maxVersion) {
				maxVersion, err := versionpkg.NewVersion(*maxVersion)
				if err != nil {
					return nil, fmt.Errorf("failed to parse maxVersion %s. Error: %v", maxVersion, err)
				}
				if maxVersion.LessThan(curVersion) {
					versionInRange = false
				}
			}
			if !versionInRange {
				// if version is not in requested range, filter it out
				filterOut(&filteredIndex[i].Versions, &versionIndex)
			}
		}
		if len(filteredIndex[i].Versions) == 0 {
			// if versions list is empty after filter, remove this index
			filterOut(&filteredIndex, &i)
		}
	}

	return filteredIndex, nil
}

// FilterDevfileStrField filters by given string field, returns unchanged index if given parameter name is unrecognized
func FilterDevfileStrField(index []indexSchema.Schema, paramName, requestedValue string, v1Index bool) FilterResult {
	filterName := fmt.Sprintf("Fuzzy_Field_Filter_On_%s", paramName)
	options := FilterOptions[string]{
		V1Index: v1Index,
	}
	switch paramName {
	case PARAM_NAME:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Name
		}
	case PARAM_DISPLAY_NAME:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.DisplayName
		}
	case PARAM_DESCRIPTION:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Description
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Description
		}
	case PARAM_ICON:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Icon
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Icon
		}
	case PARAM_PROJECT_TYPE:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.ProjectType
		}
	case PARAM_LANGUAGE:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Language
		}
	case PARAM_VERSION:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Version
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Version
		}
	case PARAM_SCHEMA_VERSION:
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.SchemaVersion
		}
	case PARAM_DEFAULT:
		options.GetFromIndexField = nil
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return fmt.Sprintf("%v", v.Default)
		}
	case PARAM_GIT_URL:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.Url
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.Url
		}
	case PARAM_GIT_REMOTE_NAME:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.RemoteName
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.RemoteName
		}
	case PARAM_GIT_SUBDIR:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.SubDir
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.SubDir
		}
	case PARAM_GIT_REVISION:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.Revision
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.Revision
		}
	case PARAM_PROVIDER:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Provider
		}
	case PARAM_SUPPORT_URL:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.SupportUrl
		}
	default:
		return FilterResult{
			Name:  filterName,
			Index: index,
		}
	}

	return FilterResult{
		Name:  filterName,
		Index: filterDevfileFieldFuzzy(index, requestedValue, options),
	}
}

// AndFilter filters results of given filters to only overlapping results
func AndFilter(results ...*FilterResult) FilterResult {
	schemaCounts := map[string]*struct {
		count  int
		schema indexSchema.Schema
	}{}
	resultNames := func() []string {
		names := []string{}
		for _, result := range results {
			names = append(names, result.Name)
		}
		return names
	}()
	andResult := FilterResult{
		Name:  fmt.Sprintf("And(%s)", strings.Join(resultNames, ", ")),
		Index: []indexSchema.Schema{},
	}

	for _, result := range results {
		// If a filter returns an error, return as overall result
		if result.Error != nil {
			andResult.Error = fmt.Errorf("filter failed on '%s': %v", result.Name, result.Error)
			return andResult
		}

		for _, schema := range result.Index {
			schemaCount, found := schemaCounts[schema.Name]
			// if not found, initize is a seen counter of one and the current seen schema
			// else increment seen counter and re-assign current seen schema if versions have been filtered
			if !found {
				schemaCounts[schema.Name] = &struct {
					count  int
					schema indexSchema.Schema
				}{
					count:  1,
					schema: schema,
				}
			} else {
				schemaCounts[schema.Name].count += 1
				if len(schema.Versions) < len(schemaCount.schema.Versions) {
					schemaCounts[schema.Name].schema = schema
				}
			}
		}
	}

	// build results of filters into new index schema
	for _, v := range schemaCounts {
		// if result is in every filter result then add to array
		if v.count == len(results) {
			andResult.Index = append(andResult.Index, v.schema)
		}
	}

	return andResult
}

// FilterDevfileStrArrayField filters devfiles based on an array field
func FilterDevfileStrArrayField(index []indexSchema.Schema, paramName string, requestedValues []string, v1Index bool) FilterResult {
	filterName := fmt.Sprintf("Fuzzy_Array_Filter_On_%s", paramName)
	options := FilterOptions[[]string]{
		FilterOutEmpty: true,
		V1Index:        v1Index,
	}
	switch paramName {
	case ARRAY_PARAM_ATTRIBUTE_NAMES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			names := []string{}

			for name := range s.Attributes {
				names = append(names, name)
			}

			return names
		}
	case ARRAY_PARAM_ARCHITECTURES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Architectures
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Architectures
		}
		options.FilterOutEmpty = false
	case ARRAY_PARAM_TAGS:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Tags
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Tags
		}
	case ARRAY_PARAM_RESOURCES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Resources
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Resources
		}
	case ARRAY_PARAM_STARTER_PROJECTS:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.StarterProjects
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.StarterProjects
		}
	case ARRAY_PARAM_LINKS:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			links := []string{}

			for linkName := range s.Links {
				links = append(links, linkName)
			}

			return links
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			links := []string{}

			for linkName := range v.Links {
				links = append(links, linkName)
			}

			return links
		}
	case ARRAY_PARAM_COMMAND_GROUPS:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			commandGroups := []string{}

			for commandGroup, isSet := range s.CommandGroups {
				if isSet {
					commandGroups = append(commandGroups, string(commandGroup))
				}
			}

			return commandGroups
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			commandGroups := []string{}

			for commandGroup, isSet := range v.CommandGroups {
				if isSet {
					commandGroups = append(commandGroups, string(commandGroup))
				}
			}

			return commandGroups
		}
	case ARRAY_PARAM_GIT_REMOTE_NAMES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			gitRemoteNames := []string{}

			if s.Git != nil {
				for remoteName := range s.Git.Remotes {
					gitRemoteNames = append(gitRemoteNames, remoteName)
				}
			}

			return gitRemoteNames
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			gitRemoteNames := []string{}

			if v.Git != nil {
				for remoteName := range v.Git.Remotes {
					gitRemoteNames = append(gitRemoteNames, remoteName)
				}
			}

			return gitRemoteNames
		}
	case ARRAY_PARAM_GIT_REMOTES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			gitRemotes := []string{}

			if s.Git != nil {
				for _, remoteUrl := range s.Git.Remotes {
					gitRemotes = append(gitRemotes, remoteUrl)
				}
			}

			return gitRemotes
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			gitRemotes := []string{}

			if v.Git != nil {
				for _, remoteUrl := range v.Git.Remotes {
					gitRemotes = append(gitRemotes, remoteUrl)
				}
			}

			return gitRemotes
		}
	default:
		return FilterResult{
			Name:  filterName,
			Index: index,
		}
	}

	return FilterResult{
		Name:  filterName,
		Index: filterDevfileArrayFuzzy(index, requestedValues, options),
	}
}
