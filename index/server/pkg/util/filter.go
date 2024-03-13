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

// FilterResult result entity of filtering the index schema
type FilterResult struct {
	filterFn func(*FilterResult)
	children []FilterResult

	// Index schema result
	Index []indexSchema.Schema
	// First error returned in result
	Error error
	// Is the FilterResult evaluated
	IsEval bool
}

// Eval evaluates the filter results
func (fr *FilterResult) Eval() {
	fr.filterFn(fr)
	fr.IsEval = true
}

// FilterOptions provides filtering options to filters operations
type FilterOptions[T any] struct {
	GetFromIndexField   func(*indexSchema.Schema) T
	GetFromVersionField func(*indexSchema.Version) T
	FilterOutEmpty      bool
	V1Index             bool
}

// indexFieldEmptyHandler handles what to do with empty index array fields
func indexFieldEmptyHandler(index *[]indexSchema.Schema, idx *int, requestedValues []string, options FilterOptions[[]string]) bool {
	schema := &(*index)[*idx]

	// If filtering out empty, assume that if a stack has no field values mentioned and field value are request, then filter out entry
	if options.FilterOutEmpty && len(requestedValues) != 0 && len(options.GetFromIndexField(schema)) == 0 {
		filterOut(index, idx)
		return true
	}

	// Else assume that if a stack has no array field values mentioned, then assume all possible are valid
	return len(options.GetFromIndexField(schema)) == 0
}

