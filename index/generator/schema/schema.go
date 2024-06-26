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

package schema

import (
	"time"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

/*
Sample index file:
[
  {
    "name": "java-maven",
    "displayName": "Maven Java",
    "description": "Upstream Maven and OpenJDK 11",
    "type": "stack",
    "tags": [
      "Java",
      "Maven"
    ],
    "architectures": [
      "amd64",
      "arm64",
      "s390x"
    ],
    "projectType": "maven",
    "language": "java",
	"lastModified": "2024-05-13T12:32:02+02:00",
    "versions": [
      {
        "version": "1.1.0",
        "schemaVersion": "2.1.0",
        "default": true,
        "description": "Upstream Maven and OpenJDK 11",
        "tags": [
          "Java",
          "Maven"
        ],
        "architectures": [
          "amd64",
          "arm64",
          "s390x"
        ],
        "links": {
          "self": "devfile-catalog/java-maven:1.1.0"
        },
        "resources": [
          "devfile.yaml"
        ],
        "starterProjects": [
          "springbootproject"
        ],
		"lastModified": "2024-05-13T12:32:02+02:00"
      }
    ]
  },
  {
    "name": "java-quarkus",
    "displayName": "Quarkus Java",
    "description": "Quarkus with Java",
    "type": "stack",
    "tags": [
      "Java",
      "Quarkus"
    ],
    "architectures": [
      "amd64"
    ],
    "projectType": "quarkus",
    "language": "java",
	"lastModified": "2024-04-29T17:08:43+03:00",
    "versions": [
      {
        "version": "1.1.0",
        "schemaVersion": "2.0.0",
        "default": true,
        "description": "Quarkus with Java",
        "tags": [
          "Java",
          "Quarkus"
        ],
        "architectures": [
          "amd64"
        ],
        "links": {
          "self": "devfile-catalog/java-quarkus:1.1.0"
        },
        "resources": [
          "devfile.yaml"
        ],
		"commandGroups": {
		  "build": true,
		  "run": true,
		  "test": false,
		  "debug": false,
		  "deploy": false
		},
        "starterProjects": [
          "community",
          "redhat-product"
        ],
		"lastModified": "2024-04-29T17:08:43+03:00"
      }
    ]
  }
]
*/

/*
Index file schema definition
name: string - The stack name
version: string - The stack version
attributes: map[string]apiext.JSON - Map of implementation-dependant free-form YAML attributes
displayName: string - The display name of devfile
description: string - The description of devfile
type: DevfileType - The type of the devfile, currently supports stack and sample
tags: string[] - The tags associated to devfile
icon: string - The devfile icon
globalMemoryLimit: string - The devfile global memory limit
projectType: string - The project framework that is used in the devfile
language: string - The project language that is used in the devfile
links: map[string]string - Links related to the devfile
commandGroups: map[CommandGroupKind]bool - The command groups that are used in the devfile
deploymentScopes: map[DeploymentScopeKind]bool - The deployment scope that are detected in the devfile
resources: []string - The file resources that compose a devfile stack.
starterProjects: string[] - The project templates that can be used in the devfile
git: *git - The information of remote repositories
provider: string - The devfile provider information
versions: []Version - The list of stack versions information
lastModified: string - The date that a version of this stack/sample was last changed
*/

// Schema is the index file schema
type Schema struct {
	Name              string                       `yaml:"name,omitempty" json:"name,omitempty"`
	Version           string                       `yaml:"version,omitempty" json:"version,omitempty"`
	Attributes        map[string]apiext.JSON       `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	DisplayName       string                       `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description       string                       `yaml:"description,omitempty" json:"description,omitempty"`
	Type              DevfileType                  `yaml:"type,omitempty" json:"type,omitempty"`
	Tags              []string                     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Architectures     []string                     `yaml:"architectures,omitempty" json:"architectures,omitempty"`
	Icon              string                       `yaml:"icon,omitempty" json:"icon,omitempty"`
	GlobalMemoryLimit string                       `yaml:"globalMemoryLimit,omitempty" json:"globalMemoryLimit,omitempty"`
	ProjectType       string                       `yaml:"projectType,omitempty" json:"projectType,omitempty"`
	Language          string                       `yaml:"language,omitempty" json:"language,omitempty"`
	Links             map[string]string            `yaml:"links,omitempty" json:"links,omitempty"`
	CommandGroups     map[CommandGroupKind]bool    `yaml:"commandGroups,omitempty" json:"commandGroups,omitempty"`
	DeploymentScopes  map[DeploymentScopeKind]bool `yaml:"deploymentScopes,omitempty" json:"deploymentScopes,omitempty"`
	Resources         []string                     `yaml:"resources,omitempty" json:"resources,omitempty"`
	StarterProjects   []string                     `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
	Git               *Git                         `yaml:"git,omitempty" json:"git,omitempty"`
	Provider          string                       `yaml:"provider,omitempty" json:"provider,omitempty"`
	SupportUrl        string                       `yaml:"supportUrl,omitempty" json:"supportUrl,omitempty"`
	Versions          []Version                    `yaml:"versions,omitempty" json:"versions,omitempty"`
	LastModified      string                       `yaml:"lastModified,omitempty" json:"lastModified,omitempty"`
}

// DevfileType describes the type of devfile
type DevfileType string

const (
	// SampleDevfileType represents a sample devfile
	SampleDevfileType DevfileType = "sample"

	// StackDevfileType represents a stack devfile
	StackDevfileType DevfileType = "stack"
)

// CommandGroupKind describes the kind of command group
type CommandGroupKind string

const (
	BuildCommandGroupKind  CommandGroupKind = "build"
	RunCommandGroupKind    CommandGroupKind = "run"
	TestCommandGroupKind   CommandGroupKind = "test"
	DebugCommandGroupKind  CommandGroupKind = "debug"
	DeployCommandGroupKind CommandGroupKind = "deploy"
)

// DeploymentScopeKind describes the kind of deployment scope
type DeploymentScopeKind string

const (
	InnerloopKind DeploymentScopeKind = "innerloop"
	OuterloopKind DeploymentScopeKind = "outerloop"
)

// StarterProject is the devfile starter project
type StarterProject struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Commands stores the command information
type Commands struct {
	Id        string      `yaml:"id,omitempty" json:"id,omitempty"`
	Exec      CommandType `yaml:"exec,omitempty" json:"exec,omitempty"`
	Apply     CommandType `yaml:"apply,omitempty" json:"apply,omitempty"`
	Composite CommandType `yaml:"composite,omitempty" json:"composite,omitempty"`
}

// CommandType stores the group for a command
type CommandType struct {
	Group CommandGroup `yaml:"group,omitempty" json:"group,omitempty"`
}

// CommandGroup stores the group information for a command
type CommandGroup struct {
	Kind      CommandGroupKind `yaml:"kind,omitempty" json:"kind,omitempty"`
	IsDefault bool             `yaml:"isDefault,omitempty" json:"isDefault,omitempty"`
}

// Devfile is the devfile structure that is used by index component
type Devfile struct {
	Meta            Schema           `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	StarterProjects []StarterProject `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
	Commands        []Commands       `yaml:"commands,omitempty" json:"commands,omitempty"`
	SchemaVersion   string           `yaml:"schemaVersion,omitempty" json:"schemaVersion,omitempty"`
}

// Git stores the information of remote repositories
type Git struct {
	Remotes    map[string]string `yaml:"remotes,omitempty" json:"remotes,omitempty"`
	Url        string            `yaml:"url,omitempty" json:"url,omitempty"`
	RemoteName string            `yaml:"remoteName,omitempty" json:"remoteName,omitempty"`
	SubDir     string            `yaml:"subDir,omitempty" json:"subDir,omitempty"`
	Revision   string            `yaml:"revision,omitempty" json:"revision,omitempty"`
}

// ExtraDevfileEntries is the extraDevfileEntries structure that is used by index component
type ExtraDevfileEntries struct {
	Samples []Schema `yaml:"samples,omitempty" json:"samples,omitempty"`
	Stacks  []Schema `yaml:"stacks,omitempty" json:"stacks,omitempty"`
}

// StackInfo stores the top-level stack information defined within stack.yaml
type StackInfo struct {
	Name        string    `yaml:"name,omitempty" json:"name,omitempty"`
	DisplayName string    `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	Icon        string    `yaml:"icon,omitempty" json:"icon,omitempty"`
	Versions    []Version `yaml:"versions,omitempty" json:"versions,omitempty"`
}

// Version stores the information for each stack version
type Version struct {
	Version          string                       `yaml:"version,omitempty" json:"version,omitempty"`
	SchemaVersion    string                       `yaml:"schemaVersion,omitempty" json:"schemaVersion,omitempty"`
	Default          bool                         `yaml:"default,omitempty" json:"default,omitempty"`
	Git              *Git                         `yaml:"git,omitempty" json:"git,omitempty"`
	Description      string                       `yaml:"description,omitempty" json:"description,omitempty"`
	Tags             []string                     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Architectures    []string                     `yaml:"architectures,omitempty" json:"architectures,omitempty"`
	Icon             string                       `yaml:"icon,omitempty" json:"icon,omitempty"`
	Links            map[string]string            `yaml:"links,omitempty" json:"links,omitempty"`
	CommandGroups    map[CommandGroupKind]bool    `yaml:"commandGroups,omitempty" json:"commandGroups,omitempty"`
	DeploymentScopes map[DeploymentScopeKind]bool `yaml:"deploymentScopes,omitempty" json:"deploymentScopes,omitempty"`
	Resources        []string                     `yaml:"resources,omitempty" json:"resources,omitempty"`
	StarterProjects  []string                     `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
	LastModified     string                       `yaml:"lastModified,omitempty" json:"lastModified,omitempty"`
}

type LastModifiedEntry struct {
	Name         string    `yaml:"name,omitempty" json:"name,omitempty"`
	Version      string    `yaml:"version,omitempty" json:"version,omitempty"`
	LastModified time.Time `yaml:"lastModified,omitempty" json:"lastModified,omitempty"`
}

type LastModifiedInfo struct {
	Stacks  []LastModifiedEntry `yaml:"stacks,omitempty" json:"stacks,omitempty"`
	Samples []LastModifiedEntry `yaml:"samples,omitempty" json:"samples,omitempty"`
}
