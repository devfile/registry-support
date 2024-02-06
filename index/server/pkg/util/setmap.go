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

// A map type with any typed keys and boolean values (to use as a set type)
type SetMap map[any]bool

// SetMapHas checks if a value v is in a SetMap
func (setMap SetMap) SetMapHas(v any) bool {
	_, found := setMap[v]

	return found
}

// StrArrayToSetMap converts a string array into SetMap
func StrArrayToSetMap(strArr []string) SetMap {
	setMap := SetMap{}

	for _, v := range strArr {
		setMap[v] = true
	}

	return setMap
}
