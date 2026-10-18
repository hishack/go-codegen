package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// HandlerGen 生成 hertz handler 文件:{domain}_handler.go。
// 方法是"绑定参数 -> 调 service -> 包装响应"的胶水代码:
//   - 不存在则插入;签名没变则原样保留(用户可能加过埋点等代码);
//     签名变了才整体替换
type HandlerGen struct {
	cfg *config.CodeGenConfig
}

func NewHandlerGen(cfg *config.CodeGenConfig) *HandlerGen { return &HandlerGen{cfg: cfg} }

func (g *HandlerGen) Name() string { return "handler" }

func (g *HandlerGen) Generate(ctx *GenContext, report *Report) {
	for _, d := range groupDomains(ctx.Yamls) {
		path := filepath.Join(g.cfg.HandlerOutPath, d.Name+"_handler.go")
		f, err := mergeGoFile(path, g.cfg.Module)
		if err != nil {
			report.warn("handler 生成失败(" + d.Name + "): " + err.Error())
			continue
		}
		if !f.Exists() {
			f.SetPackage(pkgNameOfDir(g.cfg.HandlerOutPath, "handler"))
		}

		g.registerStructAndCtor(f, d, report)
		g.registerMethods(f, d, report)
		g.registerImports(f, d)

		if err := writeMerge(f, report); err != nil {
			report.warn("handler 写入失败(" + d.Name + "): " + err.Error())
		}
	}
}

// registerStructAndCtor handler 结构体 + 构造函数(已存在则跳过)
func (g *HandlerGen) registerStructAndCtor(f *merge.File, d *domain, report *Report) {
	// 结构体:仅当不存在时插入(用户可能加过字段)
	if len(f.TypeNames()) == 0 || !containsName(f.TypeNames(), d.handlerType()) {
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
	if len(f.FuncNames("")) == 0 || !containsName(f.FuncNames(""), ctorName) {
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

// registerMethods 注册/更新各接口方法
func (g *HandlerGen) registerMethods(f *merge.File, d *domain, report *Report) {
	for _, api := range d.Apis {
		a, err := f.UpsertFunc("*"+d.handlerType(), api.ApiName, g.methodText(d, api), merge.PolicyAlways)
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
func (g *HandlerGen) registerImports(f *merge.File, d *domain) {
	f.NeedImports("context",
		"github.com/cloudwego/hertz/pkg/app",
		"github.com/cloudwego/hertz/pkg/protocol/consts",
		g.cfg.Module+"/pkg/response")
	if d.hasReq() {
		f.NeedImports(g.cfg.Module + "/internal/api")
	}
	if d.hasVoRes() {
		f.NeedImports(g.cfg.Module + "/internal/dal/vo")
	}
	if d.hasTypedParam() {
		f.NeedImports("strconv", g.cfg.Module+"/pkg/errorutil")
	}
}

// methodText 渲染一个 handler 方法(胶水代码)
func (g *HandlerGen) methodText(d *domain, api *field.Api) string {
	comment := docComment(api)

	var b strings.Builder
	// 文档注释 + swagger 注解
	b.WriteString("// " + api.ApiName + " " + comment + "\n")
	b.WriteString("// @Summary " + comment + "\n")
	b.WriteString("// @Description " + comment + "\n")
	b.WriteString("// @Tags " + d.Pkg + "\n")
	b.WriteString("// @Accept json\n")
	b.WriteString("// @Produce json\n")
	for _, p := range api.Params {
		b.WriteString("// @Param " + p.Tag + " path " + swaggerType(p.Type) + " true \"" + p.Tag + "\"\n")
	}
	if api.HasReq() {
		if api.Method == "GET" || api.Method == "DELETE" {
			for _, fl := range api.Req.Fields {
				c := fl.Comment
				if c == "" {
					c = fl.Tag
				}
				b.WriteString("// @Param " + fl.Tag + " query " + swaggerType(fl.Type) + " false \"" + c + "\"\n")
			}
		} else {
			b.WriteString("// @Param req body api." + api.Req.Name + " true \"" + comment + "请求\"\n")
		}
	}
	if api.HasRes() {
		b.WriteString("// @Success 200 {object} api." + responseAlias(g.cfg, api) + " \"" + comment + "响应\"\n")
	} else {
		b.WriteString("// @Success 200 {object} bool \"" + comment + "响应\"\n")
	}
	b.WriteString("// @Router " + api.RoutePath() + " [" + strings.ToLower(api.Method) + "]\n")

	b.WriteString("func (u *" + d.handlerType() + ") " + api.ApiName +
		"(ctx context.Context, c *app.RequestContext) {\n")

	errStmt := "c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))"
	paramsErrorStmt := "c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))"
	section := func() { b.WriteString("\n") }

	// 1. 绑定 body/query 参数
	if api.HasReq() {
		b.WriteString("\treq := &api." + api.Req.Name + "{}\n")
		b.WriteString("\terr := c.BindAndValidate(req)\n")
		b.WriteString("\tif err != nil {\n\t\t" + errStmt + "\n\t\treturn\n\t}\n")
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

	// 3. 调用 service
	call := "u." + d.handlerSvcField() + "." + api.ApiName + "(" + serviceArgs(api) + ")"
	switch {
	case api.HasRes():
		b.WriteString("\tres, err := " + call + "\n")
	case api.HasReq() || len(api.Params) > 0:
		b.WriteString("\terr = " + call + "\n")
	default:
		b.WriteString("\terr := " + call + "\n")
	}
	b.WriteString("\tif err != nil {\n\t\t" + errStmt + "\n\t\treturn\n\t}\n")

	// 4. 包装响应
	if api.HasRes() {
		b.WriteString("\tc.JSON(consts.StatusOK, response.NewSuccessResponse[" + resDataT(api) + "](res))\n")
	} else {
		b.WriteString("\tc.JSON(consts.StatusOK, response.NewSuccessResponse[any](true))\n")
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
