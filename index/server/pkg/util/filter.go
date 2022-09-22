//
// Copyright 2022 Red Hat, Inc.
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
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	versionpkg "github.com/hashicorp/go-version"
	"strings"
)

// FilterDevfileArchitectures filters devfiles based on architectures
func FilterDevfileArchitectures(index []indexSchema.Schema, archs []string, v1Index bool) []indexSchema.Schema {
	for i := 0; i < len(index); i++ {
		if len(index[i].Architectures) == 0 {
			// If a stack has no architectures mentioned, then it supports all architectures
			continue
		}

		filterIn := true

		for _, requestedArch := range archs {
			isArchPresent := false
			for _, devfileArch := range index[i].Architectures {
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
			// if an arch requested is not present in a devfile, filter it out
			index = append(index[:i], index[i+1:]...)

			// decrement counter, since we shifted the array
			i--
			continue
		}

		// go through each version's architecture if multi-version stack is supported
		if !v1Index {
			for versionIndex := 0; versionIndex < len(index[i].Versions); versionIndex++ {
				versionArchs := index[i].Versions[versionIndex].Architectures
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
					// if an arch requested is not present in a devfile, filter it out
					index[i].Versions = append(index[i].Versions[:versionIndex], index[i].Versions[versionIndex+1:]...)

					// decrement counter, since we shifted the array
					versionIndex--
				}

			}
		}
	}

	return index
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
