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
	ArrayFilterIfEmpty = iota
	// Omits entries from filtering with empty field (leaves entries in result)
	ArraySkipIfEmpty = iota

	/* Parameter Names */

	// Parameter 'name'
	ParamName = "name"
	// Parameter 'displayName'
	ParamDisplayName = "displayName"
	// Parameter 'description'
	ParamDescription = "description"
	// Parameter 'icon'
	ParamIcon = "iconUri"
	// Parameter 'projectType'
	ParamProjectType = "projectType"
	// Parameter 'language'
	ParamLanguage = "language"
	// Parameter 'version'
	ParamVersion = "version"
	// Parameter 'schemaVersion'
	ParamSchemaVersion = "schemaVersion"
	// Parameter 'default'
	ParamDefault = "default"
	// Parameter 'git.url'
	ParamGitUrl = "gitUrl"
	// Parameter 'git.remoteName'
	ParamGitRemoteName = "gitRemoteName"
	// Parameter 'git.subDir'
	ParamGitSubDir = "gitSubDir"
	// Parameter 'git.revision'
	ParamGitRevision = "gitRevision"
	// Parameter 'provider'
	ParamProvider = "provider"
	// Parameter 'supportUrl'
	ParamSupportUrl = "supportUrl"
	// Parameter 'lastModified'
	ParamLastModified = "lastModified"

	/* Array Parameter Names */

	// Parameter 'attributeNames'
	ArrayParamAttributeNames = "attributeNames"
	// Parameter 'tags'
	ArrayParamTags = "tags"
	// Parameter 'architectures'
	ArrayParamArchitectures = "arch"
	// Parameter 'resources'
	ArrayParamResources = "resources"
	// Parameter 'starterProjects'
	ArrayParamStarterProjects = "starterProjects"
	// Parameter 'links'
	ArrayParamLinks = "links"
	// Parameter 'commandGroups'
	ArrayParamCommandGroups = "commandGroups"
	// Parameter 'deploymentScopes'
	ArrayParamDeploymentScopes = "deploymentScopes"
	// Parameter 'gitRemoteNames'
	ArrayParamGitRemoteNames = "gitRemoteNames"
	// Parameter 'gitRemotes'
	ArrayParamGitRemotes = "gitRemotes"
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
		ParamName,
		ParamDisplayName,
		ParamDescription,
		ParamIcon,
		ParamProjectType,
		ParamLanguage,
		ParamVersion,
		ParamSchemaVersion,
		ParamDefault,
		ParamGitUrl,
		ParamGitRemoteName,
		ParamGitSubDir,
		ParamGitRevision,
		ParamProvider,
		ParamSupportUrl,
	})

	return parameterNames.Contains(name)
}

