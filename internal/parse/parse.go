package parse

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hishack/go-codegen/internal/field"
	"gopkg.in/yaml.v3"
)

// YamlParse 负责把 yaml 目录解析成 field.YamlList。
// 解析基于 yaml.Node(而不是 map),这样才能拿到注释:
//   - api key 上方的注释            -> Api.Comment
//   - req/res 字段的行注释或上方注释 -> Field.Comment
type YamlParse struct {
	YamlDir string
}

func NewYamlParse(yamlDir string) (*YamlParse, error) {
	info, err := os.Stat(yamlDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("yaml 目录不存在: %s", yamlDir)
	}
	return &YamlParse{YamlDir: yamlDir}, nil
}

// ParseAllYaml 解析目录下所有 .yaml/.yml 文件。
// prefix 是全局的 api 结构体名前缀(来自配置,可为空),
// 只作用于 api 包的请求结构体,如 prefix=YiKou 时生成 YiKouCreateUserRequest。
func (y *YamlParse) ParseAllYaml(prefix string) (*field.YamlList, error) {
	files, err := y.listYamlFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("yaml 目录下没有 .yaml/.yml 文件: %s", y.YamlDir)
	}

	list := &field.YamlList{}
	for _, f := range files {
		yy, err := ParseYamlFile(f, prefix)
		if err != nil {
			return nil, err
		}
		list.YamlList = append(list.YamlList, yy)
	}
	if err := validateGlobal(list.YamlList); err != nil {
		return nil, err
	}
	return list, nil
}

// listYamlFiles 收集目录下的 yaml 文件,按文件名排序保证处理顺序稳定
func (y *YamlParse) listYamlFiles() ([]string, error) {
	entries, err := os.ReadDir(y.YamlDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, filepath.Join(y.YamlDir, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

// ============ 单文件解析 ============

// fileParser 单个文件的解析上下文
type fileParser struct {
	reg    *field.Registry
	prefix string
}

// ParseYamlFile 解析一个 yaml 文件。
// 文件头两行是 filename / package,其余顶层 key 都是一个接口定义。
func ParseYamlFile(path string, prefix string) (*field.Yaml, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// 去掉 UTF-8 BOM,否则 yaml.Unmarshal 会把 BOM 当成内容
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%s: yaml 语法错误: %w", filepath.Base(path), err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: 文件内容为空", filepath.Base(path))
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: 文件顶层必须是 key-value 形式", filepath.Base(path))
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	yy := &field.Yaml{
		FileName:    base,
		Reg:         field.NewRegistry(),
		PackageName: base, // package 字段缺失时的兜底
	}
	p := &fileParser{reg: yy.Reg, prefix: prefix}

	for i := 0; i+1 < len(root.Content); i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		switch k.Value {
		case "filename":
			if err := validateLowerCamel(v.Value, "filename"); err != nil {
				return nil, fmt.Errorf("%s: %w", base, err)
			}
			yy.FileName = v.Value
		case "package":
			if strings.TrimSpace(v.Value) == "" {
				return nil, fmt.Errorf("%s: package 不能为空", base)
			}
			yy.PackageName = v.Value
		default:
			api, err := p.parseApi(k, v)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", base, err)
			}
			yy.ApiList = append(yy.ApiList, api)
		}
	}

	if len(yy.ApiList) == 0 {
		return nil, fmt.Errorf("%s: 没有定义任何接口", base)
	}
	return yy, nil
}

