// example/cmd/gen 是示例工程侧的生成入口。
//
// 它演示了"以库的方式引用 go-codegen":本模块并不 fork 生成器代码,
// 而是在 go.mod 中声明依赖后,直接 import 公开门面包
// github.com/hishack/go-codegen/pkg/codegen 调用同一套生成流程。
//
// 用法:
//
//	cd example
//	go run ./cmd/gen -config cmd/gen/gen.yaml   # 读本目录的 yaml/ 示例
//	go run ./cmd/gen -config ../gen.yaml        # 读仓库根的 yamlPath/ 主测试集
//
// 也可以用 go generate 触发:
//
//	go generate ./...
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hishack/go-codegen/pkg/codegen"
)

func main() {
	cfgPath := flag.String("config", "cmd/gen/gen.yaml", "生成器配置文件路径")
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

	fmt.Printf("✔ 示例工程生成完成:%s\n", report)
	if report.HasWarnings() {
		fmt.Println("  注意事项:")
		for _, w := range report.Warnings {
			fmt.Printf("    - %s\n", strings.TrimSpace(w))
		}
	}
	return nil
}
