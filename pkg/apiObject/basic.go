package apiObject

// 我们常见的API对象的yaml文件中，都有apiVersion、kind、metadata三个字段
// Basic包含的是除了Spec的所有字段
type Metadata struct {
	UUID       string            `json:"uuid" yaml:"uuid"`
	Name       string            `json:"name" yaml:"name"`
	Namespace  string            `yaml:"namespace"`
	Label      map[string]string `yaml:"label"`
	Annotation map[string]string `yaml:"annotation"`
}

type Basic struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
}