// parseApi 解析一个接口定义。k 是接口名节点(可带注释),v 是接口体
func (p *fileParser) parseApi(k, v *yaml.Node) (*field.Api, error) {
	apiName := k.Value
	if err := validateUpperCamel(apiName, "接口名"); err != nil {
		return nil, fmt.Errorf("%s: %w", apiName, err)
	}
	if v.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: 接口定义必须是对象(method/path/req/res)", apiName)
	}

	api := &field.Api{
		ApiName: apiName,
		Comment: firstComment(k.HeadComment, k.LineComment),
	}

	var reqNode, resNode *yaml.Node
	var reqKey, resKey *yaml.Node
	for i := 0; i+1 < len(v.Content); i += 2 {
		key, val := v.Content[i], v.Content[i+1]
		switch key.Value {
		case "method":
			m := strings.ToLower(strings.TrimSpace(val.Value))
			if !methodSet[m] {
				return nil, fmt.Errorf("%s.method 非法值 %q(允许:get/post/put/patch/delete,不区分大小写)", apiName, val.Value)
			}
			api.Method = strings.ToUpper(m)
		case "path":
			if err := validatePath(val.Value); err != nil {
				return nil, fmt.Errorf("%s.path %w", apiName, err)
			}
			api.Path = val.Value
		case "filename":
			// 允许接口级 filename 覆盖,该接口会生成到独立的一组文件里
			if err := validateLowerCamel(val.Value, "filename"); err != nil {
				return nil, fmt.Errorf("%s.filename: %w", apiName, err)
			}
			api.FileName = val.Value
		case "req":
			reqKey, reqNode = key, val
		case "res":
			resKey, resNode = key, val
		default:
			return nil, fmt.Errorf("%s: 不支持的字段 %q(仅允许 method/path/req/res/filename)", apiName, key.Value)
		}
	}

	if api.Method == "" {
		return nil, fmt.Errorf("%s: 缺少 method", apiName)
	}
	if api.Path == "" {
		return nil, fmt.Errorf("%s: 缺少 path", apiName)
	}

	params, err := parsePathParams(api.Path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", apiName, err)
	}
	api.Params = params

	// req/res 是普通对象,空对象等价于没写
	if api.Req, err = p.parseObject(reqNode, p.reqStructName(apiName), structComment(reqKey, reqNode)); err != nil {
		return nil, fmt.Errorf("%s.req: %w", apiName, err)
	}
	if api.Res, err = p.parseObject(resNode, p.resStructName(apiName), structComment(resKey, resNode)); err != nil {
		return nil, fmt.Errorf("%s.res: %w", apiName, err)
	}
	return api, nil
}

// reqStructName 请求结构体名:前缀 + 接口名 + Request,如 YiKouCreateUserRequest
func (p *fileParser) reqStructName(apiName string) string {
	return p.prefix + apiName + "Request"
}

// resStructName 响应数据结构名(vo):接口名 + Vo;接口名已带 Vo 后缀时不再重复
func (p *fileParser) resStructName(apiName string) string {
	if strings.HasSuffix(apiName, "Vo") {
		return apiName
	}
	return apiName + "Vo"
}

// structComment 取 req:/res: 的行注释或上方注释,作为结构体注释
func structComment(key, node *yaml.Node) string {
	if key == nil && node == nil {
		return ""
	}
	var cands []string
	if node != nil {
		cands = append(cands, node.LineComment)
	}
	if key != nil {
		cands = append(cands, key.LineComment, key.HeadComment)
	}
	return firstComment(cands...)
}

// parseObject 把 req/res 的 mapping 解析成命名结构体,嵌套对象递归提升。
// node 为 nil 或空对象时返回 nil。
func (p *fileParser) parseObject(node *yaml.Node, name, comment string) (*field.Struct, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("必须是对象(key: 类型),实际是 %s", kindName(node.Kind))
	}
	if len(node.Content) == 0 {
		return nil, nil
	}
	return p.parseStruct(name, node, comment)
}

// parseStruct 解析一个 mapping 节点为结构体并登记。
// 同名且内容一致直接复用(幂等);同名但内容不同是命名冲突,必须报错,
// 否则会静默生成错误的结构体(如 GetUser 的 res 名与名为 GetUserVo 的接口撞车)
func (p *fileParser) parseStruct(name string, node *yaml.Node, comment string) (*field.Struct, error) {
	s := &field.Struct{Name: name, Comment: comment}
	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		f, err := p.parseField(k, v, name)
		if err != nil {
			return nil, fmt.Errorf("字段 %q: %w", k.Value, err)
		}
		s.Fields = append(s.Fields, f)
	}

	// yaml 是树不是图,不会自引用;同名要么是同一份定义(幂等复用),
	// 要么是不同接口撞名(结构不同则报错)
	if old := p.reg.Get(name); old != nil {
		if !structEqual(old, s) {
			return nil, fmt.Errorf(
				"结构体名 %q 冲突:两个接口的请求/响应结构体同名但结构不同,请修改其中一个接口名", name)
		}
		return old, nil
	}
	p.reg.Register(s)
	return s, nil
}

// structEqual 比较两个结构体的字段名与类型是否一致(注释不参与比较)
func structEqual(a, b *field.Struct) bool {
	if len(a.Fields) != len(b.Fields) {
		return false
	}
	for i := range a.Fields {
		fa, fb := a.Fields[i], b.Fields[i]
		if fa.Name != fb.Name || fa.Type.GoString() != fb.Type.GoString() {
			return false
		}
	}
	return true
}

