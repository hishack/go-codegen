package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
)

// 本文件覆盖 5 项生成器修复:
//  1. P0-2 路径参数解析失败缺 return,导致输出双段 JSON
//  2. P1-3/P1-4 HTTP 状态码恒为 200
//  3. P1-1 请求字段无 binding 校验标签
//  4. P1-5 swagger path 参数缺format
//  5. 分组前缀 group 配置

// ---- P0-2 路径参数解析 ----

// TestPathParamErrorBranchHasReturn 解析失败的分支必须 return。
// 缺 return 时 c.JSON 已写过一次响应,请求却继续往下走,
// 最终输出两段 JSON,前端 JSON.parse 直接抛异常
func TestPathParamErrorBranchHasReturn(t *testing.T) {
	cases := []struct {
		typ  string
		want int
	}{
		{"int64", 1},
		{"int", 1},
		{"float64", 1},
		{"bool", 1},
		{"int32", 1},
		{"uint", 1},
	}
	for _, c := range cases {
		tp := &field.Type{Kind: field.KindPrimitive, Name: c.typ}
		lines, _ := pathParamParse("id", "id", tp)
		if !hasReturnAfterError(lines) {
			t.Errorf("%s: 错误分支缺少 return:\n%s", c.typ, strings.Join(lines, "\n"))
		}
		if n := countReturns(lines); n != c.want {
			t.Errorf("%s: return 语句数 = %d, 期望 %d", c.typ, n, c.want)
		}
	}
}

// TestStringParamNoErrorBranch string 参数无需类型转换,不应产生错误分支
func TestStringParamNoErrorBranch(t *testing.T) {
	tp := &field.Type{Kind: field.KindPrimitive, Name: "string"}
	lines, usedStrconv := pathParamParse("name", "name", tp)
	if usedStrconv {
		t.Error("string 参数不应需要 strconv")
	}
	for _, l := range lines {
		if strings.Contains(l, "returnError") {
			t.Errorf("string 参数不应有错误分支: %s", l)
		}
	}
}

func hasReturnAfterError(lines []string) bool {
	for i, l := range lines {
		if l == "}" && i >= 2 && strings.Contains(lines[i-1], "return") {
			return true
		}
	}
	return false
}

func countReturns(lines []string) int {
	n := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "return" {
			n++
		}
	}
	return n
}

// ---- P1-3/P1-4 状态码 ----

// TestHandlerStatusCodes 绑定错误与业务错误要用不同状态码
func TestHandlerStatusCodes(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

CreateThing:
  method: POST
  path: /things
  req:
    name: string
  res:
    id: int64
`)
	list := mustParseAll(t, yamlDir)
	d := groupDomains(list)[0]
	api := d.Apis[0]
	text := handlerMethodText(d, api, nil)

	if !strings.Contains(text, "c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))") {
		t.Errorf("绑定失败应返回 400:\n%s", text)
	}
	if !strings.Contains(text, "c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))") {
		t.Errorf("业务失败应经 HTTPStatusOf 映射状态码:\n%s", text)
	}
	if strings.Contains(text, "c.JSON(http.StatusOK, response.NewErrorResponse") {
		t.Errorf("错误响应不应再用 200:\n%s", text)
	}
}

// TestHandlerPathParamStatusCode 路径参数解析失败同样属于 400
func TestHandlerPathParamStatusCode(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

GetThing:
  method: GET
  path: /things/:thingId(int64)
`)
	list := mustParseAll(t, yamlDir)
	d := groupDomains(list)[0]
	text := handlerMethodText(d, d.Apis[0], nil)
	if !strings.Contains(text, "c.JSON(http.StatusBadRequest,") {
		t.Errorf("路径参数失败应返回 400:\n%s", text)
	}
}

