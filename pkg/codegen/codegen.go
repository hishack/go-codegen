// Package codegen 是 go-codegen 的对外公开入口。
//
// 之所以单独提供本包:生成流程的实现位于 internal/ 下,而 Go 语言规定
// internal 目录只能被其所在模块树内部导入,外部模块无法直接引用。
// 本包把这些实现收敛成一小组稳定的公开 API,供其它工程以库的方式集成:
//
//	import "github.com/hishack/go-codegen/pkg/codegen"
//
//	report, err := codegen.Run("gen.yaml")
//
// 命令行入口 (cmd/gen) 同样只是本包的一层薄封装,保证"库"与"命令"行为一致。
package codegen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/handler"
	"github.com/hishack/go-codegen/internal/parse"
)

// Config 生成器配置。别名导出,便于调用方只依赖本包
type Config = config.CodeGenConfig

// Report 汇总一次生成的结果
type Report struct {
	YamlFiles int      // 解析到的 yaml 文件数
	APIs      int      // 解析到的接口总数
	Added     int      // 新插入的声明数
	Updated   int      // 替换的声明数
	Conflict  int      // 签名冲突、保留用户代码的数量
	Warnings  []string // 需要用户手动处理的问题
}

// String 返回可直接打印的一行摘要
func (r *Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "yaml 文件 %d 个,接口 %d 个", r.YamlFiles, r.APIs)
	fmt.Fprintf(&b, ";新增 %d,更新 %d,签名冲突 %d", r.Added, r.Updated, r.Conflict)
	if len(r.Warnings) > 0 {
		fmt.Fprintf(&b, ";告警 %d 条", len(r.Warnings))
	}
	return b.String()
}

// HasWarnings 是否存在需要人工处理的告警
func (r *Report) HasWarnings() bool { return len(r.Warnings) > 0 }

// LoadConfig 从 yaml 配置文件加载配置。
//
// 配置文件中的相对路径会统一转换为相对"配置文件所在目录"的绝对路径,
// 因此在任意工作目录下调用都能得到一致的结果。
func LoadConfig(cfgPath string) (*Config, error) {
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("配置文件路径非法: %w", err)
	}
	cfg, err := config.LoadCodeGenConfig(abs)
	if err != nil {
		return nil, err
	}
	cfg.YamlInPath = absPath(cfg.YamlInPath, filepath.Dir(abs))
	cfg.ServiceIntOutPath = absPath(cfg.ServiceIntOutPath, filepath.Dir(abs))
	cfg.ModelOutPath = absPath(cfg.ModelOutPath, filepath.Dir(abs))
	cfg.LogicOutPath = absPath(cfg.LogicOutPath, filepath.Dir(abs))
	cfg.HandlerOutPath = absPath(cfg.HandlerOutPath, filepath.Dir(abs))
	cfg.RouterOutPath = absPath(cfg.RouterOutPath, filepath.Dir(abs))
	cfg.WebApiOutPath = absPath(cfg.WebApiOutPath, filepath.Dir(abs))
	return cfg, nil
}

// NewConfig 以代码方式构造配置并做必填校验,随后同样按 baseDir 展开相对路径
func NewConfig(cfg Config, baseDir string) (*Config, error) {
	if baseDir == "" {
		var err error
		if baseDir, err = os.Getwd(); err != nil {
			return nil, err
		}
	}
	c, err := config.NewCodeGenConfig(cfg)
	if err != nil {
		return nil, err
	}
	c.YamlInPath = absPath(c.YamlInPath, baseDir)
	c.ServiceIntOutPath = absPath(c.ServiceIntOutPath, baseDir)
	c.ModelOutPath = absPath(c.ModelOutPath, baseDir)
	c.LogicOutPath = absPath(c.LogicOutPath, baseDir)
	c.HandlerOutPath = absPath(c.HandlerOutPath, baseDir)
	c.RouterOutPath = absPath(c.RouterOutPath, baseDir)
	c.WebApiOutPath = absPath(c.WebApiOutPath, baseDir)
	return c, nil
}

// Run 从配置文件出发执行一次完整生成:加载配置 -> 解析 yaml -> 逐层写入目标工程。
// 生成过程是幂等的,重复执行不会重复插入声明。
func Run(cfgPath string) (*Report, error) {
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	return Generate(cfg)
}

// Generate 使用已加载的配置执行生成,便于把生成器嵌入到其它流程(如 go:generate 或 CI)
func Generate(cfg *Config) (*Report, error) {
	if cfg == nil {
		return nil, errors.New("codegen: 配置不能为空")
	}
	yp, err := parse.NewYamlParse(cfg.YamlInPath)
	if err != nil {
		return nil, err
	}
	yamls, err := yp.ParseAllYaml()
	if err != nil {
		return nil, err
	}
	rep, err := handler.RunAll(cfg, yamls)
	if err != nil {
		return nil, err
	}
	return &Report{
		YamlFiles: len(yamls.YamlList),
		APIs:      len(yamls.AllApis()),
		Added:     rep.Added,
		Updated:   rep.Updated,
		Conflict:  rep.Conflict,
		Warnings:  rep.Warnings,
	}, nil
}

// absPath 以 base 为基准把 p 展开为绝对路径;p 为空时返回空串(表示"跳过该产出")
func absPath(p, base string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(base, p)
}

// ============ 路由冲突自检(供装配层在启动前调用)============

// RouteConflictInfo 一处路由冲突的对外表示
type RouteConflictInfo struct {
	Kind   string // duplicate(路由重复) / wildcard(通配符不同名)
	Detail string // 冲突说明
	First  string // 先注册方
	Second string // 后注册方
}

// ScanRouteConflicts 扫描已生成的路由目录,返回其中存在的冲突。
//
// gin 要求:同一 METHOD+PATH 只能注册一次,且同一层路径段上的通配符必须同名,
// 否则服务启动即 panic。生成器在生成阶段已拦截,这里供 main 等装配层在启动前
// 再兜一层,把手工改坏或历史遗留的冲突变成可读的错误提示。
//
// routerDir 形如 "internal/router";目录不存在时返回空结果与 nil。
func ScanRouteConflicts(routerDir string) ([]RouteConflictInfo, error) {
	cs, err := handler.ScanRouterOutForConflicts(routerDir)
	if err != nil {
		return nil, err
	}
	out := make([]RouteConflictInfo, 0, len(cs))
	for _, c := range cs {
		out = append(out, RouteConflictInfo{
			Kind: c.Kind, Detail: c.Detail, First: c.First, Second: c.Second,
		})
	}
	return out, nil
}

// FormatRouteConflicts 把冲突列表拼成可直接展示的文本
func FormatRouteConflicts(cs []RouteConflictInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "检测到 %d 处路由冲突:", len(cs))
	for i, c := range cs {
		fmt.Fprintf(&b, "\n\n%d) [%s] %s\n    先: %s\n    后: %s",
			i+1, c.Kind, c.Detail, c.First, c.Second)
	}
	return b.String()
}
