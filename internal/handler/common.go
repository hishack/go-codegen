package handler

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
)

// ============ 命名约定(所有生成器共用,对齐 gen_review_resource 的 gin 工程) ============
//
// 以 yaml 文件 user.yaml、接口 CreateUser 为例:
//
//	请求结构体    request.CreateUserReq      (internal/model/gen/request/user.go)
//	响应结构体    response.CreateUserRes     (internal/model/gen/response/user.go)
//	service 接口  IUserService.CreateUser    (internal/service/user_service.go)
//	logic 结构体  UserService.CreateUser     (internal/logic/user_logic.go)
//	handler       UserHandler.CreateUser     (internal/handler/user_handler.go)
//	路由          UserRouter.InitUserRouter  (internal/router/user_router.go)
//	前端          UserApi.createUser         (web/src/api/user.ts)
//
// request 包按原名引用,response 包与 pkg/response 重名,统一用 res 别名

// modelImportPath 模型子包的 import 路径
func modelImportPath(cfg *config.CodeGenConfig, sub string) string {
	return cfg.Module + "/internal/model/gen/" + sub
}

// pascal 域名大驼峰:user -> User
func (d *domain) pascal() string { return field.UpperFirst(d.Name) }

// serviceIface service 接口名:IUserService
func (d *domain) serviceIface() string { return "I" + d.pascal() + "Service" }

// handlerType handler 结构体名:UserHandler
func (d *domain) handlerType() string { return d.pascal() + "Handler" }

// handlerSvcField handler 里的 service 字段名:userService
func (d *domain) handlerSvcField() string { return field.LowerFirst(d.pascal()) + "Service" }

// logicType logic 结构体名:UserService
func (d *domain) logicType() string { return d.pascal() + "Service" }

// logicCtor logic 构造函数名:NewUserService
func (d *domain) logicCtor() string { return "New" + d.logicType() }

// routerType 路由结构体名:UserRouter
func (d *domain) routerType() string { return d.pascal() + "Router" }

// routerInit 路由注册方法名:InitUserRouter
func (d *domain) routerInit() string { return "Init" + d.pascal() + "Router" }

// routerGroupVar 路由分组变量名:userRouter
func (d *domain) routerGroupVar() string { return field.LowerFirst(d.pascal()) + "Router" }

// tsClass 前端 API 类名:UserApi
func (d *domain) tsClass() string { return d.pascal() + "Api" }

// docComment 取接口注释,没有则退化为接口名
func docComment(api *field.Api) string {
	if api.Comment != "" {
		return api.Comment
	}
	return api.ApiName
}

// ============ service 方法签名(接口/logic/handler 三处共用) ============

// serviceParams 生成参数列表:"c *gin.Context, userId int64, req request.XxxReq"
func serviceParams(api *field.Api) string {
	parts := []string{"c *gin.Context"}
	for _, p := range api.Params {
		parts = append(parts, field.LowerFirst(p.Name)+" "+p.Type.GoString())
	}
	if api.HasReq() {
		parts = append(parts, "req request."+api.Req.Name)
	}
	return strings.Join(parts, ", ")
}

// serviceResults 生成返回值列表:"(res.XxxRes, error)" / "(int64, error)" / "error"
func serviceResults(api *field.Api) string {
	if t := api.ScalarRes(); t != nil {
		return "(" + t.GoString() + ", error)"
	}
	if api.HasRes() {
		return "(res." + api.Res.Name + ", error)"
	}
	return "error"
}

// serviceSig 完整方法签名:"CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error)"
func serviceSig(api *field.Api) string {
	return api.ApiName + "(" + serviceParams(api) + ") " + serviceResults(api)
}

// resDataT handler 里 NewSuccessResponse[T] 的泛型实参
func resDataT(api *field.Api) string {
	if t := api.ScalarRes(); t != nil {
		return t.GoString()
	}
	if api.HasRes() {
		return "res." + api.Res.Name
	}
	return "any"
}

// serviceArgs 生成调用 service 的实参列表:"c, userId, req"
func serviceArgs(api *field.Api) string {
	parts := []string{"c"}
	for _, p := range api.Params {
		parts = append(parts, field.LowerFirst(p.Name))
	}
	if api.HasReq() {
		parts = append(parts, "req")
	}
	return strings.Join(parts, ", ")
}

// zeroValue 基础类型的零值字面量(logic 空实现用)
func zeroValue(t *field.Type) string {
	switch t.Name {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "time.Time":
		return "time.Time{}"
	case "any", "interface{}":
		return "nil"
	default:
		return "0" // 各类整型、浮点、time.Duration
	}
}

// reqHasFile req 里是否有文件上传字段(决定 handler 的绑定方式)
func reqHasFile(api *field.Api) bool {
	if !api.HasReq() {
		return false
	}
	for _, f := range api.Req.Fields {
		if f.Type.IsFile() {
			return true
		}
	}
	return false
}

// bindExpr 生成 gin 参数绑定语句
func bindExpr(api *field.Api) string {
	switch {
	case reqHasFile(api):
		return "c.ShouldBind(&req)" // multipart 表单
	case api.Method == "GET" || api.Method == "DELETE":
		return "c.ShouldBindQuery(&req)" // query 参数
	default:
		return "c.ShouldBindJSON(&req)" // json body
	}
}