func IsArrayParameter(name string) bool {
	parameterNames := sets.From([]string{
		ArrayParamAttributeNames,
		ArrayParamArchitectures,
		ArrayParamTags,
		ArrayParamResources,
		ArrayParamStarterProjects,
		ArrayParamLinks,
		ArrayParamCommandGroups,
		ArrayParamDeploymentScopes,
		ArrayParamGitRemoteNames,
		ArrayParamGitRemotes,
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

// FilterDevfileDeprecated inplace filters devfiles based on stack deprecation
func FilterDevfileDeprecated(index *[]indexSchema.Schema, deprecated, v1Index bool) {
	for i := 0; i < len(*index); i++ {
		toFilterOutIndex := !deprecated
		foundDeprecated := false

		for _, tag := range (*index)[i].Tags {
			if tag == "Deprecated" {
				foundDeprecated = true
				break
			}
		}

		if !foundDeprecated {
			toFilterOutIndex = deprecated

			if !v1Index {
				for versionIndex := 0; versionIndex < len((*index)[i].Versions); versionIndex++ {
					if (*index)[i].Versions[versionIndex].Default {
						for _, tag := range (*index)[i].Versions[versionIndex].Tags {
							if tag == "Deprecated" {
								toFilterOutIndex = !deprecated
								break
							}
						}
						break
					}
				}
			}
		}

		if toFilterOutIndex {
			filterOut(index, &i)
		}
	}
}

// FilterDevfileStrField filters by given string field, returns unchanged index if given parameter name is unrecognized
func FilterDevfileStrField(index []indexSchema.Schema, paramName, requestedValue string, v1Index bool) FilterResult {
	filterName := fmt.Sprintf("Fuzzy_Field_Filter_On_%s", paramName)
	options := FilterOptions[string]{
		V1Index: v1Index,
	}
	switch paramName {
	case ParamName:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Name
		}
	case ParamDisplayName:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.DisplayName
		}
	case ParamDescription:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Description
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Description
		}
	case ParamIcon:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Icon
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Icon
		}
	case ParamProjectType:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.ProjectType
		}
	case ParamLanguage:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Language
		}
	case ParamVersion:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Version
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Version
		}
	case ParamSchemaVersion:
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.SchemaVersion
		}
	case ParamDefault:
		options.GetFromIndexField = nil
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return fmt.Sprintf("%v", v.Default)
		}
	case ParamGitUrl:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.Url
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.Url
		}
	case ParamGitRemoteName:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.RemoteName
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.RemoteName
		}
	case ParamGitSubDir:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.SubDir
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.SubDir
		}
	case ParamGitRevision:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Git.Revision
		}
		options.GetFromVersionField = func(v *indexSchema.Version) string {
			return v.Git.Revision
		}
	case ParamProvider:
		options.GetFromIndexField = func(s *indexSchema.Schema) string {
			return s.Provider
		}
	case ParamSupportUrl:
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
	case ArrayParamAttributeNames:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			names := []string{}

			for name := range s.Attributes {
				names = append(names, name)
			}

			return names
		}
	case ArrayParamArchitectures:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Architectures
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Architectures
		}
		options.FilterOutEmpty = false
	case ArrayParamTags:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Tags
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Tags
		}
	case ArrayParamResources:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.Resources
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.Resources
		}
	case ArrayParamStarterProjects:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			return s.StarterProjects
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			return v.StarterProjects
		}
	case ArrayParamLinks:
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
	case ArrayParamCommandGroups:
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
	case ArrayParamDeploymentScopes:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			deploymentScopes := []string{}

			for deploymentScope, isSet := range s.DeploymentScopes {
				if isSet {
					deploymentScopes = append(deploymentScopes, string(deploymentScope))
				}
			}

			return deploymentScopes
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			deploymentScopes := []string{}

			for deploymentScope, isSet := range v.DeploymentScopes {
				if isSet {
					deploymentScopes = append(deploymentScopes, string(deploymentScope))
				}
			}

			return deploymentScopes
		}
	case ArrayParamGitRemoteNames:
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
	case ArrayParamGitRemotes:
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

// FilterLastModifiedDate filters based on the last modified date of a stack or sample
func FilterLastModifiedDate(index []indexSchema.Schema, minLastModified *string, maxLastModified *string) ([]indexSchema.Schema, error) {
	filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)
	for i := 0; i < len(filteredIndex); i++ {
		for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
			currentLastModifiedDate := filteredIndex[i].Versions[versionIndex].LastModified
			matchedLastModified := false
			if StrPtrIsSet(minLastModified) && StrPtrIsSet(maxLastModified) {
				minModified, err := ConvertNonRFC3339Date(*minLastModified)
				if err != nil {
					return filteredIndex, err
				}
				maxModified, err := ConvertNonRFC3339Date(*maxLastModified)
				if err != nil {
					return filteredIndex, err
				}
				curModified, err := ConvertRFC3339Date(&currentLastModifiedDate)
				if err != nil {
					return filteredIndex, err
				}
				matchedLastModified = IsDateGreaterOrEqual(minModified, curModified) && IsDateLowerOrEqual(maxModified, curModified)
			} else if StrPtrIsSet(minLastModified) {
				minModified, err := ConvertNonRFC3339Date(*minLastModified)
				if err != nil {
					return filteredIndex, err
				}
				curModified, err := ConvertRFC3339Date(&currentLastModifiedDate)
				if err != nil {
					return filteredIndex, err
				}
				matchedLastModified = IsDateGreaterOrEqual(minModified, curModified)
			} else if StrPtrIsSet(maxLastModified) {
				maxModified, err := ConvertNonRFC3339Date(*maxLastModified)
				if err != nil {
					return filteredIndex, err
				}
				curModified, err := ConvertRFC3339Date(&currentLastModifiedDate)
				if err != nil {
					return filteredIndex, err
				}
				matchedLastModified = IsDateLowerOrEqual(maxModified, curModified)
			}

			if !matchedLastModified {
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
