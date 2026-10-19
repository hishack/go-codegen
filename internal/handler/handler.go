package handler

import (
	"errors"
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

// Generator 一类产出物的生成器。扩展新功能时实现本接口,
// 并通过 RegisterGenerator 注册,即可纳入 RunAll 的生成流程
type Generator interface {
	Name() string
	// Generate 执行生成,把过程信息(新增/更新/冲突/告警)追加到 report
	Generate(ctx *GenContext, report *Report)
}

// GeneratorFactory 生成器工厂:配置加载后创建生成器实例
type GeneratorFactory func(cfg *config.CodeGenConfig) Generator

// generatorFactories 内置生成器,按执行顺序排列:
// model(结构体) → service(接口) → logic(空实现) → handler(胶水) → router → web
var generatorFactories []GeneratorFactory

// RegisterGenerator 注册生成器工厂(扩展点)
func RegisterGenerator(f GeneratorFactory) {
	generatorFactories = append(generatorFactories, f)
}

func init() {
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewModelGen(cfg) })
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewServiceGen(cfg) })
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewLogicGen(cfg) })
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewHandlerGen(cfg) })
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewRouterGen(cfg) })
	RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewWebApiGen(cfg) })
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

// RunAll 依次执行所有已注册的生成器。
//
// 生成前先做路由冲突检测:gin 要求同一 METHOD+PATH 唯一、同一层通配符同名,
// 而 yaml 按领域组织,各领域独立定义 path 极易撞车。若不在此处拦住,
// 生成的代码能编译通过,却会在服务启动时 panic。
func RunAll(cfg *config.CodeGenConfig, yamls *field.YamlList) (*Report, error) {
	if err := checkRouteConflicts(yamls, cfg); err != nil {
		return nil, err
	}
	ctx := &GenContext{Cfg: cfg, Yamls: yamls}
	report := &Report{}
	for _, f := range generatorFactories {
		f(cfg).Generate(ctx, report)
	}
	return report, nil
}

// checkRouteConflicts 路由冲突时返回错误,无冲突返回 nil。
//
// 分两层检查:
//  1. 本次 yaml 内部的冲突
//  2. 本次 yaml 与目标工程里已生成的路由之间的冲突(跨多次配置的场景)
func checkRouteConflicts(yamls *field.YamlList, cfg *config.CodeGenConfig) error {
	conflicts := validateRoutes(yamls, cfg)
	if cfg != nil && cfg.RouterOutPath != "" {
		conflicts = append(conflicts, checkRoutesAgainstExisting(yamls, cfg)...)
	}
	if len(conflicts) == 0 {
		return nil
	}
	return errors.New(formatConflicts(conflicts))
}

// ============ 生成器共用工具 ============

// goFile 生成器管理的目标 Go 文件,封装"打开/创建 → 写内容 → 写盘"的公共流程
type goFile struct {
	*merge.File
	report *Report
}

// openGoFile 打开目标文件;不存在或为空时按新建处理(自动设置包名和文件头)
func openGoFile(path, module, pkg string, header []string, report *Report) (*goFile, error) {
	f, err := mergeGoFile(path, module)
	if err != nil {
		return nil, err
	}
	if !f.Exists() {
		f.SetPackage(pkg)
		f.SetHeader(header...)
	}
	return &goFile{File: f, report: report}, nil
}

// Flush 渲染并写盘;内容无变化时不写,保持幂等
func (g *goFile) Flush() error { return writeMerge(g.File, g.report) }

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

// writeMerge 渲染并写盘,同时把合并过程中的告警汇总到 report
func writeMerge(f *merge.File, report *Report) error {
	out, changed, err := f.Render()
	if err != nil {
		return err
	}
	// 告警(如签名冲突)与是否写盘无关,必须始终上报
	for _, w := range f.Warnings() {
		report.warn(w)
	}
	if !changed {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(f.Path()), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(f.Path(), out, 0o644); err != nil {
		return err
	}
	return nil
}

// forEachDomain 按领域分组执行生成逻辑;单个领域失败只告警,不中断其它领域
func forEachDomain(ctx *GenContext, report *Report, layer string, fn func(d *domain) error) {
	for _, d := range groupDomains(ctx.Yamls) {
		if err := fn(d); err != nil {
			report.warn(layer + " 生成失败(" + d.Name + "): " + err.Error())
		}
	}
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

func (d *domain) hasApi(name string) bool {
	for _, a := range d.Apis {
		if a.ApiName == name {
			return true
		}
	}
	return false
}

func (d *domain) hasReq() bool {
	for _, a := range d.Apis {
		if a.HasReq() {
			return true
		}
	}
	return false
}

// hasStructRes 是否存在返回结构体(非标量)的接口
func (d *domain) hasStructRes() bool {
	for _, a := range d.Apis {
		if a.HasRes() && a.ScalarRes() == nil {
			return true
		}
	}
	return false
}

// hasTypedParam 是否存在需要 strconv 解析的路径参数
func (d *domain) hasTypedParam() bool {
	for _, a := range d.Apis {
		if apiHasTypedParam(a) {
			return true
		}
	}
	return false
}

// hasTimeScalarRes 是否存在标量 res 为 time.Time/Duration 的接口(零值会用到 time 包)
func (d *domain) hasTimeScalarRes() bool {
	for _, a := range d.Apis {
		if t := a.ScalarRes(); t != nil && (t.Name == "time.Time" || t.Name == "time.Duration") {
			return true
		}
	}
	return false
}
