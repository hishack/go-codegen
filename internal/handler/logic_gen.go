package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// ServiceGen 生成 service 接口文件:{domain}_service.go。
// 接口签名完全由 yaml 推导,签名变化时直接更新(接口里没有用户逻辑);
// 用户自己加在接口上的其它方法不会被删除(只告警不清理)
type ServiceGen struct {
	cfg *config.CodeGenConfig
}

func NewServiceGen(cfg *config.CodeGenConfig) *ServiceGen { return &ServiceGen{cfg: cfg} }

func (g *ServiceGen) Name() string { return "service" }

func (g *ServiceGen) Generate(ctx *GenContext, report *Report) {
	for _, d := range groupDomains(ctx.Yamls) {
		path := filepath.Join(g.cfg.ServiceIntOutPath, d.Name+"_service.go")
		f, err := mergeGoFile(path, g.cfg.Module)
		if err != nil {
			report.warn("service 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		if !f.Exists() {
			f.SetPackage(pkgNameOfDir(g.cfg.ServiceIntOutPath, "service"))
		}

		f.NeedImports("context")
		if d.hasReq() {
			f.NeedImports(g.cfg.Module + "/internal/api")
		}
		if d.hasVoRes() {
			f.NeedImports(g.cfg.Module + "/internal/dal/vo")
		}

		iface := d.serviceIface()
		if !f.HasInterface(iface) {
			// 接口不存在:整体创建
			var b strings.Builder
			b.WriteString("type " + iface + " interface {\n")
			for _, api := range d.Apis {
				b.WriteString("\t" + serviceSig(api) + "\n")
			}
			b.WriteString("}")
			a, err := f.UpsertType(b.String(), merge.PolicyAlways)
			if err != nil {
				report.warn("service 生成失败(" + d.Name + "): " + err.Error())
				continue
			}
			report.add(a)
		} else {
			for _, api := range d.Apis {
				a, err := f.UpsertInterfaceMethod(iface, serviceSig(api))
				if err != nil {
					report.warn("service 生成失败(" + d.Name + "): " + err.Error())
					continue
				}
				report.add(a)
			}
			// yaml 里删掉的接口,用户代码里的方法不自动删,提醒手动清理
			for _, name := range f.InterfaceMethodNames(iface) {
				if !d.hasApi(name) {
					report.warn("接口 " + iface + "." + name + " 已不在 yaml 中定义,已保留,请手动清理")
				}
			}
		}
		if err := writeMerge(f, report); err != nil {
			report.warn("service 写入失败(" + d.Name + "): " + err.Error())
		}
	}
}

// hasApi 判断 domain 里是否有指定名字的接口
func (d *domain) hasApi(name string) bool {
	for _, a := range d.Apis {
		if a.ApiName == name {
			return true
		}
	}
	return false
}

// LogicGen 生成 logic 空实现文件:{domain}_logic.go。
// 这是最"用户所有"的一层:已存在的方法一律不碰,签名冲突时只告警
type LogicGen struct {
	cfg *config.CodeGenConfig
}

func NewLogicGen(cfg *config.CodeGenConfig) *LogicGen { return &LogicGen{cfg: cfg} }

func (g *LogicGen) Name() string { return "logic" }

func (g *LogicGen) Generate(ctx *GenContext, report *Report) {
	for _, d := range groupDomains(ctx.Yamls) {
		path := filepath.Join(g.cfg.LogicOutPath, d.Name+"_logic.go")
		f, err := mergeGoFile(path, g.cfg.Module)
		if err != nil {
			report.warn("logic 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		if !f.Exists() {
			f.SetPackage(pkgNameOfDir(g.cfg.LogicOutPath, "logic"))
		}

		f.NeedImports("context")
		if d.hasReq() {
			f.NeedImports(g.cfg.Module + "/internal/api")
		}
		if d.hasVoRes() {
			f.NeedImports(g.cfg.Module + "/internal/dal/vo")
		}
		if d.hasTimeScalarRes() {
			f.NeedImports("time")
		}

		// 结构体和构造函数:已存在就不动(用户可能加了 db/redis 等依赖字段)
		a, err := f.UpsertType("type "+d.logicType()+" struct {}", merge.PolicyIfAbsent)
		if err != nil {
			report.warn("logic 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		report.add(a)
		a, err = f.UpsertFunc("", d.logicCtor(),
			"func "+d.logicCtor()+"() *"+d.logicType()+" {\n\treturn &"+d.logicType()+"{}\n}",
			merge.PolicyIfAbsent)
		if err != nil {
			report.warn("logic 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		report.add(a)

		for _, api := range d.Apis {
			a, err := f.UpsertFunc("*"+d.logicType(), api.ApiName, g.methodText(d, api), merge.PolicyIfAbsent)
			if err != nil {
				report.warn("logic 生成失败(" + d.Name + "): " + err.Error())
				continue
			}
			report.add(a)
		}
		if err := writeMerge(f, report); err != nil {
			report.warn("logic 写入失败(" + d.Name + "): " + err.Error())
		}
	}
}

// methodText logic 空实现方法
func (g *LogicGen) methodText(d *domain, api *field.Api) string {
	var b strings.Builder
	comment := docComment(api)
	b.WriteString("// " + api.ApiName + " " + comment + "\n")
	b.WriteString("func (s *" + d.logicType() + ") " + serviceSig(api) + " {\n")
	b.WriteString("\t// TODO: 实现业务逻辑:" + comment + "\n")

	if t := api.ScalarRes(); t != nil {
		b.WriteString("\treturn " + zeroValue(t) + ", nil\n")
	} else if api.HasRes() {
		b.WriteString("\treturn vo." + api.Res.Name + "{}, nil\n")
	} else {
		b.WriteString("\treturn nil\n")
	}
	b.WriteString("}")
	return b.String()
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
