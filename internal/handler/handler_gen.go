package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// HandlerGen 生成 gin handler 文件:{domain}_handler.go。
// handler 是"绑定参数 -> 调 service -> 包装响应"的纯胶水层,由生成器全权管理
// (内容不一致整体替换);业务逻辑请写在 logic 层,那里已存在的方法一律不动
type HandlerGen struct {
	cfg *config.CodeGenConfig
}

func NewHandlerGen(cfg *config.CodeGenConfig) *HandlerGen { return &HandlerGen{cfg: cfg} }

func (g *HandlerGen) Name() string { return "handler" }

func (g *HandlerGen) Generate(ctx *GenContext, report *Report) {
	forEachDomain(ctx, report, "handler", func(d *domain) error {
		return g.genDomain(d, report)
	})
}

func (g *HandlerGen) genDomain(d *domain, report *Report) error {
	path := filepath.Join(g.cfg.HandlerOutPath, d.Name+"_handler.go")
	f, err := openGoFile(path, g.cfg.Module, pkgNameOfDir(g.cfg.HandlerOutPath, "handler"), nil, report)
	if err != nil {
		return err
	}

	g.registerStructAndCtor(f, d, report)
	g.registerMethods(f, d, report)
	g.registerImports(f, d)
	return f.Flush()
}

// registerStructAndCtor handler 结构体 + 构造函数(已存在则跳过)
func (g *HandlerGen) registerStructAndCtor(f *goFile, d *domain, report *Report) {
	if !containsName(f.TypeNames(), d.handlerType()) {
		text := "type " + d.handlerType() + " struct {\n\t" +
			d.handlerSvcField() + " service.I" + d.pascal() + "Service\n}"
		a, err := f.UpsertType(text, merge.PolicyIfAbsent)
		if err != nil {
			report.warn("handler 生成失败(" + d.Name + "): " + err.Error())
			return
		}
		report.add(a)
		f.NeedImports(g.cfg.Module + "/internal/service")
	}

	ctorName := "New" + d.handlerType()
	if !containsName(f.FuncNames(""), ctorName) {
		ctor := "func " + ctorName + "(userService service.I" + d.pascal() + "Service) *" + d.handlerType() + " {\n" +
			"\treturn &" + d.handlerType() + "{\n\t\t" + d.handlerSvcField() + ": userService,\n\t}\n}"
		a, err := f.UpsertFunc("", ctorName, ctor, merge.PolicyIfAbsent)
		if err != nil {
			report.warn("handler 生成失败(" + d.Name + "): " + err.Error())
			return
		}
		report.add(a)
		f.NeedImports(g.cfg.Module + "/internal/service")
	}
}

// registerMethods 注册/更新各接口方法(handler 是生成器管理的胶水层)
func (g *HandlerGen) registerMethods(f *goFile, d *domain, report *Report) {
	for _, api := range d.Apis {
		a, err := f.UpsertFunc("*"+d.handlerType(), api.ApiName, handlerMethodText(d, api), merge.PolicyAlways)
		if err != nil {
			report.warn("handler 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		report.add(a)
	}
	// yaml 里删掉的 handler 方法提醒手动清理
	for _, name := range f.FuncNames(d.handlerType()) {
		if !d.hasApi(name) {
			report.warn("handler 方法 " + d.handlerType() + "." + name + " 已不在 yaml 中定义,已保留,请手动清理")
		}
	}
}

// registerImports handler 文件 import 精确计算
func (g *HandlerGen) registerImports(f *goFile, d *domain) {
	f.NeedImports("github.com/gin-gonic/gin", "net/http", g.cfg.Module+"/pkg/response")
	registerModelImports(f.File, g.cfg, d)
	if d.hasTypedParam() {
		f.NeedImports("strconv", g.cfg.Module+"/pkg/errorutil")
	}
}

// handlerMethodText 渲染一个 handler 方法(胶水代码)
func handlerMethodText(d *domain, api *field.Api) string {
	var b strings.Builder
	b.WriteString(handlerDocBlock(d, api))
	b.WriteString("func (u *" + d.handlerType() + ") " + api.ApiName + "(c *gin.Context) {\n")

	errStmt := "c.JSON(http.StatusOK, response.NewErrorResponse[any](err))"
	paramsErrorStmt := "c.JSON(http.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))"
	section := func() { b.WriteString("\n") }

	// 1. 绑定请求参数(bindExpr 按有无 file 字段选择 ShouldBind / ShouldBindQuery / ShouldBindJSON)
	if api.HasReq() {
		b.WriteString("\tvar req request." + api.Req.Name + "\n")
		b.WriteString("\tif err := " + bindExpr(api) + "; err != nil {\n\t\t" + errStmt + "\n\t\treturn\n\t}\n")
		section()
	}

	// 2. 解析路径参数(pathParamParse 里的 "returnError" 占位替换为参数错误返回)
	for _, p := range api.Params {
		lines, _ := pathParamParse(field.LowerFirst(p.Name), p.Tag, p.Type)
		for _, l := range lines {
			l = strings.ReplaceAll(l, "returnError", paramsErrorStmt)
			b.WriteString("\t" + l + "\n")
		}
	}
	if len(api.Params) > 0 {
		section()
	}

	// 3. 调用 service(路径参数里的类型解析会声明 err;绑定错误在 if 作用域里,不影响这里)
	call := "u." + d.handlerSvcField() + "." + api.ApiName + "(" + serviceArgs(api) + ")"
	switch {
	case api.HasRes():
		b.WriteString("\tdata, err := " + call + "\n")
	case apiHasTypedParam(api):
		b.WriteString("\terr = " + call + "\n")
	default:
		b.WriteString("\terr := " + call + "\n")
	}
	b.WriteString("\tif err != nil {\n\t\t" + errStmt + "\n\t\treturn\n\t}\n")

	// 4. 包装响应
	if api.HasRes() {
		b.WriteString("\tc.JSON(http.StatusOK, response.NewSuccessResponse[" + resDataT(api) + "](data))\n")
	} else {
		b.WriteString("\tc.JSON(http.StatusOK, response.NewSuccessResponse[any](true))\n")
	}
	b.WriteString("}")
	return b.String()
}

func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
