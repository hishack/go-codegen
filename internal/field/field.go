package field

// 每一个api的小单元
// GetCharacterSkills:
// method: GET
// path: /characters/:id(int64)/skills
// res:
// list:
// - id: int64
// name: string
// level: []int
type Api struct {
	ApiName string `yaml:"apiName"`
	Method  string `yaml:"method"`
	Path    string `yaml:"path"`
	Req     any    `yaml:"req"` // 可能是嵌套类型,如何设计?
	Res     any    `yaml:"res"` // 可能是嵌套类型,如何设计?
}

// 一页的api

type Yaml struct {
	PackageName string `yaml:"packageName"`
	FileName    string `yaml:"fileName"`
	ApiList     []Api  `yaml:"apiList"`
}

// 全部的api

type YamlList struct {
	YamlList []Yaml `yaml:"yamlList"`
}