// parseField 解析单个字段:key 是字段名,value 是类型描述
func (p *fileParser) parseField(k, v *yaml.Node, structName string) (*field.Field, error) {
	if err := validateLowerCamel(k.Value, "字段名"); err != nil {
		return nil, err
	}
	t, err := p.parseType(v, structName+field.UpperFirst(k.Value))
	if err != nil {
		return nil, err
	}
	return &field.Field{
		Name:    field.UpperFirst(k.Value),
		Tag:     k.Value,
		Type:    t,
		Comment: mergeComment(cleanComment(k.HeadComment), cleanComment(v.LineComment)),
	}, nil
}

// parseType 解析类型节点,支持三种形态:
//  1. 标量        string / int64 / map[string]string
//  2. 单元素数组  [int] 或 - id: int64(对象数组)
//  3. 嵌套对象    mapping,提升为命名结构体
func (p *fileParser) parseType(node *yaml.Node, nestedName string) (*field.Type, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		v := strings.TrimSpace(node.Value)
		if strings.HasPrefix(v, "map[") {
			return parseMapType(v)
		}
		if !field.IsPrimitive(v) {
			return nil, fmt.Errorf("不支持的字段类型 %q(可用:%s 或 [类型] 或嵌套对象)", v, primitiveDoc)
		}
		return &field.Type{Kind: field.KindPrimitive, Name: v}, nil

	case yaml.SequenceNode:
		if len(node.Content) != 1 {
			return nil, fmt.Errorf("数组写法只支持单元素,如 [int] 或一个对象定义")
		}
		item := node.Content[0]
		switch item.Kind {
		case yaml.ScalarNode:
			elem, err := p.parseType(item, nestedName+"Item")
			if err != nil {
				return nil, fmt.Errorf("数组元素: %w", err)
			}
			if elem.Kind != field.KindPrimitive {
				return nil, fmt.Errorf("数组元素类型不支持 %q", item.Value)
			}
			return &field.Type{Kind: field.KindSlice, Elem: elem}, nil
		case yaml.MappingNode:
			// 对象数组:list: - id: int64,元素结构体名追加 Item
			sub, err := p.parseStruct(nestedName+"Item", item, firstComment(item.LineComment, node.HeadComment))
			if err != nil {
				return nil, err
			}
			return &field.Type{Kind: field.KindSlice, Elem: &field.Type{Kind: field.KindStruct, Struct: sub}}, nil
		default:
			return nil, fmt.Errorf("数组元素必须是类型或对象定义")
		}

	case yaml.MappingNode:
		sub, err := p.parseStruct(nestedName, node, firstComment(node.LineComment, node.HeadComment))
		if err != nil {
			return nil, err
		}
		return &field.Type{Kind: field.KindStruct, Struct: sub}, nil

	default:
		return nil, fmt.Errorf("不支持的结构 %s", kindName(node.Kind))
	}
}

// parseMapType 解析 map[key]value,value 必须是基础类型
func parseMapType(v string) (*field.Type, error) {
	m := mapTypeRe.FindStringSubmatch(v)
	if m == nil {
		return nil, fmt.Errorf("map 类型格式非法 %q,应为 map[key]value,如 map[string]string", v)
	}
	keyType, valType := m[1], m[2]
	if !field.IsPrimitive(keyType) {
		return nil, fmt.Errorf("map 的 key 不支持 %q(必须是基础类型)", keyType)
	}
	if !field.IsPrimitive(valType) {
		return nil, fmt.Errorf("map 的 value 不支持 %q(必须是基础类型)", valType)
	}
	return &field.Type{
		Kind: field.KindMap,
		Key:  &field.Type{Kind: field.KindPrimitive, Name: keyType},
		Elem: &field.Type{Kind: field.KindPrimitive, Name: valType},
	}, nil
}

// ============ 路径参数 ============

// parsePathParams 从 path 中解析参数:/users/:userId(int64) -> [{Name:"UserId", Tag:"userId", Type:int64}]
// 没写类型标注的参数默认 string
func parsePathParams(path string) ([]*field.Field, error) {
	var out []*field.Field
	seen := map[string]bool{}
	matches := pathParamRe.FindAllStringSubmatch(path, -1)
	for _, m := range matches {
		name, typ := m[1], m[2]
		if seen[name] {
			return nil, fmt.Errorf("路径参数 %q 重复", name)
		}
		seen[name] = true
		if reservedParamNames[name] {
			return nil, fmt.Errorf("路径参数 %q 是保留变量名,请换个名字", name)
		}
		if typ == "" {
			typ = "string"
		}
		if !paramTypeSet[typ] {
			return nil, fmt.Errorf("路径参数 :%s(%s) 类型不支持,可用:%s", name, typ, strings.Join(paramTypeList, "/"))
		}
		out = append(out, &field.Field{
			Name: field.UpperFirst(name),
			Tag:  name,
			Type: &field.Type{Kind: field.KindPrimitive, Name: typ},
		})
	}
	return out, nil
}

