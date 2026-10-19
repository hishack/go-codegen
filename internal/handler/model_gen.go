package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// ModelGen 生成模型层(参考 gen_review_resource 的 model/gen 布局,req/res 同父目录分包):
//
//	{modelOut}/request/{domain}.go    package request   请求结构体 {Api}Req
//	{modelOut}/response/{domain}.go   package response  响应结构体 {Api}Res
//
// 文件由生成器全权管理(字段与 yaml 严格同步,内容不一致整体替换)
type ModelGen struct {
	cfg *config.CodeGenConfig
}

func NewModelGen(cfg *config.CodeGenConfig) *ModelGen { return &ModelGen{cfg: cfg} }

func (g *ModelGen) Name() string { return "model" }

func (g *ModelGen) Generate(ctx *GenContext, report *Report) {
	forEachDomain(ctx, report, "model", func(d *domain) error {
		if err := g.genSide(d, "request", "request", true, report); err != nil {
			return err
		}
		return g.genSide(d, "response", "response", false, report)
	})
}

// genSide 生成模型的一侧(request 或 response)
func (g *ModelGen) genSide(d *domain, sub, pkg string, isReq bool, report *Report) error {
	path := filepath.Join(g.cfg.ModelOutPath, sub, d.Name+".go")
	f, err := openGoFile(path, g.cfg.Module, pkg, genHeader(g.cfg), report)
	if err != nil {
		return err
	}
	for _, api := range d.Apis {
		for _, s := range modelStructs(api, isReq) {
			if hasTimeField(s) {
				f.NeedImports("time")
			}
			if isReq && hasFileField(s) {
				// file 字段的 GoString 是 *multipart.FileHeader
				f.NeedImports("mime/multipart")
			}
			a, err := f.UpsertType(structText(s, isReq), merge.PolicyAlways)
			if err != nil {
				return err
			}
			report.add(a)
		}
	}
	return f.Flush()
}

// modelStructs 取一个接口在某侧要生成的全部结构体:
// request 侧取 req 结构体树;response 侧取 res 结构体树(标量 res 不生成结构体)
func modelStructs(api *field.Api, isReq bool) []*field.Struct {
	if isReq {
		return structTypes(api.Req)
	}
	if !api.HasRes() || api.ScalarRes() != nil {
		return nil
	}
	return structTypes(api.Res)
}

func (g *ModelGen) genRequest(d *domain, report *Report) error {
	dir := filepath.Join(g.cfg.ModelOutPath, "request")
	path := filepath.Join(dir, d.Name+".go")
	f, err := mergeGoFile(path, g.cfg.Module)
	if err != nil {
		return err
	}
	if !f.Exists() {
		f.SetPackage("request")
		f.SetHeader(genHeader(g.cfg)...)
	}

	for _, api := range d.Apis {
		if !api.HasReq() {
			continue
		}
		for _, s := range structTypes(api.Req) {
			if hasTimeField(s) {
				f.NeedImports("time")
			}
			if hasFileField(s) {
				// file 字段的 GoString 是 *multipart.FileHeader
				f.NeedImports("mime/multipart")
			}
			a, err := f.UpsertType(structText(s, true), merge.PolicyAlways)
			if err != nil {
				return err
			}
			report.add(a)
		}
	}
	return writeMerge(f, report)
}

func (g *ModelGen) genResponse(d *domain, report *Report) error {
	dir := filepath.Join(g.cfg.ModelOutPath, "response")
	path := filepath.Join(dir, d.Name+".go")
	f, err := mergeGoFile(path, g.cfg.Module)
	if err != nil {
		return err
	}
	if !f.Exists() {
		f.SetPackage("response")
		f.SetHeader(genHeader(g.cfg)...)
	}

	for _, api := range d.Apis {
		// 标量 res(单个基础类型字段)不生成结构体,service 直接返回该类型
		if !api.HasRes() || api.ScalarRes() != nil {
			continue
		}
		for _, s := range structTypes(api.Res) {
			if hasTimeField(s) {
				f.NeedImports("time")
			}
			a, err := f.UpsertType(structText(s, false), merge.PolicyAlways)
			if err != nil {
				return err
			}
			report.add(a)
		}
	}
	return writeMerge(f, report)
}

// ============ 结构体渲染 ============

// structText 渲染一个结构体声明。
// isReq=true 时字段 tag 为 json+form+binding(gin 的 query/form 绑定依赖 form 标签,
// 校验依赖 binding 标签);file 字段按参考工程惯例只留 form 标签(multipart 不走 json)
func structText(s *field.Struct, isReq bool) string {
	type row struct {
		comment string
		name    string
		typ     string
		tag     string
	}
	var rows []row
	nameW, typeW := 0, 0
	for _, fl := range s.Fields {
		var tag string
		switch {
		case fl.Type.IsFile():
			tag = `form:"` + fl.Tag + `"`
		case isReq:
			tag = `json:"` + fl.Tag + `" form:"` + fl.Tag + `"`
			if bt := bindingTag(fl); bt != "" {
				tag += ` binding:"` + bt + `"`
			}
		default:
			tag = `json:"` + fl.Tag + `"`
		}
		// swag 不认识 time.Duration(底层是int64),会报
		// "cannot find type definition",补 swaggertype 让它按整数输出
		if fl.Type.Kind == field.KindPrimitive && fl.Type.Name == "time.Duration" {
			tag += ` swaggertype:"integer"`
		}
		typ := fl.Type.GoString()
		if len(fl.Name) > nameW {
			nameW = len(fl.Name)
		}
		if len(typ) > typeW {
			typeW = len(typ)
		}
		rows = append(rows, row{comment: fl.Comment, name: fl.Name, typ: typ, tag: tag})
	}

	var b strings.Builder
	if s.Comment != "" {
		b.WriteString("// " + s.Name + " " + s.Comment + "\n")
	}
	b.WriteString("type " + s.Name + " struct {\n")
	// 字段名与类型按列对齐,与 gofmt 的输出保持一致
	for _, r := range rows {
		if r.comment != "" {
			b.WriteString("\t// " + r.comment + "\n")
		}
		b.WriteString("\t" + r.name +
			strings.Repeat(" ", nameW-len(r.name)+1) +
			r.typ +
			strings.Repeat(" ", typeW-len(r.typ)+1) +
			"`" + r.tag + "`\n")
	}
	b.WriteString("}")
	return b.String()
}

