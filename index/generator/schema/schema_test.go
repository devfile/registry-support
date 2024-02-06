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
package schema

import (
	"strings"
	"testing"
)

func TestGetName(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch name from Schema",
			schema: &Schema{
				Name: "nodejs",
			},
			want: "nodejs",
		},
		{
			name: "Fetch blank name from Schema",
			schema: &Schema{
				Name: "",
			},
			want: "",
		},
		{
			name:   "Fetch name from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch name from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetName(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch display name from Schema",
			schema: &Schema{
				DisplayName: "Node.js",
			},
			want: "Node.js",
		},
		{
			name: "Fetch blank display name from Schema",
			schema: &Schema{
				DisplayName: "",
			},
			want: "",
		},
		{
			name:   "Fetch display name from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch display name from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetDisplayName(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetDescription(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch description from Schema type",
			schema: &Schema{
				Description: "The is a test.",
			},
			want: "The is a test.",
		},
		{
			name: "Fetch description from Version type",
			schema: &Version{
				Description: "The is a test.",
			},
			want: "The is a test.",
		},
		{
			name:    "Fetch description from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch description from non pointer",
			schema: Schema{
				Description: "The is a test.",
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch description from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetDescription(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetIcon(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch icon from Schema type",
			schema: &Schema{
				Icon: "test.png",
			},
			want: "test.png",
		},
		{
			name: "Fetch icon from Version type",
			schema: &Version{
				Icon: "test.png",
			},
			want: "test.png",
		},
		{
			name:    "Fetch icon from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch icon from non pointer",
			schema: Schema{
				Icon: "test.png",
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch icon from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetIcon(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetProjectType(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch project type from Schema",
			schema: &Schema{
				ProjectType: "Flask",
			},
			want: "Flask",
		},
		{
			name: "Fetch blank project type from Schema",
			schema: &Schema{
				ProjectType: "",
			},
			want: "",
		},
		{
			name:   "Fetch project type from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch project type from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetProjectType(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch language from Schema",
			schema: &Schema{
				Language: "Python",
			},
			want: "Python",
		},
		{
			name: "Fetch blank language from Schema",
			schema: &Schema{
				Language: "",
			},
			want: "",
		},
		{
			name:   "Fetch language from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch language from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetLanguage(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch version from Schema type",
			schema: &Schema{
				Version: "2.0.0",
			},
			want: "2.0.0",
		},
		{
			name: "Fetch version from Version type",
			schema: &Version{
				Version: "2.0.0",
			},
			want: "2.0.0",
		},
		{
			name:    "Fetch version from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch version from non pointer",
			schema: Schema{
				Version: "2.0.0",
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch version from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetVersion(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetSchemaVersion(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Version
		want    string
		wantErr bool
	}{
		{
			name: "Fetch schema version from Version schema",
			schema: &Version{
				SchemaVersion: "2.2.0",
			},
			want: "2.2.0",
		},
		{
			name: "Fetch blank schema version from Version schema",
			schema: &Version{
				SchemaVersion: "",
			},
			want: "",
		},
		{
			name:   "Fetch schema version from empty Version schema",
			schema: &Version{},
			want:   "",
		},
		{
			name:    "Fetch schema version from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetSchemaVersion(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestIsDefault(t *testing.T) {
	tests := []struct {
		name       string
		schema     *Version
		want       bool
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "Check if default from Version schema (true)",
			schema: &Version{
				Default: true,
			},
			want: true,
		},
		{
			name: "Check if default from Version schema (false)",
			schema: &Version{
				Default: false,
			},
			want: false,
		},
		{
			name:   "Check if default from empty Version schema",
			schema: &Version{},
			want:   false,
		},
		{
			name:       "Check if default from empty",
			wantErrStr: "invalid memory address or nil pointer dereference",
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := IsDefault(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.wantErrStr) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetGlobalMemoryLimit(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch global memory limit from Schema",
			schema: &Schema{
				GlobalMemoryLimit: "256Mi",
			},
			want: "256Mi",
		},
		{
			name: "Fetch blank global memory limit from Schema",
			schema: &Schema{
				GlobalMemoryLimit: "",
			},
			want: "",
		},
		{
			name:   "Fetch global memory limit from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch global memory limit from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetGlobalMemoryLimit(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetGitUrl(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch git url from Schema type",
			schema: &Schema{
				Git: &Git{
					Url: "git.server.org/repo.git",
				},
			},
			want: "git.server.org/repo.git",
		},
		{
			name: "Fetch git url from Version type",
			schema: &Version{
				Git: &Git{
					Url: "git.server.org/repo.git",
				},
			},
			want: "git.server.org/repo.git",
		},
		{
			name:   "Fetch git url from Schema type with git unset",
			schema: &Schema{},
			want:   "",
		},
		{
			name:   "Fetch git url from Version type with git unset",
			schema: &Version{},
			want:   "",
		},
		{
			name:    "Fetch git url from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch git url from non pointer",
			schema: Schema{
				Git: &Git{
					Url: "git.server.org/repo.git",
				},
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch git url from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetGitUrl(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetGitRemoteName(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch git remote name from Schema type",
			schema: &Schema{
				Git: &Git{
					RemoteName: "git-server",
				},
			},
			want: "git-server",
		},
		{
			name: "Fetch git remote name from Version type",
			schema: &Version{
				Git: &Git{
					RemoteName: "git-server",
				},
			},
			want: "git-server",
		},
		{
			name:   "Fetch git remote name from Schema type with git unset",
			schema: &Schema{},
			want:   "",
		},
		{
			name:   "Fetch git remote name from Version type with git unset",
			schema: &Version{},
			want:   "",
		},
		{
			name:    "Fetch git remote name from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch git remote name from non pointer",
			schema: Schema{
				Git: &Git{
					RemoteName: "git-server",
				},
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch git remote name from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetGitRemoteName(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetGitSubDir(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch git subdirectory from Schema type",
			schema: &Schema{
				Git: &Git{
					SubDir: "python-project",
				},
			},
			want: "python-project",
		},
		{
			name: "Fetch git subdirectory from Version type",
			schema: &Version{
				Git: &Git{
					SubDir: "python-project",
				},
			},
			want: "python-project",
		},
		{
			name:   "Fetch git subdirectory from Schema type with git unset",
			schema: &Schema{},
			want:   "",
		},
		{
			name:   "Fetch git subdirectory from Version type with git unset",
			schema: &Version{},
			want:   "",
		},
		{
			name:    "Fetch git subdirectory from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch git subdirectory from non pointer",
			schema: Schema{
				Git: &Git{
					SubDir: "python-project",
				},
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch git subdirectory from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetGitSubDir(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetGitRevision(t *testing.T) {
	tests := []struct {
		name    string
		schema  interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Fetch git revision from Schema type",
			schema: &Schema{
				Git: &Git{
					Revision: "v1",
				},
			},
			want: "v1",
		},
		{
			name: "Fetch git revision from Version type",
			schema: &Version{
				Git: &Git{
					Revision: "v1",
				},
			},
			want: "v1",
		},
		{
			name:   "Fetch git revision from Schema type with git unset",
			schema: &Schema{},
			want:   "",
		},
		{
			name:   "Fetch git revision from Version type with git unset",
			schema: &Version{},
			want:   "",
		},
		{
			name:    "Fetch git revision from incorrect type",
			schema:  &struct{}{},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name: "Fetch git revision from non pointer",
			schema: Schema{
				Git: &Git{
					Revision: "v1",
				},
			},
			want:    "incorrect type, expected: *Schema or *Version, got: ",
			wantErr: true,
		},
		{
			name:    "Fetch git revision from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetGitRevision(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch provider from Schema",
			schema: &Schema{
				Provider: "devfile.io",
			},
			want: "devfile.io",
		},
		{
			name: "Fetch blank provider from Schema",
			schema: &Schema{
				Provider: "",
			},
			want: "",
		},
		{
			name:   "Fetch provider from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch provider from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetProvider(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}

func TestGetSupportUrl(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		want    string
		wantErr bool
	}{
		{
			name: "Fetch support url from Schema",
			schema: &Schema{
				SupportUrl: "https://devfile.io/docs",
			},
			want: "https://devfile.io/docs",
		},
		{
			name: "Fetch blank support url from Schema",
			schema: &Schema{
				SupportUrl: "",
			},
			want: "",
		},
		{
			name:   "Fetch support url from empty Schema",
			schema: &Schema{},
			want:   "",
		},
		{
			name:    "Fetch support url from empty",
			want:    "invalid memory address or nil pointer dereference",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetSupportUrl(test.schema)
			if !test.wantErr && got != test.want {
				t.Errorf("Got: %v, Expected: %v", got, test.want)
			} else if test.wantErr && !strings.HasPrefix(err.Error(), test.want) {
				t.Errorf("Got: %v, Expected: %v", err.Error(), test.want)
			}
		})
	}
}
