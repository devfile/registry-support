// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package server

func (params *ServeDevfileIndexV2Params) toIndexParams() IndexParams {
	return IndexParams{
		Name:             params.Name,
		DisplayName:      params.DisplayName,
		Description:      params.Description,
		AttributeNames:   params.AttributeNames,
		Tags:             params.Tags,
		Icon:             params.Icon,
		IconUri:          params.IconUri,
		Arch:             params.Arch,
		ProjectType:      params.ProjectType,
		Language:         params.Language,
		MinVersion:       params.MinVersion,
		MaxVersion:       params.MaxVersion,
		MinSchemaVersion: params.MinSchemaVersion,
		MaxSchemaVersion: params.MaxSchemaVersion,
		Default:          params.Default,
		Resources:        params.Resources,
		StarterProjects:  params.StarterProjects,
		LinkNames:        params.LinkNames,
		Links:            params.Links,
		CommandGroups:    params.CommandGroups,
		GitRemoteNames:   params.GitRemoteNames,
		GitRemotes:       params.GitRemotes,
		GitUrl:           params.GitUrl,
		GitRemoteName:    params.GitRemoteName,
		GitSubDir:        params.GitSubDir,
		GitRevision:      params.GitRevision,
		Provider:         params.Provider,
		SupportUrl:       params.SupportUrl,
	}
}

func (params *ServeDevfileIndexV2WithTypeParams) toIndexParams() IndexParams {
	return IndexParams{
		Name:             params.Name,
		DisplayName:      params.DisplayName,
		Description:      params.Description,
		AttributeNames:   params.AttributeNames,
		Tags:             params.Tags,
		Icon:             params.Icon,
		IconUri:          params.IconUri,
		Arch:             params.Arch,
		ProjectType:      params.ProjectType,
		Language:         params.Language,
		MinVersion:       params.MinVersion,
		MaxVersion:       params.MaxVersion,
		MinSchemaVersion: params.MinSchemaVersion,
		MaxSchemaVersion: params.MaxSchemaVersion,
		Default:          params.Default,
		Resources:        params.Resources,
		StarterProjects:  params.StarterProjects,
		LinkNames:        params.LinkNames,
		Links:            params.Links,
		CommandGroups:    params.CommandGroups,
		GitRemoteNames:   params.GitRemoteNames,
		GitRemotes:       params.GitRemotes,
		GitUrl:           params.GitUrl,
		GitRemoteName:    params.GitRemoteName,
		GitSubDir:        params.GitSubDir,
		GitRevision:      params.GitRevision,
		Provider:         params.Provider,
		SupportUrl:       params.SupportUrl,
	}
}

func (params *ServeDevfileIndexV1Params) toIndexParams() IndexParams {
	return IndexParams{
		Name:             params.Name,
		DisplayName:      params.DisplayName,
		Description:      params.Description,
		AttributeNames:   params.AttributeNames,
		Tags:             params.Tags,
		Icon:             params.Icon,
		IconUri:          params.IconUri,
		Arch:             params.Arch,
		ProjectType:      params.ProjectType,
		Language:         params.Language,
		MinSchemaVersion: params.MinSchemaVersion,
		MaxSchemaVersion: params.MaxSchemaVersion,
		Resources:        params.Resources,
		StarterProjects:  params.StarterProjects,
		LinkNames:        params.LinkNames,
		Links:            params.Links,
		GitRemoteNames:   params.GitRemoteNames,
		GitRemotes:       params.GitRemotes,
		GitUrl:           params.GitUrl,
		GitRemoteName:    params.GitRemoteName,
		GitSubDir:        params.GitSubDir,
		GitRevision:      params.GitRevision,
		Provider:         params.Provider,
		SupportUrl:       params.SupportUrl,
	}
}

func (params *ServeDevfileIndexV1WithTypeParams) toIndexParams() IndexParams {
	return IndexParams{
		Name:             params.Name,
		DisplayName:      params.DisplayName,
		Description:      params.Description,
		AttributeNames:   params.AttributeNames,
		Tags:             params.Tags,
		Icon:             params.Icon,
		IconUri:          params.IconUri,
		Arch:             params.Arch,
		ProjectType:      params.ProjectType,
		Language:         params.Language,
		MinSchemaVersion: params.MinSchemaVersion,
		MaxSchemaVersion: params.MaxSchemaVersion,
		Resources:        params.Resources,
		StarterProjects:  params.StarterProjects,
		LinkNames:        params.LinkNames,
		Links:            params.Links,
		GitRemoteNames:   params.GitRemoteNames,
		GitRemotes:       params.GitRemotes,
		GitUrl:           params.GitUrl,
		GitRemoteName:    params.GitRemoteName,
		GitSubDir:        params.GitSubDir,
		GitRevision:      params.GitRevision,
		Provider:         params.Provider,
		SupportUrl:       params.SupportUrl,
	}
}
