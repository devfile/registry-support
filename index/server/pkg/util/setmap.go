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

// A map type with keys and boolean values (to use as a set type)
type SetMap[T comparable] map[T]bool

// SetMapHas checks if a value v is in a SetMap
func (setMap SetMap[T]) Has(v T) bool {
	_, found := setMap[v]

	return found
}

func (setMap SetMap[T]) Union(other SetMap[T]) SetMap[T] {
	union := SetMap[T]{}

	for k, v := range setMap {
		union[k] = v
	}

	for k, v := range other {
		union[k] = v
	}

	return union
}

// ToArray converts the SetMap into an array
func (setMap SetMap[T]) ToArray() []T {
	arr := []T{}

	for v := range setMap {
		arr = append(arr, v)
	}

	return arr
}

// ArrayToSetMap converts an array into SetMap
func ArrayToSetMap[T comparable](arr []T) SetMap[T] {
	setMap := SetMap[T]{}

	for _, v := range arr {
		setMap[v] = true
	}

	return setMap
}

// StrArrayToSetMap converts a string array into SetMap
func StrArrayToSetMap(strArr []string) SetMap[string] {
	setMap := SetMap[string]{}

	for _, v := range strArr {
		setMap[v] = true
	}

	return setMap
}
