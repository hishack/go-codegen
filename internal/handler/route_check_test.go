package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/parse"
)

// 路由冲突检测的单元测试。
//
// 这类校验的价值在于:没有它,生成的代码能编译通过,却会在服务启动时 panic。
// 因此必须覆盖 gin 的两条硬性约束:
//  1. 同一 METHOD + PATH 只能注册一次
//  2. 同一层路径段上的通配符名字必须一致

// ---- 测试用 yaml ----

const dupYamlA = `filename: alpha
package: system

ListThings:
  method: GET
  path: /things
`

const dupYamlB = `filename: beta
package: system

QueryThings:
  method: GET
  path: /things
`

const wildcardYamlA = `filename: gamma
package: system

GetUserDetail:
  method: GET
  path: /users/:userId(int64)/detail
`

const wildcardYamlB = `filename: delta
package: system

GetUser:
  method: GET
  path: /users/:id(int64)
`

// parseTo 解析一组 yaml 内容为 YamlList
func parseTo(t *testing.T, files map[string]string) *fieldYamlList {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		mustWrite(filepath.Join(dir, name), content)
	}
	list, err := parseAll(t, dir)
	if err != nil {
		t.Fatalf("解析 yaml 失败: %v", err)
	}
	return list
}

// TestValidateRoutesNoConflict 无冲突时应返回空
func TestValidateRoutesNoConflict(t *testing.T) {
	list := parseTo(t, map[string]string{
		"a.yaml": "filename: alpha\npackage: system\n\nGetAlpha:\n  method: GET\n  path: /alpha\n",
		"b.yaml": "filename: beta\npackage: system\n\nGetBeta:\n  method: POST\n  path: /beta\n",
	})
	if got := validateRoutes(list); len(got) != 0 {
		t.Errorf("期望无冲突,实际 %d 处:\n%s", len(got), formatConflicts(got))
	}
}

// TestValidateRoutesDuplicate 同 METHOD+PATH 重复必须报错
func TestValidateRoutesDuplicate(t *testing.T) {
	list := parseTo(t, map[string]string{
		"alpha.yaml": dupYamlA,
		"beta.yaml":  dupYamlB,
	})
	got := validateRoutes(list)
	if len(got) == 0 {
		t.Fatal("期望检出 GET /things 重复")
	}
	c := got[0]
	if c.Kind != "duplicate" {
		t.Errorf("冲突类型 = %q,期望 duplicate", c.Kind)
	}
	if !strings.Contains(c.Detail, "GET /things") {
		t.Errorf("冲突说明应含具体路由,实际: %s", c.Detail)
	}
	// 应指出是哪两个接口冲突
	if !strings.Contains(c.First, "ListThings") || !strings.Contains(c.Second, "QueryThings") {
		t.Errorf("应指出冲突双方,实际:\n  先: %s\n  后: %s", c.First, c.Second)
	}
	// 应带来源文件,便于定位
	if !strings.Contains(c.First, "alpha") || !strings.Contains(c.Second, "beta") {
		t.Errorf("应指出来源 yaml,实际:\n  先: %s\n  后: %s", c.First, c.Second)
	}
}

// TestValidateRoutesDuplicateSameFile 同一文件内重复也要报
func TestValidateRoutesDuplicateSameFile(t *testing.T) {
	list := parseTo(t, map[string]string{
		"alpha.yaml": `filename: alpha
package: system

ListThings:
  method: GET
  path: /things

QueryThings:
  method: GET
  path: /things
`,
	})
	if got := validateRoutes(list); len(got) == 0 {
		t.Fatal("同一文件内的重复路由也应被检出")
	}
}

// TestValidateRoutesWildcard 同层通配符命名不一致必须报错
func TestValidateRoutesWildcard(t *testing.T) {
	list := parseTo(t, map[string]string{
		"gamma.yaml": wildcardYamlA,
		"delta.yaml": wildcardYamlB,
	})
	got := validateRoutes(list)
	if len(got) == 0 {
		t.Fatal("期望检出 /users 下通配符命名冲突")
	}
	c := got[0]
	if c.Kind != "wildcard" {
		t.Errorf("冲突类型 = %q,期望 wildcard", c.Kind)
	}
	if !strings.Contains(c.Detail, ":userId") || !strings.Contains(c.Detail, ":id") {
		t.Errorf("冲突说明应含两个参数名,实际: %s", c.Detail)
	}
}

