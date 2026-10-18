package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
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
	forEachDomain(ctx, report, "service", func(d *domain) error {
		return g.genDomain(d, report)
	})
}

func (g *ServiceGen) genDomain(d *domain, report *Report) error {
	path := filepath.Join(g.cfg.ServiceIntOutPath, d.Name+"_service.go")
	f, err := openGoFile(path, g.cfg.Module, pkgNameOfDir(g.cfg.ServiceIntOutPath, "service"), nil, report)
	if err != nil {
		return err
	}

	registerModelImports(f.File, g.cfg, d)
	f.NeedImports("github.com/gin-gonic/gin")

	iface := d.serviceIface()
	if !f.HasInterface(iface) {
		var b strings.Builder
		b.WriteString("type " + iface + " interface {\n")
		for _, api := range d.Apis {
			b.WriteString("\t" + serviceSig(api) + "\n")
		}
		b.WriteString("}")
		a, err := f.UpsertType(b.String(), merge.PolicyAlways)
		if err != nil {
			return err
		}
		report.add(a)
	} else {
		for _, api := range d.Apis {
			a, err := f.UpsertInterfaceMethod(iface, serviceSig(api))
			if err != nil {
				return err
			}
			report.add(a)
		}
		for _, name := range f.InterfaceMethodNames(iface) {
			if !d.hasApi(name) {
				report.warn("接口 " + iface + "." + name + " 已不在 yaml 中定义,已保留,请手动清理")
			}
		}
	}
	return f.Flush()
}

// registerModelImports 按需登记 request/response 模型包的 import
func registerModelImports(f *merge.File, cfg *config.CodeGenConfig, d *domain) {
	if d.hasReq() {
		f.NeedImports(modelImportPath(cfg, "request"))
	}
	if d.hasStructRes() {
		f.NeedAliasedImport("res", modelImportPath(cfg, "response"))
	}
}
