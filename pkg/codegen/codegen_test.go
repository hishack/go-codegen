package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile 写文件并自动创建父目录
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败 %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写文件失败 %s: %v", path, err)
	}
}

// newFixture 造一个最小可用的生成环境:yaml 输入目录 + 配置文件,输出目录在 out 下
func newFixture(t *testing.T) (root, cfgPath string) {
	t.Helper()
	root = t.TempDir()

	writeFile(t, filepath.Join(root, "yaml", "user.yaml"), `filename: user
package: system

CreateUser:
  method: POST
  path: /users
  req:
    name: string
    age: int
  res:
    id: int64
    createdAt: time.Time

ListUsers:
  method: GET
  path: /users
  res:
    total: int64
    list:
      - id: int64
        name: string
`)

	cfgPath = filepath.Join(root, "gen.yaml")
	writeFile(t, cfgPath, `
yaml: ./yaml
module: demo
serviceOut: ./out/service
modelOut: ./out/model
logicOut: ./out/logic
handlerOut: ./out/handler
routerOut: ./out/router
webOut: ./out/web
author: tester
`)
	return root, cfgPath
}

// TestRunGeneratesFromCleanDir 从零生成一次,确认各层产物落盘
func TestRunGeneratesFromCleanDir(t *testing.T) {
	_, cfgPath := newFixture(t)

	report, err := Run(cfgPath)
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	if report.YamlFiles != 1 {
		t.Errorf("yaml 文件数 = %d, 期望 1", report.YamlFiles)
	}
	if report.APIs != 2 {
		t.Errorf("接口数 = %d, 期望 2", report.APIs)
	}
	if report.Added == 0 {
		t.Error("期望至少新增若干声明,实际为 0")
	}

	root := filepath.Dir(cfgPath)
	for _, rel := range []string{
		"out/service/user_service.go",
		"out/model/request/user.go",
		"out/logic/user_logic.go",
		"out/handler/user_handler.go",
		"out/router/user_router.go",
		"out/web/user.ts",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("期望产物存在 %s: %v", rel, err)
		}
	}
}

// TestRunIsIdempotent 重复执行不应产生新的插入,保证可安全地反复运行
func TestRunIsIdempotent(t *testing.T) {
	_, cfgPath := newFixture(t)

	first, err := Run(cfgPath)
	if err != nil {
		t.Fatalf("首次 Run 失败: %v", err)
	}
	second, err := Run(cfgPath)
	if err != nil {
		t.Fatalf("二次 Run 失败: %v", err)
	}
	if second.Added != 0 {
		t.Errorf("二次运行新增 = %d, 期望 0(不幂等)", second.Added)
	}
	if second.YamlFiles != first.YamlFiles || second.APIs != first.APIs {
		t.Errorf("二次运行统计与首次不一致: %+v vs %+v", second, first)
	}
}

// TestPathsResolveRelativeToConfigFile 是本次重构的关键约定:
// 配置文件中的相对路径按"配置文件所在目录"解析,因此在任意工作目录下调用结果一致
func TestPathsResolveRelativeToConfigFile(t *testing.T) {
	root, cfgPath := newFixture(t)

	// 先在根目录下生成一次,产物落在 root/out
	if _, err := Run(cfgPath); err != nil {
		t.Fatalf("Run 失败: %v", err)
	}

	// 换一个完全不同的工作目录再跑,产物仍应落在 root/out,而不是 cwd/out
	other := t.TempDir()
	t.Chdir(other)
	if _, err := Run(cfgPath); err != nil {
		t.Fatalf("切换工作目录后 Run 失败: %v", err)
	}

	if _, err := os.Stat(filepath.Join(other, "out")); err == nil {
		t.Error("产物不应写到当前工作目录下")
	}
	if _, err := os.Stat(filepath.Join(root, "out", "service", "user_service.go")); err != nil {
		t.Errorf("产物应仍在配置文件同级 out 目录: %v", err)
	}
}

// TestConfigPathsAreAbsolute LoadConfig 之后所有路径都应是绝对路径
func TestConfigPathsAreAbsolute(t *testing.T) {
	root, cfgPath := newFixture(t)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig 失败: %v", err)
	}
	fields := map[string]string{
		"YamlInPath":        cfg.YamlInPath,
		"ServiceIntOutPath": cfg.ServiceIntOutPath,
		"ModelOutPath":      cfg.ModelOutPath,
		"LogicOutPath":      cfg.LogicOutPath,
		"HandlerOutPath":    cfg.HandlerOutPath,
		"RouterOutPath":     cfg.RouterOutPath,
		"WebApiOutPath":     cfg.WebApiOutPath,
	}
	for name, v := range fields {
		if v == "" {
			t.Errorf("%s 不应为空", name)
			continue
		}
		if !filepath.IsAbs(v) {
			t.Errorf("%s = %q, 期望绝对路径", name, v)
		}
		if !strings.HasPrefix(v, root) {
			t.Errorf("%s = %q, 期望位于 %s 下", name, v, root)
		}
	}
}

// TestLoadConfigMissingFile 配置不存在时应返回可读的错误而不是 panic
func TestLoadConfigMissingFile(t *testing.T) {
	if _, err := Run(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Error("期望配置文件缺失时报错")
	}
}

// TestGenerateNilConfig 传入 nil 配置应报错
func TestGenerateNilConfig(t *testing.T) {
	if _, err := Generate(nil); err == nil {
		t.Error("期望 nil 配置时报错")
	}
}

// TestReportString Report 的摘要应包含关键统计
func TestReportString(t *testing.T) {
	r := &Report{YamlFiles: 4, APIs: 28, Added: 127, Warnings: []string{"x"}}
	s := r.String()
	for _, want := range []string{"4", "28", "127"} {
		if !strings.Contains(s, want) {
			t.Errorf("摘要 %q 应包含 %q", s, want)
		}
	}
	if !r.HasWarnings() {
		t.Error("存在告警时 HasWarnings 应为 true")
	}
	if (&Report{}).HasWarnings() {
		t.Error("无告警时 HasWarnings 应为 false")
	}
}
