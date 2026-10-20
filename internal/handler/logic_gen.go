package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// LogicGen 生成 logic 空实现文件:{domain}_logic.go。
// 这是最"用户所有"的一层:已存在的方法一律不碰,签名冲突时只告警
type LogicGen struct {
	cfg *config.CodeGenConfig
}

func NewLogicGen(cfg *config.CodeGenConfig) *LogicGen { return &LogicGen{cfg: cfg} }

func (g *LogicGen) Name() string { return "logic" }

func (g *LogicGen) Generate(ctx *GenContext, report *Report) {
	forEachDomain(ctx, report, "logic", func(d *domain) error {
		return g.genDomain(d, report)
	})
}

func (g *LogicGen) genDomain(d *domain, report *Report) error {
	path := filepath.Join(g.cfg.LogicOutPath, d.Name+"_logic.go")
	f, err := openGoFile(path, g.cfg.Module, pkgNameOfDir(g.cfg.LogicOutPath, "logic"), nil, report)
	if err != nil {
		return err
	}

	registerModelImports(f.File, g.cfg, d)
	f.NeedImports("github.com/gin-gonic/gin")
	if d.hasTimeScalarRes() {
		f.NeedImports("time")
	}

	// 结构体和构造函数:已存在就不动(用户可能加了 db 等依赖字段)
	a, err := f.UpsertType("type "+d.logicType()+" struct {}", merge.PolicyIfAbsent)
	if err != nil {
		return err
	}
	report.add(a)
	a, err = f.UpsertFunc("", d.logicCtor(),
		"func "+d.logicCtor()+"() *"+d.logicType()+" {\n\treturn &"+d.logicType()+"{}\n}",
		merge.PolicyIfAbsent)
	if err != nil {
		return err
	}
	report.add(a)

	for _, api := range d.Apis {
		a, err := f.UpsertFunc("*"+d.logicType(), api.ApiName, logicMethodText(d, api), merge.PolicyIfAbsent)
		if err != nil {
			return err
		}
		report.add(a)
	}
	return f.Flush()
}

// logicMethodText logic 空实现方法。
// 方法体只放零值返回,不写 TODO 占位注释 —— 接口说明已经在函数上方的文档注释里,
// 体内再重复一遍只会让每次生成都多出一行噪音,用户实现后还得手动删
func logicMethodText(d *domain, api *field.Api) string {
	var b strings.Builder
	b.WriteString("// " + api.ApiName + " " + docComment(api) + "\n")
	b.WriteString("func (s *" + d.logicType() + ") " + serviceSig(api) + " {\n")

	if t := api.ScalarRes(); t != nil {
		b.WriteString("\treturn " + zeroValue(t) + ", nil\n")
	} else if api.HasRes() {
		b.WriteString("\treturn res." + api.Res.Name + "{}, nil\n")
	} else {
		b.WriteString("\treturn nil\n")
	}
	b.WriteString("}")
	return b.String()
}
