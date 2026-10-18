package handler

import (
	"os"
	"path/filepath"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// GenContext 传给各生成器的上下文
type GenContext struct {
	Cfg   *config.CodeGenConfig
	Yamls *field.YamlList
}

// Generator 一类产出物的生成器(api/service/handler/logic/router/web)
type Generator interface {
	Name() string
	// Generate 执行生成,把过程信息(新增/更新/冲突/告警)追加到 report
	Generate(ctx *GenContext, report *Report)
}

// Report 汇总一次生成的结果,供命令行打印
type Report struct {
	Added    int      // 新插入的声明数
	Updated  int      // 替换的声明数
	Conflict int      // 签名冲突、保留用户代码的数量
	Warnings []string // 需要用户手动处理的问题
}

func (r *Report) add(a merge.Action) {
	switch a {
	case merge.ActionAdded:
		r.Added++
	case merge.ActionUpdated:
		r.Updated++
	case merge.ActionConflict:
		r.Conflict++
	}
}

func (r *Report) warn(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

// RunAll 依次执行全部生成器。
// 生成顺序:api(结构体) → service(接口) → logic(空实现) → handler(胶水) → router → web。
func RunAll(cfg *config.CodeGenConfig, yamls *field.YamlList) (*Report, error) {
	ctx := &GenContext{Cfg: cfg, Yamls: yamls}
	report := &Report{}
	for _, g := range AllGenerators(cfg) {
		g.Generate(ctx, report)
	}
	return report, nil
}

// AllGenerators 返回生成器列表
func AllGenerators(cfg *config.CodeGenConfig) []Generator {
	return []Generator{
		NewApiGen(cfg),
		NewServiceGen(cfg),
		NewLogicGen(cfg),
		NewHandlerGen(cfg),
		NewRouterGen(cfg),
		NewWebApiGen(cfg),
	}
}

// mergeGoFile 打开目标 Go 文件并确保目录存在,module 用于新建文件的 import 分组
func mergeGoFile(path, module string) (*merge.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := merge.Load(path)
	if err != nil {
		return nil, err
	}
	f.SetModule(module)
	return f, nil
}

// writeMerge 把合并结果写盘;内容无变化时不写,保持幂等
func writeMerge(f *merge.File, report *Report) error {
	out, changed, err := f.Render()
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if err := os.WriteFile(f.Path(), out, 0o644); err != nil {
		return err
	}
	for _, w := range f.Warnings() {
		report.warn(w)
	}
	return nil
}

// ============ 领域分组 ============

// domain 是一次产出文件的分组:同名 filename 的接口生成同一组文件。
// 一个 yaml 文件通常就是一个 domain;接口级 filename 覆盖后会产生多个 domain
type domain struct {
	Name string // 文件名,如 user
	Pkg  string // yaml 里的 package,用于 swagger @Tags
	Apis []*field.Api
}

// groupDomains 按 filename 把所有接口分组,顺序稳定(按 yaml 文件、接口声明顺序)
func groupDomains(yamls *field.YamlList) []*domain {
	var out []*domain
	idx := map[string]*domain{}
	for _, y := range yamls.YamlList {
		for _, api := range y.ApiList {
			name := api.FileName
			if name == "" {
				name = y.FileName
			}
			d, ok := idx[name]
			if !ok {
				d = &domain{Name: name, Pkg: y.PackageName}
				idx[name] = d
				out = append(out, d)
			}
			d.Apis = append(d.Apis, api)
		}
	}
	return out
}

// HasReq 判断 domain 里是否存在需要请求体的接口
func (d *domain) hasReq() bool {
	for _, a := range d.Apis {
		if a.HasReq() {
			return true
		}
	}
	return false
}

// hasVoRes 判断是否存在返回结构体(非标量)的接口
func (d *domain) hasVoRes() bool {
	for _, a := range d.Apis {
		if a.HasRes() && a.ScalarRes() == nil {
			return true
		}
	}
	return false
}

// hasTypedParam 判断是否存在需要 strconv 解析的路径参数
func (d *domain) hasTypedParam() bool {
	for _, a := range d.Apis {
		for _, p := range a.Params {
			if p.Type.Name != "string" {
				return true
			}
		}
	}
	return false
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
