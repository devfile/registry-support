package types

// Meta represents meta.yaml file
type Meta struct {
	Name              string   `yaml:"name,omitempty" json:"name,omitempty"`
	DisplayName       string   `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description       string   `yaml:"description,omitempty" json:"description,omitempty"`
	Tags              []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	GlobalMemoryLimit string   `yaml:"globalMemoryLimit,omitempty" json:"globalMemoryLimit,omitempty"`
	Icon              string   `yaml:"icon,omitempty" json:"icon,omitempty"`
	ProjectType       string   `yaml:"projectType,omitempty" json:"projectType,omitempty"`
	Language          string   `yaml:"language,omitempty" json:"language,omitempty"`
}

// MetaIndex is one item in index.json
// This is Meta extended with Links field
type MetaIndex struct {
	Meta
	Links Links `yaml:"links,omitempty" json:"links,omitempty"`
}
type Links struct {
	Self string `yaml:"self,omitempty" json:"self,omitempty"`
}
