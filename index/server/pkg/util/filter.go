package util

import (
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
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
					for _, versionArch := range versionArchs{
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
					continue
				}

			}
		}
	}

	return index
}