// ============ 全局校验 ============

// validateGlobal 跨文件校验:接口名唯一、文件名唯一。
// api/service/handler 等包都是单包,不同 yaml 文件的接口名/文件名撞车会导致生成代码编译冲突
func validateGlobal(yamls []*field.Yaml) error {
	apiOwner := map[string]string{}
	fileOwner := map[string]string{}
	for _, y := range yamls {
		if old, ok := fileOwner[y.FileName]; ok {
			return fmt.Errorf("filename %q 重复(%s 与 %s)", y.FileName, old, y.FileName)
		}
		fileOwner[y.FileName] = y.FileName

		for _, api := range y.ApiList {
			if old, ok := apiOwner[api.ApiName]; ok {
				return fmt.Errorf("接口名 %s 重复(%s 与 %s),api/service 等是单包,重名会编译冲突",
					api.ApiName, old, y.FileName)
			}
			apiOwner[api.ApiName] = y.FileName

			// 接口级 filename 覆盖后的实际归属文件也要查重
			eff := api.FileName
			if eff == "" {
				eff = y.FileName
			}
			if old, ok := fileOwner[eff]; ok && old != y.FileName {
				return fmt.Errorf("接口 %s 的 filename %q 与文件 %s 冲突", api.ApiName, eff, old)
			}
			if _, ok := fileOwner[eff]; !ok {
				fileOwner[eff] = y.FileName
			}
		}
	}
	return nil
}

// ============ 校验工具 ============

var (
	lowerCamelRe = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	upperCamelRe = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	mapTypeRe    = regexp.MustCompile(`^map\[([A-Za-z0-9.]+)\]([A-Za-z0-9.]+)$`)
	pathParamRe  = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)(?:\(([^)]*)\))?`)
)

var methodSet = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true, "delete": true,
}

// paramTypeSet 路径参数允许的类型(string/int/...;不支持 time.Time 等)
var paramTypeSet = map[string]bool{
	"string": true, "bool": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
}

// reservedParamNames 生成代码要用的变量名,路径参数不允许占用
var reservedParamNames = map[string]bool{
	"ctx": true, "c": true, "err": true, "req": true, "res": true,
}

var paramTypeList = []string{
	"string", "bool", "int", "int8", "int16", "int32", "int64",
	"uint", "uint16", "uint32", "uint64", "float32", "float64",
}

const primitiveDoc = "string/int/int8.../uint.../float32/float64/bool/time.Time/time.Duration/any/map[k]v"

func validateLowerCamel(s, what string) error {
	if !lowerCamelRe.MatchString(s) {
		return fmt.Errorf("%s %q 非法:要求小驼峰单词条(小写字母开头,仅字母数字)", what, s)
	}
	return nil
}

func validateUpperCamel(s, what string) error {
	if !upperCamelRe.MatchString(s) {
		return fmt.Errorf("%s %q 非法:要求大驼峰(大写字母开头,仅字母数字)", what, s)
	}
	return nil
}

func validatePath(p string) error {
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("%q 必须以 / 开头", p)
	}
	if strings.ContainsAny(p, " \t?") {
		return fmt.Errorf("%q 不能包含空格或 ?", p)
	}
	return nil
}

// firstComment 依次取第一个非空注释并清洗
func firstComment(candidates ...string) string {
	for _, c := range candidates {
		if v := cleanComment(c); v != "" {
			return v
		}
	}
	return ""
}

// mergeComment 合并上方注释与行注释,两处注释的信息都保留,如 "用户账号;4-12位"
func mergeComment(head, line string) string {
	switch {
	case head == "":
		return line
	case line == "":
		return head
	default:
		return head + ";" + line
	}
}

// cleanComment 把 yaml 注释节点清洗成纯文本:
// 去掉每行开头的 "#",多行用空格拼接
func cleanComment(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	var parts []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, " ")
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "文档节点"
	case yaml.SequenceNode:
		return "数组"
	case yaml.MappingNode:
		return "对象"
	case yaml.ScalarNode:
		return "标量"
	case yaml.AliasNode:
		return "引用"
	}
	return "未知"
}
