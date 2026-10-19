// example/cmd/gen 是示例工程侧的生成入口。
//
// 它演示了"以库的方式引用 go-codegen":本模块并不 fork 生成器代码,
// 而是在 go.mod 中声明依赖后,直接 import 公开门面包
// github.com/hishack/go-codegen/pkg/codegen 调用同一套生成流程。
//
// 用法:
//
//	cd example
//	go run ./cmd/gen                 # 使用默认配置 ../gen.yaml
//	go run ./cmd/gen -config ../gen.yaml
//
// 也可以用 go generate 触发(等价于上面第一条命令):
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

//go:generate go run . -config ../gen.yaml

func main() {
	cfgPath := flag.String("config", "../gen.yaml", "生成器配置文件路径")
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
