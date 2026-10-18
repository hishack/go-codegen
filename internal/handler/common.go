package handler

import (
	"strconv"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
)

// ============ 命名约定(所有生成器共用,保证各层名字一致) ============
//
// 以 yaml 文件 user.yaml、接口 CreateUser 为例(前缀 prefix="YiKou" 时):
//
//	请求结构体    api.YiKouCreateUserRequest           (api 包)
//	响应别名      api.YiKouCreateUserResponse          (api 包,swagger 用)
//	响应数据      vo.CreateUserVo(及嵌套 ...VoListItem)(vo 包)
//	service 接口  IUserService,方法 CreateUser         (service 包)
//	logic 结构体  UserService,方法 CreateUser          (logic 包)
//	handler       UserHandler,方法 CreateUser          (handler 包)
//	路由函数      RegisterUserRoutes                    (router 包)
//	前端类        UserApi,方法 createUser               (web/src/api)

// pascal 域名大驼峰:user -> User
func (d *domain) pascal() string { return field.UpperFirst(d.Name) }

// serviceIface service 接口名:IUserService
func (d *domain) serviceIface() string { return "I" + d.pascal() + "Service" }

// handlerType handler 结构体名:UserHandler
func (d *domain) handlerType() string { return d.pascal() + "Handler" }

// handlerSvcField handler 里的 service 字段名:userService
func (d *domain) handlerSvcField() string { return field.LowerFirst(d.pascal()) + "Service" }

// logicType logic 结构体名:UserService(与 shack 项目保持一致)
func (d *domain) logicType() string { return d.pascal() + "Service" }

// logicCtor logic 构造函数名:NewUserService
func (d *domain) logicCtor() string { return "New" + d.logicType() }

// routerFunc 路由注册函数名:RegisterUserRoutes
func (d *domain) routerFunc() string { return "Register" + d.pascal() + "Routes" }

// routerGroupVar 路由分组变量名:userGroup
func (d *domain) routerGroupVar() string { return field.LowerFirst(d.pascal()) + "Group" }

// tsClass 前端 API 类名:UserApi
func (d *domain) tsClass() string { return d.pascal() + "Api" }

// responseAlias 响应别名(swagger 文档引用):YiKouCreateUserResponse
func responseAlias(cfg *config.CodeGenConfig, api *field.Api) string {
	return cfg.Prefix + api.ApiName + "Response"
}

// docComment 取接口注释,没有则退化为接口名
func docComment(api *field.Api) string {
	if api.Comment != "" {
		return api.Comment
	}
	return api.ApiName
}

// ============ service 方法签名(接口/logic/handler 三处共用) ============

// serviceParams 生成参数列表:"ctx context.Context, userId int64, req *api.XRequest"
func serviceParams(api *field.Api) string {
	parts := []string{"ctx context.Context"}
	for _, p := range api.Params {
		parts = append(parts, field.LowerFirst(p.Name)+" "+p.Type.GoString())
	}
	if api.HasReq() {
		parts = append(parts, "req *api."+api.Req.Name)
	}
	return strings.Join(parts, ", ")
}

// serviceResults 生成返回值列表:"(vo.XVo, error)" / "(int64, error)" / "error"
func serviceResults(api *field.Api) string {
	if t := api.ScalarRes(); t != nil {
		return "(" + t.GoString() + ", error)"
	}
	if api.HasRes() {
		return "(vo." + api.Res.Name + ", error)"
	}
	return "error"
}

// serviceSig 完整方法签名:"CreateUser(ctx context.Context, req *api.XRequest) (vo.XVo, error)"
func serviceSig(api *field.Api) string {
	return api.ApiName + "(" + serviceParams(api) + ") " + serviceResults(api)
}

// resDataT handler 里 NewSuccessResponse[T] 的泛型实参
func resDataT(api *field.Api) string {
	if t := api.ScalarRes(); t != nil {
		return t.GoString()
	}
	if api.HasRes() {
		return "vo." + api.Res.Name
	}
	return "any"
}

// serviceArgs 生成调用 service 的实参列表:"ctx, userId, req"
func serviceArgs(api *field.Api) string {
	parts := []string{"ctx"}
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
