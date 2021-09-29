package util

import (
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
)

// FilterDevfileArchitectures filters devfiles based on architectures
func FilterDevfileArchitectures(index []indexSchema.Schema, archs []string) []indexSchema.Schema {
	for i := 0; i < len(index); i++ {
		if len(index[i].Architectures) == 0 {
			// If a devfile has no architectures mentioned, then it supports all architectures
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
		}
	}

	return index
}
