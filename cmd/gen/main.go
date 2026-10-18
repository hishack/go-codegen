// go-codegen 命令行入口。
//
// 用法:
//
//	go run ./cmd/gen -config gen.yaml
//
// 流程:读配置 -> 解析 yaml 目录(带注释) -> 按生成器依次增量写入目标工程
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/handler"
	"github.com/hishack/go-codegen/internal/parse"
)

func main() {
	cfgPath := flag.String("config", "gen.yaml", "配置文件路径")
	flag.Parse()

	if err := run(*cfgPath); err != nil {
		fmt.Fprintln(os.Stderr, "生成失败:", err)
		os.Exit(1)
	}
}

func run(cfgPath string) error {
	cfg, err := config.LoadCodeGenConfig(cfgPath)
	if err != nil {
		return err
	}

	yp, err := parse.NewYamlParse(cfg.YamlInPath)
	if err != nil {
		return err
	}
	yamls, err := yp.ParseAllYaml()
	if err != nil {
		return err
	}

	report, err := handler.RunAll(cfg, yamls)
	if err != nil {
		return err
	}

	fmt.Printf("✔ 生成完成:yaml 文件 %d 个,接口 %d 个\n", len(yamls.YamlList), len(yamls.AllApis()))
	fmt.Printf("  新增声明 %d,更新 %d,签名冲突 %d\n", report.Added, report.Updated, report.Conflict)
	if len(report.Warnings) > 0 {
		fmt.Println("  注意事项:")
		for _, w := range report.Warnings {
			fmt.Printf("    - %s\n", strings.TrimSpace(w))
		}
	}
	_ = filepath.Base // 保持 import
	return nil
}
