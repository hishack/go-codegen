package field

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// pathParamRe 匹配路径参数及其可选类型标注::id 或 :id(int64)
var pathParamRe = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)(?:\([^)]*\))?`)

// ============ Type:一个字段的类型 ============

// Kind 区分类型的形态,决定后面怎么拼出 Go 类型、怎么递归解析
type Kind int

const (
	KindInvalid   Kind = iota
	KindPrimitive      // string / int64 / time.Time ...
	KindSlice          // []int
	KindMap            // map[string]int
	KindStruct         // 嵌套对象,已提升为命名结构体
)

func (k Kind) String() string {
	switch k {
	case KindPrimitive:
		return "primitive"
	case KindSlice:
		return "slice"
	case KindMap:
		return "map"
	case KindStruct:
		return "struct"
	}
	return "invalid"
}

// Type 是一个已解析的类型。嵌套用指针递归表达,
// 这样任意深度的嵌套都能表示,且不会无限展开
type Type struct {
	Kind Kind

	// KindPrimitive 时:基础类型名,如 "int64"、"string"、"time.Time"
	// KindStruct 时:结构体名,如 "GetCharacterSkillsResListItem"
	Name string

	// KindSlice 时:元素类型,如 []int 的 Elem 是 int
	Elem *Type
	// KindMap 时:key 类型,如 map[string]int 的 Key 是 string
	Key *Type

	// KindStruct 时:指向真正的结构体定义,便于生成时拿到全部字段
	Struct *Struct
}

// GoString 输出可直接写进 Go 源码的类型表达式
func (t *Type) GoString() string {
	if t == nil {
		return "any"
	}
	switch t.Kind {
	case KindPrimitive:
		if t.Name == "file" {
			return "*multipart.FileHeader"
		}
		return t.Name
	case KindSlice:
		return "[]" + t.Elem.GoString()
	case KindMap:
		return "map[" + t.Key.GoString() + "]" + t.Elem.GoString()
	case KindStruct:
		return "*" + t.Struct.Name
	}
	return "any"
}

// IsFile 是否是文件上传类型(yaml 里写 file / [file])
func (t *Type) IsFile() bool {
	return t != nil && t.Kind == KindPrimitive && t.Name == "file"
}

// IsPtr 判断是否是指针类型(嵌套结构体统一用指针,避免递归结构体无限增长)
func (t *Type) IsPtr() bool {
	return t != nil && t.Kind == KindStruct
}

// ============ Field:一个字段 ============

// Field 是结构体里的一个字段
type Field struct {
	Name      string // Go 字段名,大驼峰,如 Name / UserId
	Tag       string // yaml/json 里的原始名字,小驼峰,如 name / user_id
	Type      *Type
	Comment   string // 字段注释,可从 yaml 的注释里提取
	OmitEmpty bool   // 是否带 omitempty
}

// ============ Struct:一个命名结构体 ============

// Struct 是一个可以生成代码的命名结构体。
// 顶层是 {ApiName}Req / {ApiName}Res,嵌套的会被提升为 {前缀}{字段名}
type Struct struct {
	Name    string
	Fields  []*Field
	Comment string
}

// GoString 输出结构体定义
func (s *Struct) GoString() string {
	var b strings.Builder
	if s.Comment != "" {
		b.WriteString("// " + s.Comment + "\n")
	}
	b.WriteString("type " + s.Name + " struct {\n")
	for _, f := range s.Fields {
		if f.Comment != "" {
			b.WriteString("\t// " + f.Comment + "\n")
		}
		tag := f.Tag
		if f.OmitEmpty {
			tag += ",omitempty"
		}
		fmt.Fprintf(&b, "\t%s %s `json:%q`\n", f.Name, f.Type.GoString(), tag)
	}
	b.WriteString("}")
	return b.String()
}

// Field 按名字查找
func (s *Struct) Field(name string) *Field {
	for _, f := range s.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// ============ Api:一个接口 ============

// Api 是一个接口的完整描述,对应 yaml 里的一个 key
type Api struct {
	ApiName string // 接口名,如 GetCharacterSkills
	Method  string // 已转成大写,如 GET
	Path    string // 如 /characters/:id(int64)/skills
	Comment string // 接口注释,取自 api key 上方的注释,没有则为空

	// Params 是从 path 中解析出来的参数,如 /users/:id(int64) -> [{Id int64}]
	Params []*Field

	// Req / Res 是请求/响应体,已由嵌套结构提升为命名结构体。
	// yaml 中允许没有 req 或没有 res,所以可能为 nil
	Req *Struct
	Res *Struct

	// 归属信息,用于决定生成到哪个文件/包
	FileName string
	PkgName  string
}

// HasReq / HasRes 便于模板里做条件判断
func (a *Api) HasReq() bool { return a.Req != nil && len(a.Req.Fields) > 0 }
func (a *Api) HasRes() bool { return a.Res != nil && len(a.Res.Fields) > 0 }

// ScalarRes 当 res 是"仅有一个基础类型字段"时,返回该字段的类型;
// 这种接口 service 直接返回基础类型(如 (int64, error)),不生成 Vo 结构体。
// 其余情况返回 nil。
func (a *Api) ScalarRes() *Type {
	if !a.HasRes() || len(a.Res.Fields) != 1 {
		return nil
	}
	f := a.Res.Fields[0]
	if f.Type == nil || f.Type.Kind != KindPrimitive {
		return nil
	}
	return f.Type
}

// RoutePath 去掉路径参数的类型标注,得到可直接注册的路由:
// /users/:id(int64) -> /users/:id
func (a *Api) RoutePath() string {
	if a.Path == "" {
		return ""
	}
	out := pathParamRe.ReplaceAllString(a.Path, ":$1")
	return out
}

// ============ Yaml:一个 yaml 文件 ============

// Yaml 对应 docs 目录下的一个 yaml 文件
type Yaml struct {
	FileName    string // 文件名,如 user
	PackageName string // 包名,如 system
	ApiList     []*Api

	// Reg 收集本文件里定义的所有结构体(含嵌套提升出来的),
	// 生成 model 时直接遍历它,保证一个类型只生成一次
	Reg *Registry
}

// YamlList 对应 yaml 目录下的所有文件
type YamlList struct {
	YamlList []*Yaml
}

// AllStructs 汇总所有文件的所有结构体,按文件顺序、登记顺序排列
func (l *YamlList) AllStructs() []*Struct {
	var out []*Struct
	for _, y := range l.YamlList {
		if y == nil || y.Reg == nil {
			continue
		}
		out = append(out, y.Reg.All()...)
	}
	return out
}

// AllApis 汇总所有接口
func (l *YamlList) AllApis() []*Api {
	var out []*Api
	for _, y := range l.YamlList {
		if y == nil {
			continue
		}
		out = append(out, y.ApiList...)
	}
	return out
}

// ============ Registry:结构体登记表 ============

// Registry 负责给匿名嵌套结构体分配唯一名字并去重。
// 名字用前缀拼接(ApiName + 字段名)保证不同 api 之间不会撞车,
// 同名重复登记时直接复用,天然幂等
type Registry struct {
	structs map[string]*Struct // name -> 定义
	order   []*Struct          // 保持登记顺序,让生成结果稳定
}

func NewRegistry() *Registry {
	return &Registry{structs: make(map[string]*Struct)}
}

// Register 登记一个结构体;已存在则返回旧的那个(幂等)
func (r *Registry) Register(s *Struct) *Struct {
	if r.structs == nil {
		r.structs = make(map[string]*Struct)
	}
	if old, ok := r.structs[s.Name]; ok {
		return old
	}
	r.structs[s.Name] = s
	r.order = append(r.order, s)
	return s
}

// Has 判断是否已登记
func (r *Registry) Has(name string) bool {
	_, ok := r.structs[name]
	return ok
}

// Get 按名字取
func (r *Registry) Get(name string) *Struct {
	return r.structs[name]
}

// All 按登记顺序返回全部
func (r *Registry) All() []*Struct {
	return r.order
}

// ============ 辅助函数 ============

// UpperFirst 小驼峰 -> 大驼峰
func UpperFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// LowerFirst 大驼峰 -> 小驼峰
func LowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// SortedKeys 返回排序后的 key。
// 必须排序!直接遍历 map 顺序是随机的,
// 会导致每次生成的代码字段顺序不同,破坏幂等
func SortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// primitives 是允许直接使用的 go 基础类型白名单。
// file 是特殊类型:映射为 *multipart.FileHeader,只能用在 req 字段
var primitives = map[string]bool{
	"bool": true, "string": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
	"byte": true, "rune": true,
	"time.Time": true, "time.Duration": true,
	"any": true, "interface{}": true,
	"file": true,
}

// IsPrimitive 判断是否是合法的基础类型
func IsPrimitive(name string) bool {
	return primitives[name]
}
