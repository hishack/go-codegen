// go-codegen 命令行入口。
//
// 用法:
//
//	go run github.com/hishack/go-codegen/cmd/gen -config gen.yaml
//
// 流程:读配置 -> 解析 yaml 目录(带注释) -> 按生成器依次增量写入目标工程。
// 本文件只是 pkg/codegen 的一层薄封装,保证"库"与"命令"行为完全一致,
// 第三方工程既可以 go run 本命令,也可以直接 import pkg/codegen 嵌入自己的流程。
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hishack/go-codegen/pkg/codegen"
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
	report, err := codegen.Run(cfgPath)
	if err != nil {
		return err
	}

	fmt.Printf("✔ 生成完成:%s\n", report)
	if report.HasWarnings() {
		fmt.Println("  注意事项:")
		for _, w := range report.Warnings {
			fmt.Printf("    - %s\n", strings.TrimSpace(w))
		}
	}
	return nil
}
