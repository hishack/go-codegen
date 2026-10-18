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
// isReq=true 时字段 tag 为 json+form(gin 的 query/form 绑定依赖 form 标签);
// file 字段按参考工程惯例只留 form 标签(multipart 不走 json)
func structText(s *field.Struct, isReq bool) string {
	var b strings.Builder
	if s.Comment != "" {
		b.WriteString("// " + s.Name + " " + s.Comment + "\n")
	}
	b.WriteString("type " + s.Name + " struct {\n")
	for _, fl := range s.Fields {
		if fl.Comment != "" {
			b.WriteString("\t// " + fl.Comment + "\n")
		}
		var tag string
		switch {
		case fl.Type.IsFile():
			tag = `form:"` + fl.Tag + `"`
		case isReq:
			tag = `json:"` + fl.Tag + `" form:"` + fl.Tag + `"`
		default:
			tag = `json:"` + fl.Tag + `"`
		}
		b.WriteString("\t" + fl.Name + " " + fl.Type.GoString() + " `" + tag + "`\n")
	}
	b.WriteString("}")
	return b.String()
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