// versionFieldEmptyHandler handles what to do with empty version array fields
func versionFieldEmptyHandler(versions *[]indexSchema.Version, idx *int, requestedValues []string, options FilterOptions[[]string]) bool {
	version := &(*versions)[*idx]

	// If filtering out empty, assume that if a stack has no field values mentioned and field value are request, then filter out entry
	if options.FilterOutEmpty && len(requestedValues) != 0 && len(options.GetFromVersionField(version)) == 0 {
		filterOut(versions, idx)
		return true
	}

	// Else assume that if a stack has no array field values mentioned, then assume all possible are valid
	return len(options.GetFromVersionField(version)) == 0
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

// trimPunc Trims punctuation from a string
func trimPunc(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// preProcessString pre-process give string to perform fuzzy matching
func preProcessString(s string) string {
	sLower := strings.ToLower(s)
	return trimPunc(trimExtraSpace(sLower))
}

// getFuzzySetFromArray gets a fuzzy pre-processed set from given array
func getFuzzySetFromArray(arr []string) *sets.Set[string] {
	preProcessedArray := []string{}

	for i := 0; i < len(arr); i++ {
		preProcessedArray = append(preProcessedArray, preProcessString(arr[i]))
	}

	return sets.From(preProcessedArray)
}

// fuzzyMatch fuzzy compare function
func fuzzyMatch(a, b string) bool {
	return strings.Contains(preProcessString(a), preProcessString(b))
}

// fuzzyMatchInSet fuzzy compare function on fuzzy pre-processed set
func fuzzyMatchInSet(fuzzySet *sets.Set[string], matchVal string) bool {
	return fuzzySet.Contains(preProcessString(matchVal))
}

// filterDevfileFieldFuzzy filters devfiles based on fuzzy filtering of string fields
func filterDevfileFieldFuzzy(index []indexSchema.Schema, requestedValue string, options FilterOptions[string]) FilterResult {
	return FilterResult{
		filterFn: func(fr *FilterResult) {
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

			fr.Index = filteredIndex
		},
	}
}

// filterDevfileFieldFuzzy filters devfiles based on fuzzy filtering of string array fields
func filterDevfileArrayFuzzy(index []indexSchema.Schema, requestedValues []string, options FilterOptions[[]string]) FilterResult {
	return FilterResult{
		filterFn: func(fr *FilterResult) {
			filteredIndex := deepcopy.Copy(index).([]indexSchema.Schema)

			if options.GetFromIndexField != nil || options.GetFromVersionField != nil {
				for i := 0; i < len(filteredIndex); i++ {
					if options.GetFromIndexField != nil {
						if indexFieldEmptyHandler(&filteredIndex, &i, requestedValues, options) {
							continue
						}

						filterIn := true
						valuesInIndex := getFuzzySetFromArray(options.GetFromIndexField(&filteredIndex[i]))

						for _, requestedValue := range requestedValues {
							if !fuzzyMatchInSet(valuesInIndex, requestedValue) {
								filterIn = false
								break
							}
						}

						if !filterIn {
							filterOut(&filteredIndex, &i)
							continue
						}
					}

					// go through each version's tags if multi-version stack is supported
					if !options.V1Index && options.GetFromVersionField != nil {
						for versionIndex := 0; versionIndex < len(filteredIndex[i].Versions); versionIndex++ {
							if versionFieldEmptyHandler(&filteredIndex[i].Versions, &versionIndex, requestedValues, options) {
								continue
							}

							filterVersion := true
							valuesInVersion := getFuzzySetFromArray(options.GetFromVersionField(&filteredIndex[i].Versions[versionIndex]))

							for _, requestedValue := range requestedValues {
								if !fuzzyMatchInSet(valuesInVersion, requestedValue) {
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
			}

			fr.Index = filteredIndex
		},
	}
}

// FilterDevfileSchemaVersion filters devfiles based on schema version
func FilterDevfileSchemaVersion(index []indexSchema.Schema, minSchemaVersion, maxSchemaVersion string) FilterResult {
	return FilterResult{
		filterFn: func(fr *FilterResult) {
			for i := 0; i < len(index); i++ {
				for versionIndex := 0; versionIndex < len(index[i].Versions); versionIndex++ {
					currectSchemaVersion := index[i].Versions[versionIndex].SchemaVersion
					schemaVersionWithoutServiceVersion := currectSchemaVersion[:strings.LastIndex(currectSchemaVersion, ".")]
					curVersion, err := versionpkg.NewVersion(schemaVersionWithoutServiceVersion)
					if err != nil {
						fr.Error = fmt.Errorf("failed to parse schemaVersion %s for stack: %s, version %s. Error: %v", currectSchemaVersion, index[i].Name, index[i].Versions[versionIndex].Version, err)
						return
					}

					versionInRange := true
					if minSchemaVersion != "" {
						minVersion, err := versionpkg.NewVersion(minSchemaVersion)
						if err != nil {
							fr.Error = fmt.Errorf("failed to parse minSchemaVersion %s. Error: %v", minSchemaVersion, err)
							return
						}
						if minVersion.GreaterThan(curVersion) {
							versionInRange = false
						}
					}
					if versionInRange && maxSchemaVersion != "" {
						maxVersion, err := versionpkg.NewVersion(maxSchemaVersion)
						if err != nil {
							fr.Error = fmt.Errorf("failed to parse maxSchemaVersion %s. Error: %v", maxSchemaVersion, err)
							return
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

			fr.Index = index
		},
	}
}

// FilterDevfileStrField filters by given string field, returns unchanged index if given parameter name is unrecognized
func FilterDevfileStrField(index []indexSchema.Schema, paramName, requestedValue string, v1Index bool) FilterResult {
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
			filterFn: func(fr *FilterResult) {
				fr.Index = index
			},
		}
	}

	return filterDevfileFieldFuzzy(index, requestedValue, options)
}

// AndFilter filters results of given filters to only overlapping results
func AndFilter(results ...FilterResult) FilterResult {
	return FilterResult{
		children: results,
		filterFn: func(fr *FilterResult) {
			andResultsMap := map[string]*struct {
				count  int
				schema indexSchema.Schema
			}{}
			andResults := []indexSchema.Schema{}

			for _, result := range fr.children {

				// Evaluates filter if not already evaluated
				if !result.IsEval {
					result.Eval()
				}

				// If a filter returns an error, return as overall result
				if result.Error != nil {
					fr.Error = result.Error
					return
				}

				for _, schema := range result.Index {
					andResult, found := andResultsMap[schema.Name]
					// if not found, initize is a seen counter of one and the current seen schema
					// else increment seen counter and re-assign current seen schema if versions have been filtered
					if !found {
						andResultsMap[schema.Name] = &struct {
							count  int
							schema indexSchema.Schema
						}{
							count:  1,
							schema: schema,
						}
					} else {
						andResultsMap[schema.Name].count += 1
						if len(schema.Versions) < len(andResult.schema.Version) {
							andResultsMap[schema.Name].schema = schema
						}
					}
				}
			}

			// build results of filters into new index schema
			for _, v := range andResultsMap {
				// if result is in every filter result then add to array
				if v.count == len(results) {
					andResults = append(andResults, v.schema)
				}
			}

			// set new index schema as the and filter result
			fr.Index = andResults
		},
	}
}

// FilterDevfileStrArrayField filters devfiles based on an array field
func FilterDevfileStrArrayField(index []indexSchema.Schema, paramName string, requestedValues []string, v1Index bool) FilterResult {
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

			for _, link := range s.Links {
				links = append(links, link)
			}

			return links
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			links := []string{}

			for _, link := range v.Links {
				links = append(links, link)
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
	case ARRAY_PARAM_GIT_REMOTES:
		options.GetFromIndexField = func(s *indexSchema.Schema) []string {
			gitRemotes := []string{}

			if s.Git != nil {
				for remoteName := range s.Git.Remotes {
					gitRemotes = append(gitRemotes, remoteName)
				}
			}

			return gitRemotes
		}
		options.GetFromVersionField = func(v *indexSchema.Version) []string {
			gitRemotes := []string{}

			if v.Git != nil {
				for remoteName := range v.Git.Remotes {
					gitRemotes = append(gitRemotes, remoteName)
				}
			}

			return gitRemotes
		}
	default:
		return FilterResult{
			filterFn: func(fr *FilterResult) {
				fr.Index = index
			},
		}
	}

	return filterDevfileArrayFuzzy(index, requestedValues, options)
}