// TestValidateRoutesDifferentMethodSamePath 不同方法可共存
func TestValidateRoutesDifferentMethodSamePath(t *testing.T) {
	list := parseTo(t, map[string]string{
		"alpha.yaml": "filename: alpha\npackage: system\n\nGetThings:\n  method: GET\n  path: /things\n",
		"beta.yaml":  "filename: beta\npackage: system\n\nCreateThing:\n  method: POST\n  path: /things\n",
	})
	if got := validateRoutes(list); len(got) != 0 {
		t.Errorf("不同 METHOD 的同一路径不应算冲突,实际:\n%s", formatConflicts(got))
	}
}

// TestValidateRoutesWildcardDifferentLevel 不同前缀下的同名参数不冲突
func TestValidateRoutesWildcardDifferentLevel(t *testing.T) {
	list := parseTo(t, map[string]string{
		"a.yaml": "filename: alpha\npackage: system\n\nGetUser:\n  method: GET\n  path: /users/:userId(int64)\n",
		"b.yaml": "filename: beta\npackage: system\n\nGetOrder:\n  method: GET\n  path: /orders/:orderId(int64)\n",
	})
	if got := validateRoutes(list); len(got) != 0 {
		t.Errorf("不同前缀下的通配符不应算冲突,实际:\n%s", formatConflicts(got))
	}
}

// TestValidateRoutesNestedSameName 嵌套层同名参数应视为合法
func TestValidateRoutesNestedSameName(t *testing.T) {
	list := parseTo(t, map[string]string{
		"a.yaml": "filename: alpha\npackage: system\n\nGetUser:\n  method: GET\n  path: /users/:userId(int64)\n",
		"b.yaml": "filename: beta\npackage: system\n\nGetUserOrder:\n  method: GET\n  path: /users/:userId(int64)/orders/:orderId(int64)\n",
	})
	if got := validateRoutes(list); len(got) != 0 {
		t.Errorf("嵌套层同名参数应合法,实际:\n%s", formatConflicts(got))
	}
}

// ---- 跨配置(已生成文件)检测 ----

// 造一个已生成路由文件的目录
func writeRouterFile(t *testing.T, dir, domain, body string) {
	t.Helper()
	content := `package router

import (
	"app/internal/handler"

	"github.com/gin-gonic/gin"
)

type ` + strings.ToUpper(domain[:1]) + domain[1:] + `Router struct{}

func (s *` + strings.ToUpper(domain[:1]) + domain[1:] + `Router) Init` + strings.ToUpper(domain[:1]) +
		domain[1:] + `Router(Router *gin.RouterGroup, h *handler.` + strings.ToUpper(domain[:1]) +
		domain[1:] + `Handler) (R gin.IRoutes) {
	r := Router.Group("")
	{
` + body + `
	}
	return r
}
`
	mustWrite(filepath.Join(dir, domain+"_router.go"), content)
}

// TestScanRouterOutDetectsDuplicate 已生成文件之间的重复应被检出
func TestScanRouterOutDetectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeRouterFile(t, dir, "alpha", "\t\tr.GET(\"/things\", h.ListThings)")
	writeRouterFile(t, dir, "beta", "\t\tr.GET(\"/things\", h.QueryThings)")

	got, err := ScanRouterOutForConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("期望检出已生成文件间的 GET /things 重复")
	}
	if got[0].Kind != "duplicate" {
		t.Errorf("类型 = %q,期望 duplicate", got[0].Kind)
	}
}

// TestScanRouterOutDetectsWildcard 已生成文件之间的通配符冲突应被检出
func TestScanRouterOutDetectsWildcard(t *testing.T) {
	dir := t.TempDir()
	writeRouterFile(t, dir, "gamma", "\t\tr.GET(\"/users/:userId/detail\", h.GetUserDetail)")
	writeRouterFile(t, dir, "delta", "\t\tr.GET(\"/users/:id\", h.GetUser)")

	got, err := ScanRouterOutForConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("期望检出 /users 下通配符命名冲突")
	}
	if got[0].Kind != "wildcard" {
		t.Errorf("类型 = %q,期望 wildcard", got[0].Kind)
	}
}