// pathParamParse 生成路径参数解析语句。
// 语句中的 "returnError" 占位由调用方替换成实际的错误返回代码。
// 返回 (语句列表, 是否用到 strconv)
func pathParamParse(name, tag string, t *field.Type) ([]string, bool) {
	if t.Name == "string" {
		return []string{name + " := c.Param(" + strconv.Quote(tag) + ")"}, false
	}
	raw := name + "Str"
	needsConv := false
	parse := "strconv.ParseInt(" + raw + ", 10, 64)"
	switch t.Name {
	case "int":
		parse = "strconv.Atoi(" + raw + ")"
	case "int64", "uint64":
	case "int8", "int16", "int32", "uint8", "uint16", "uint32":
		parse = "strconv.ParseInt(" + raw + ", 10, " + bitsize(t.Name) + ")"
		needsConv = true
	case "uint":
		parse = "strconv.ParseUint(" + raw + ", 10, 64)"
		needsConv = true
	case "float32":
		parse = "strconv.ParseFloat(" + raw + ", 32)"
		needsConv = true
	case "float64":
		parse = "strconv.ParseFloat(" + raw + ", 64)"
	case "bool":
		parse = "strconv.ParseBool(" + raw + ")"
	}

	lines := []string{raw + " := c.Param(" + strconv.Quote(tag) + ")"}
	if !needsConv {
		lines = append(lines,
			name+", err := "+parse,
			"if err != nil {",
			"\treturnError",
			"}",
		)
		return lines, true
	}
	conv := t.Name
	switch t.Name {
	case "uint8", "uint16", "uint32":
		parse = "strconv.ParseUint(" + raw + ", 10, " + bitsize(t.Name) + ")"
		conv = "uint" + strings.TrimPrefix(t.Name, "uint")
	case "uint":
		conv = "uint64"
	case "int8", "int16", "int32":
		conv = "int" + strings.TrimPrefix(t.Name, "int")
	}
	lines = append(lines,
		raw+"Parsed, err := "+parse,
		"if err != nil {",
		"\treturnError",
		"}",
		name+" := "+conv+"("+raw+"Parsed)",
	)
	return lines, true
}

func bitsize(name string) string {
	switch name {
	case "int8", "uint8":
		return "8"
	case "int16", "uint16":
		return "16"
	case "int32", "uint32":
		return "32"
	}
	return "64"
}

// swaggerType 基础类型对应的 swagger 类型
func swaggerType(t *field.Type) string {
	switch t.Name {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "file":
		return "file"
	}
	return "string"
}

// ============ 前端 TS 类型映射 ============

// tsType Go 类型 -> TS 类型。nested 把嵌套结构体映射成 TS interface 名
func tsType(t *field.Type, nested func(s *field.Struct) string) string {
	if t == nil {
		return "unknown"
	}
	switch t.Kind {
	case field.KindPrimitive:
		switch t.Name {
		case "bool":
			return "boolean"
		case "string", "time.Time":
			return "string"
		case "any", "interface{}":
			return "unknown"
		case "file":
			return "File"
		case "time.Duration":
			return "number"
		default:
			return "number" // 各类整型浮点
		}
	case field.KindSlice:
		return tsType(t.Elem, nested) + "[]"
	case field.KindMap:
		return "Record<" + tsType(t.Key, nested) + ", " + tsType(t.Elem, nested) + ">"
	case field.KindStruct:
		return nested(t.Struct)
	}
	return "unknown"
}

// ============ 公共小工具 ============

// genHeader 生成文件头(作者署名)
func genHeader(cfg *config.CodeGenConfig) []string {
	return []string{
		"Code generated by go-codegen. DO NOT EDIT.",
		"Author: " + cfg.UserName + "  Github: " + cfg.Github,
	}
}

// pkgNameOfDir 由输出目录推断包名:internal/service -> service
func pkgNameOfDir(dir, fallback string) string {
	base := filepath.Base(strings.TrimRight(filepath.ToSlash(dir), "/"))
	if base == "" || strings.ContainsAny(base, " .") {
		return fallback
	}
	return base
}

// apiHasTypedParam 接口是否存在需要 strconv 解析的路径参数(会声明 err 变量)
func apiHasTypedParam(api *field.Api) bool {
	for _, p := range api.Params {
		if p.Type.Name != "string" {
			return true
		}
	}
	return false
}

// handlerDocBlock 渲染 handler 方法的文档注释 + swagger 注解块(含 @Router)
func handlerDocBlock(d *domain, api *field.Api) string {
	comment := docComment(api)
	var b strings.Builder
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
		switch {
		case api.Method == "GET" || api.Method == "DELETE":
			for _, fl := range api.Req.Fields {
				c := fl.Comment
				if c == "" {
					c = fl.Tag
				}
				b.WriteString("// @Param " + fl.Tag + " query " + swaggerType(fl.Type) + " false \"" + c + "\"\n")
			}
		case reqHasFile(api):
			// multipart 表单:文件字段逐个标注 formData
			for _, fl := range api.Req.Fields {
				if fl.Type.IsFile() {
					c := fl.Comment
					if c == "" {
						c = fl.Tag
					}
					b.WriteString("// @Param " + fl.Tag + " formData file true \"" + c + "\"\n")
				}
			}
		default:
			b.WriteString("// @Param req body request." + api.Req.Name + " true \"" + comment + "请求\"\n")
		}
	}
	switch {
	case api.HasRes() && api.ScalarRes() != nil:
		b.WriteString("// @Success 200 {object} " + api.ScalarRes().GoString() + " \"" + comment + "响应\"\n")
	case api.HasRes():
		b.WriteString("// @Success 200 {object} res." + api.Res.Name + " \"" + comment + "响应\"\n")
	default:
		b.WriteString("// @Success 200 {object} bool \"" + comment + "响应\"\n")
	}
	b.WriteString("// @Router " + api.RoutePath() + " [" + strings.ToLower(api.Method) + "]\n")
	return b.String()
}