// TestHandlerNoErrorStatus200 成功响应保持 200
func TestHandlerNoErrorStatus200(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

GetThing:
  method: GET
  path: /things
  res:
    total: int64
`)
	list := mustParseAll(t, yamlDir)
	d := groupDomains(list)[0]
	text := handlerMethodText(d, d.Apis[0], nil)
	if !strings.Contains(text, "c.JSON(http.StatusOK, response.NewSuccessResponse") {
		t.Errorf("成功响应应为 200:\n%s", text)
	}
}

// ---- P1-1 binding 标签 ----

// TestBindingTagByType 按字段类型自动推导校验规则
func TestBindingTagByType(t *testing.T) {
	cases := []struct {
		name string
		fl   *field.Field
		want string
	}{
		{"关键字符串必填", mkField("Name", "name", "string"), "required,max=255"},
		{"标题必填", mkField("Title", "title", "string"), "required,max=255"},
		{"普通字符串可选", mkField("Remark", "remark", "string"), "omitempty,max=255"},
		{"整数非负", mkField("Age", "age", "int"), "omitempty,gte=0"},
		{"int64 非负", mkField("Score", "score", "int64"), "omitempty,gte=0"},
		{"浮点非负", mkField("Price", "price", "float64"), "omitempty,gte=0"},
		{"布尔可选", mkField("Enabled", "enabled", "bool"), "omitempty"},
	}
	for _, c := range cases {
		if got := bindingTag(c.fl); got != c.want {
			t.Errorf("%s: bindingTag = %q, 期望 %q", c.name, got, c.want)
		}
	}
}

// TestBindingTagNested 嵌套对象与切片只在传入时校验,不强���必填。
// yaml 语法无法区分"必填数组"与"可选数组",一律 required 会误拒合法请求
func TestBindingTagNested(t *testing.T) {
	nested := &field.Field{
		Name: "Detail", Tag: "detail",
		Type: &field.Type{Kind: field.KindStruct, Struct: &field.Struct{Name: "D"}},
	}
	if got := bindingTag(nested); got != "omitempty,dive" {
		t.Errorf("嵌套对象 bindingTag = %q, 期望 omitempty,dive", got)
	}
	slice := &field.Field{
		Name: "Items", Tag: "items",
		Type: &field.Type{Kind: field.KindSlice, Elem: &field.Type{Kind: field.KindPrimitive, Name: "string"}},
	}
	if got := bindingTag(slice); got != "omitempty,dive" {
		t.Errorf("切片 bindingTag = %q, 期望 omitempty,dive", got)
	}
}

// TestBindingTagOmitEmpty 显式 omitempty 的字段不强制必填
func TestBindingTagOmitEmpty(t *testing.T) {
	fl := mkField("Name", "name", "string")
	fl.OmitEmpty = true
	if got := bindingTag(fl); got != "omitempty,max=255" {
		t.Errorf("omitempty 字段应可省略,实际 %q", got)
	}
}

// TestBindingTagKeyIdentWord 必填判定按"完整的词"而非后缀,
// nickname 整体是一个词、不该因为含name 而被判必填
func TestBindingTagKeyIdentWord(t *testing.T) {
	cases := []struct {
		tag      string
		required bool
	}{
		{"name", true},
		{"title", true},
		{"userName", true},
		{"user_name", true},
		{"orderTitle", true},
		{"nickname", false},
		{"remark", false},
		{"address", false},
		{"filename", false},
	}
	for _, c := range cases {
		fl := mkField("X", c.tag, "string")
		got := strings.HasPrefix(bindingTag(fl), "required")
		if got != c.required {
			t.Errorf("%s: 必填判定 = %v, 期望 %v (tag=%q)", c.tag, got, c.required, bindingTag(fl))
		}
	}
}

// TestRequestStructHasBindingTag 生成的请求结构体必须带 binding 标签
func TestRequestStructHasBindingTag(t *testing.T) {
	s := &field.Struct{
		Name: "CreateUserReq",
		Fields: []*field.Field{
			mkField("Name", "name", "string"),
			mkField("Age", "age", "int"),
		},
	}
	text := structText(s, true)
	if !strings.Contains(text, `binding:"required,max=255"`) {
		t.Errorf("name 字段缺少 required 标签:\n%s", text)
	}
	if !strings.Contains(text, `binding:"omitempty,gte=0"`) {
		t.Errorf("age 字段缺少 gte 标签:\n%s", text)
	}
}

// TestResponseStructNoBindingTag 响应结构体不该有 binding 标签,
// 否则回写时会被误校验
func TestResponseStructNoBindingTag(t *testing.T) {
	s := &field.Struct{
		Name:   "UserRes",
		Fields: []*field.Field{mkField("Name", "name", "string")},
	}
	text := structText(s, false)
	if strings.Contains(text, "binding:") {
		t.Errorf("响应结构体不应有 binding 标签:\n%s", text)
	}
}

// ---- P1-5 swagger format ----

// TestSwaggerParamFormat path/query 参数应带 format 属性,让前端拿到具体类型
func TestSwaggerParamFormat(t *testing.T) {
	cases := []struct {
		typ  string
		want string
	}{
		{"int64", "integer format(int64)"},
		{"int", "integer format(int32)"},
		{"int32", "integer format(int32)"},
		{"uint64", "integer format(int64)"},
		{"float64", "number format(double)"},
		{"float32", "number format(float)"},
		{"string", "string"},
		{"bool", "boolean"},
	}
	for _, c := range cases {
		tp := &field.Type{Kind: field.KindPrimitive, Name: c.typ}
		got := swaggerType(tp) + swaggerFormatAttr(tp)
		if got != c.want {
			t.Errorf("%s: swagger 参数 = %q, 期望 %q", c.typ, got, c.want)
		}
	}
}

// TestSwaggerPathParamFormatInAnnotation 生成的注解里 format 必须在描述之后,
// 放在类型位置上 swag 会解析失败
func TestSwaggerPathParamFormatInAnnotation(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

GetThing:
  method: GET
  path: /things/:thingId(int64)
`)
	list := mustParseAll(t, yamlDir)
	d := groupDomains(list)[0]
	text := handlerDocBlock(d, d.Apis[0], nil)
	want := `@Param thingId path integer true "thingId" format(int64)`
	if !strings.Contains(text, want) {
		t.Errorf("path 参数 format 位置不对,期望:\n  %s\n实际:\n%s", want, text)
	}
}