// TestScanRouterOutNoConflict 干净目录不应误报
func TestScanRouterOutNoConflict(t *testing.T) {
	dir := t.TempDir()
	writeRouterFile(t, dir, "alpha", "\t\tr.GET(\"/things\", h.ListThings)")
	writeRouterFile(t, dir, "beta", "\t\tr.POST(\"/things\", h.CreateThing)\n\t\tr.GET(\"/others/:id\", h.GetOther)")

	got, err := ScanRouterOutForConflicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("无冲突时应返回空,实际 %d 处:\n%s", len(got), formatConflicts(got))
	}
}

// TestScanRouterOutEmptyDir 目录不存在时应返回空且不报错
func TestScanRouterOutEmptyDir(t *testing.T) {
	got, err := ScanRouterOutForConflicts(filepath.Join(t.TempDir(), "not-exist"))
	if err != nil {
		t.Fatalf("目录不存在不应报错: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("目录不存在时应返回空,实际 %d 处", len(got))
	}
}

// TestRunAllAbortsOnConflict RunAll 必须在生成前拦下冲突,不产出坏代码
func TestRunAllAbortsOnConflict(t *testing.T) {
	dir := t.TempDir()
	yamlDir := filepath.Join(dir, "yaml")
	mustWrite(filepath.Join(yamlDir, "alpha.yaml"), dupYamlA)
	mustWrite(filepath.Join(yamlDir, "beta.yaml"), dupYamlB)

	cfg := mustConfig(config.CodeGenConfig{
		YamlInPath:        yamlDir,
		Module:            "app",
		ServiceIntOutPath: filepath.Join(dir, "out", "service"),
		ModelOutPath:      filepath.Join(dir, "out", "model"),
		LogicOutPath:      filepath.Join(dir, "out", "logic"),
		HandlerOutPath:    filepath.Join(dir, "out", "handler"),
		RouterOutPath:     filepath.Join(dir, "out", "router"),
	})

	list, err := parseAll(t, yamlDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RunAll(cfg, list); err == nil {
		t.Fatal("RunAll 应因路由冲突而报错")
	} else if !strings.Contains(err.Error(), "路由冲突") {
		t.Errorf("错误信息应说明是路由冲突,实际: %v", err)
	}

	// 关键:冲突时不应写出任何文件
	if _, err := os.Stat(cfg.RouterOutPath); err == nil {
		t.Error("冲突时应中止生成,不应产出路由文件")
	}
}

// TestRunAllCrossConfigConflict 第二次生成与已有文件冲突时也要拦住
func TestRunAllCrossConfigConflict(t *testing.T) {
	dir := t.TempDir()
	routerOut := filepath.Join(dir, "out", "router")
	// 预置一个已生成的路由文件,含 GET /things
	writeRouterFile(t, routerOut, "alpha", "\t\tr.GET(\"/things\", h.ListThings)")

	yamlDir := filepath.Join(dir, "yaml")
	// 本次要生成 beta 领域,同样使用 GET /things
	mustWrite(filepath.Join(yamlDir, "beta.yaml"), dupYamlB)

	cfg := mustConfig(config.CodeGenConfig{
		YamlInPath:        yamlDir,
		Module:            "app",
		ServiceIntOutPath: filepath.Join(dir, "out", "service"),
		ModelOutPath:      filepath.Join(dir, "out", "model"),
		LogicOutPath:      filepath.Join(dir, "out", "logic"),
		HandlerOutPath:    filepath.Join(dir, "out", "handler"),
		RouterOutPath:     routerOut,
	})

	list, err := parseAll(t, yamlDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RunAll(cfg, list); err == nil {
		t.Fatal("跨配置冲突也应被 RunAll 拦下")
	}
	// 预置文件不应被改动
	data, _ := os.ReadFile(filepath.Join(routerOut, "alpha_router.go"))
	if strings.Contains(string(data), "beta") {
		t.Error("冲突中止时不应修改已有文件")
	}
}

// fieldYamlList 是 field.YamlList 的本地别名,便于测试签名书写
type fieldYamlList = field.YamlList

// parseAll 解析目录下的全部 yaml
func parseAll(t *testing.T, dir string) (*fieldYamlList, error) {
	t.Helper()
	yp, err := parse.NewYamlParse(dir)
	if err != nil {
		return nil, err
	}
	return yp.ParseAllYaml()
}
