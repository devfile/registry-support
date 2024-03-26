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

package server

import (
	"fmt"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/index/server/pkg/util"
)

func filterFieldbyParam(index []indexSchema.Schema, wantV1Index bool, paramName string, paramValue any) util.FilterResult {
	switch typedValue := paramValue.(type) {
	case string:
		return util.FilterDevfileStrField(index, paramName, typedValue, wantV1Index)
	default:
		return util.FilterDevfileStrField(index, paramName, fmt.Sprintf("%v", typedValue), wantV1Index)
	}
}

func filterFieldsByParams(index []indexSchema.Schema, wantV1Index bool, params IndexParams) ([]indexSchema.Schema, error) {
	paramsMap := util.StructToMap(params)
	results := []util.FilterResult{}
	var andResult util.FilterResult

	if len(paramsMap) == 0 {
		return index, nil
	}

	for paramName, paramValue := range paramsMap {
		if util.IsFieldParameter(paramName) {
			result := filterFieldbyParam(index, wantV1Index, paramName, paramValue)
			results = append(results, result)
		} else if util.IsArrayParameter(paramName) {
			typedValues := paramValue.([]string)
			result := util.FilterDevfileStrArrayField(index, paramName, typedValues, wantV1Index)
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return index, nil
	}

	andResult = util.AndFilter(results...)

	return andResult.Index, andResult.Error
}