// ---- 分组前缀 ----

// TestNormalizeGroup 前缀需规整:补前导斜杠、去尾斜杠
func TestNormalizeGroup(t *testing.T) {
	cases := [][2]string{
		{"/api/v1", "/api/v1"},
		{"api/v1", "/api/v1"},
		{"/api/v1/", "/api/v1"},
		{"", ""},
		{"/", ""},
		{"  /api/v2  ", "/api/v2"},
	}
	for _, c := range cases {
		if got := config.NormalizeGroup(c[0]); got != c[1] {
			t.Errorf("NormalizeGroup(%q) = %q, 期望 %q", c[0], got, c[1])
		}
	}
}

// TestGroupPrefixInRouter 分组前缀应体现在路由注册的 Group 上
func TestGroupPrefixInRouter(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

ListThings:
  method: GET
  path: /things
`)
	list := mustParseAll(t, yamlDir)
	cfg := mustConfig(config.CodeGenConfig{
		YamlInPath: yamlDir, Module: "m",
		Group:             "/api/v1",
		ServiceIntOutPath: filepath.Join(dir, "svc"),
		ModelOutPath:      filepath.Join(dir, "model"),
		LogicOutPath:      filepath.Join(dir, "logic"),
		HandlerOutPath:    filepath.Join(dir, "handler"),
		RouterOutPath:     filepath.Join(dir, "router"),
	})
	d := groupDomains(list)[0]
	text := routerFuncText(d, cfg)
	if !strings.Contains(text, `Router.Group("/api/v1")`) {
		t.Errorf("路由组应带前缀:\n%s", text)
	}
	if !strings.Contains(text, `thingRouter.GET("/things", thingHandler.ListThings)`) {
		t.Errorf("组内路径应去掉前缀避免重复:\n%s", text)
	}
}

// TestGroupPrefixInSwaggerAndWeb 前缀需同步进 @Router 与前端 url
func TestGroupPrefixInSwaggerAndWeb(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "thing.yaml"), `filename: thing
package: system

ListThings:
  method: GET
  path: /things
`)
	list := mustParseAll(t, yamlDir)
	cfg := mustConfig(config.CodeGenConfig{
		YamlInPath: yamlDir, Module: "m",
		Group:             "/api/v1",
		ServiceIntOutPath: filepath.Join(dir, "svc"),
		ModelOutPath:      filepath.Join(dir, "model"),
		LogicOutPath:      filepath.Join(dir, "logic"),
		HandlerOutPath:    filepath.Join(dir, "handler"),
		RouterOutPath:     filepath.Join(dir, "router"),
	})
	d := groupDomains(list)[0]
	if !strings.Contains(handlerDocBlock(d, d.Apis[0], cfg), "@Router /api/v1/things [get]") {
		t.Errorf("@Router 应带前缀:\n%s", handlerDocBlock(d, d.Apis[0], cfg))
	}
}

// TestGroupPrefixConflictStillDetected 加前缀后冲突仍要能检出,
// 避免"两个领域路径相同但前缀不同"被误判为无冲突
func TestGroupPrefixConflictStillDetected(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "a.yaml"), `filename: alpha
package: system

ListThings:
  method: GET
  path: /things
`)
	mustWrite(filepath.Join(yamlDir, "b.yaml"), `filename: beta
package: system

QueryThings:
  method: GET
  path: /things
`)
	list := mustParseAll(t, yamlDir)
	cfg := mustConfig(config.CodeGenConfig{
		YamlInPath: yamlDir, Module: "m", Group: "/api/v1",
		ServiceIntOutPath: filepath.Join(dir, "svc"),
		ModelOutPath:      filepath.Join(dir, "model"),
		LogicOutPath:      filepath.Join(dir, "logic"),
		HandlerOutPath:    filepath.Join(dir, "handler"),
		RouterOutPath:     filepath.Join(dir, "router"),
	})
	got := validateRoutes(list, cfg)
	if len(got) == 0 {
		t.Error("带前缀时仍应检出重复路由")
	}
}

// TestExistingRouteScanKeepsGroup 已生成的路由文件里带前缀的路径,
// 跨配置检测时要还原成完整路径再比对
func TestExistingRouteScanKeepsGroup(t *testing.T) {
	dir := t.TempDir()
	routerDir := filepath.Join(dir, "router")
	mustWrite(filepath.Join(routerDir, "old_router.go"), `package router

import (
	"github.com/gin-gonic/gin"
	"m/internal/handler"
)

func (s *OldRouter) InitOldRouter(Router *gin.RouterGroup, oldHandler *handler.OldHandler) gin.IRoutes {
	oldRouter := Router.Group("/api/v1")
	{
		oldRouter.GET("/things", oldHandler.ListThings)
	}
	return oldRouter
}
`)
	routes, err := collectExistingRoutes(routerDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 1 {
		t.Fatalf("应提取到 1 条路由, 实际 %d", len(routes))
	}
	if routes[0].path != "/api/v1/things" {
		t.Errorf("路径应还原为完整路径, 实际 %q", routes[0].path)
	}
}

// ---- 辅助 ----

func mkField(name, tag, typ string) *field.Field {
	return &field.Field{
		Name: name, Tag: tag,
		Type: &field.Type{Kind: field.KindPrimitive, Name: typ},
	}
}

func mustParseAll(t *testing.T, dir string) *field.YamlList {
	t.Helper()
	list, err := parseAll(t, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.YamlList) == 0 {
		t.Fatalf("未从 %s 解析到任何 yaml", dir)
	}
	return list
}

var _ = os.Stat