// bindingTag 按字段类型与命名自动推导 gin 的 binding 校验规则。
//
// yaml 里不必显式声明约束,生成器已知类型即可给出基础校验,
// 让非法请求在 ShouldBind 阶段就被拦下,不必把脏数据传给 logic 层。
//
// 推导规则:
//
//	关键标识字段(name/title/... 的 string) -> required,max=255
//	其余 string-> omitempty,max=255
//	整型 / 浮点                              -> omitempty,gte=0
//	bool                                    -> omitempty
//	嵌套对象 / 切片                          -> omitempty,dive
//	时间、any、file                         -> omitempty
//
// 注意:切片与嵌套对象一律用 omitempty 而非 required。
// yaml 里无法区分"必填数组"与"可选数组",一律强制必填会让
// 合法请求被误拒(如可选的附件列表),因此只校验"传了之后是否合法"。
func bindingTag(f *field.Field) string {
	if f == nil || f.Type == nil {
		return ""
	}
	switch f.Type.Kind {
	case field.KindPrimitive:
		switch f.Type.Name {
		case "string":
			if isKeyIdent(f) {
				return "required,max=255"
			}
			return "omitempty,max=255"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			return "omitempty,gte=0"
		default:
			return "omitempty"
		}
	case field.KindStruct, field.KindSlice:
		return "omitempty,dive"
	default:
		return "omitempty"
	}
}

// keyIdents 视为"关键标识"的字段名,命中则要求必填
var keyIdents = map[string]bool{
	"name": true, "title": true, "username": true,
	"password": true, "phone": true, "email": true,
}

// isKeyIdent 判断字段是否属于"缺失即无意义"的关键标识字段。
//
// 只按"完整的词"匹配,不按字符串后缀:userName 的末段是 Name 算命中,
// 而 nickname 整体是一个词、并非 name 的派生,不应被判为必填,
// 否则可选的昵称字段会被误判成必填,合法请求直接被拒
func isKeyIdent(f *field.Field) bool {
	if f.OmitEmpty {
		return false
	}
	return keyIdents[identOf(f.Tag)]
}

// identOf 取字段名的末段单词:userName -> name、user_name -> name、nickname -> nickname
func identOf(tag string) string {
	if i := strings.LastIndexAny(tag, "_-"); i >= 0 {
		return strings.ToLower(tag[i+1:])
	}
	// 驼峰:从末尾往前找每个大写字母的起点
	runes := []rune(tag)
	for i := len(runes) - 1; i > 0; i-- {
		if runes[i] >= 'A' && runes[i] <= 'Z' {
			return strings.ToLower(string(runes[i:]))
		}
	}
	return strings.ToLower(tag)
}

// structTypes 收集 root 结构体及其全部嵌套结构体(含数组元素),按声明顺序返回
func structTypes(root *field.Struct) []*field.Struct {
	if root == nil {
		return nil
	}
	var out []*field.Struct
	seen := map[*field.Struct]bool{}
	var walk func(t *field.Type)
	walk = func(t *field.Type) {
		if t == nil {
			return
		}
		switch t.Kind {
		case field.KindStruct:
			if !seen[t.Struct] {
				seen[t.Struct] = true
				out = append(out, t.Struct)
			}
			for _, f := range t.Struct.Fields {
				walk(f.Type)
			}
		case field.KindSlice:
			walk(t.Elem)
		}
	}
	walk(&field.Type{Kind: field.KindStruct, Struct: root})
	return out
}

func hasTimeField(s *field.Struct) bool {
	for _, fl := range s.Fields {
		if typeUsesTime(fl.Type) {
			return true
		}
	}
	return false
}

func typeUsesTime(t *field.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case field.KindPrimitive:
		return t.Name == "time.Time" || t.Name == "time.Duration"
	case field.KindSlice:
		return typeUsesTime(t.Elem)
	case field.KindMap:
		return typeUsesTime(t.Key) || typeUsesTime(t.Elem)
	}
	return false
}

func hasFileField(s *field.Struct) bool {
	for _, fl := range s.Fields {
		if fl.Type.IsFile() {
			return true
		}
	}
	return false
}

// writeMerge 把合并结果写盘;内容无变化时不写,保持幂等
// pkgNameOfDir 由输出目录推断包名:internal/service -> service
