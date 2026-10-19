// example 是 go-codegen 生成结果的演示工程。
//
// 本文件是手写的装配层:把生成出来的 logic / handler / router
// 串成一个可运行的 Gin 服务,用来验证"生成的脚手架能直接跑起来"。
//
// 注意:路由统一经由 internal/routereg 注册,原因见该包注释——
// 生成的多领域路由会因同层通配符命名不一致导致 gin 启动 panic。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"codegen-example/internal/handler"
	"codegen-example/internal/logic"
	"codegen-example/internal/router"
	"codegen-example/internal/routereg"

	"github.com/hishack/go-codegen/pkg/codegen"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := flag.String("addr", ":8080", "监听地址")
	mode := flag.String("mode", "release", "gin 模式:debug/release/test")
	flag.Parse()

	gin.SetMode(*mode)
	r, reg, hasSwagger := newEngine(*addr)

	fmt.Printf("example 服务已启动: http://127.0.0.1%s (mode=%s)\n", *addr, *mode)
	fmt.Println("健康检查:  curl http://127.0.0.1" + *addr + "/_ops/health")
	fmt.Println("接口清单:  curl http://127.0.0.1" + *addr + "/routes")
	fmt.Println("路由冲突:  curl http://127.0.0.1" + *addr + "/conflicts")
	if hasSwagger {
		fmt.Println("接口文档:  curl http://127.0.0.1" + *addr + "/swagger/index.html")
	}

	// swagger 首页兜底:g Static 占用了 /swagger/*filepath,同层无法再挂路由,
	// 这里用 NoRoute 把 /swagger 根路径导向 index.html
	if hasSwagger {
		r.NoRoute(func(c *gin.Context) {
			if c.Request.URL.Path == "/swagger" || c.Request.URL.Path == "/swagger/" {
				c.File("docs/index.html")
				return
			}
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40400, "message": "路由不存在: " + c.Request.URL.Path,
			})
		})
	}

	// 打印生成器引入的路由冲突,便于评估生成质量
	if cs := reg.Conflicts(); len(cs) > 0 {
		fmt.Printf("\n! 检测到 %d 处路由参数命名冲突(已自动归一化,详见 /conflicts)\n", len(cs))
		seen := map[string]bool{}
		for _, c := range cs {
			k := c.Existing + "|" + c.Introduce
			if seen[k] {
				continue
			}
			seen[k] = true
			fmt.Printf("  同层参数名 '%s' 与 '%s' 不一致,示例: %s\n", c.Existing, c.Introduce, c.Path)
		}
	}

	if err := r.Run(*addr); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}

// precheckRoutes 启动前自检已生成的路由文件。
//
// gin 对路由有两���硬性约束(同 METHOD+PATH 唯一、同层通配符同名),
// 违反会直接 panic 且信息晦涩。生成器已在生成阶段拦截,
// 这里再兜一层:对手工改坏或历史遗留的冲突给出可读的错误提示。
func precheckRoutes() {
	conflicts, err := codegen.ScanRouteConflicts("internal/router")
	if err != nil {
		log.Printf("路由自检失败(忽略): %v", err)
		return
	}
	if len(conflicts) == 0 {
		fmt.Println("路由自检: 通过")
		return
	}
	fmt.Println("\n[路由自检] " + codegen.FormatRouteConflicts(conflicts))
	fmt.Println("\n请修正对应 yaml 后重新生成(go run ./cmd/gen -config ...)")
	log.Fatal("存在路由冲突,已中止启动")
}

// newEngine 手动装配全部领域并注册路由
func newEngine(addr string) (*gin.Engine, *routereg.Registry, bool) {
	precheckRoutes() // 冲突时直接终止,避免 gin panic
	r := gin.New()
	r.Use(gin.Recovery())

	reg := routereg.New(r)
	registerAll(r)

	// ---- 运维辅助接口 ----
	r.GET("/_ops/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
	})

	r.GET("/routes", func(c *gin.Context) {
		type route struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		list := make([]route, 0, 64)
		for _, ri := range r.Routes() {
			switch {
			case ri.Path == "/_ops/health", ri.Path == "/routes", ri.Path == "/conflicts",
				strings.HasPrefix(ri.Path, "/swagger"):
				continue
			}
			list = append(list, route{Method: ri.Method, Path: ri.Path})
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "total": len(list), "data": list})
	})

	// /conflicts 暴露生成器引入的路由冲突,供测试脚本断言
	r.GET("/conflicts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":     0,
			"total":    len(reg.Conflicts()),
			"data":     reg.Conflicts(),
			"rewrites": reg.Rewrites(),
		})
	})

	// swagger 文档目录(由 swag init 生成 docs/ 后可访问)
	// 注意:gin 的 Static 会在 /swagger 下注册 /*filepath 通配路由,该层
	// 不能再挂任何静态或通配路由,否则启动即 panic。因此首页改由 NoRoute 兜底。
	_, docsErr := os.Stat("docs/swagger.json")
	hasSwagger := docsErr == nil
	if hasSwagger {
		r.Static("/swagger", "./docs")
		fmt.Println("swagger 文档: 已挂载 /swagger/(swag init 产物)")
	} else {
		fmt.Println("swagger 文档: 未找到 docs/,请先执行  swag init -g main.go -o ./docs")
	}

	return r, reg, hasSwagger
}

// registerAll 逐个装配业务领域。
//
// 生成器已在生成阶段拦截路由冲突(同 METHOD+PATH 重复、同层通配符不同名),
// 因此这里可以放心直接注册;启动前还会再自检一次已生成的路由文件,
// 万一手工改坏了也能给出明确提示而不是 panic。
func registerAll(r *gin.Engine) {
	// 注册即视为成功;若仍发生冲突,由 GinPrecheckRouteConflicts 提前拦截
	safe := func(fn func()) { fn() }

	api := r.Group("")

	safe(func() {
		(&router.UserRouter{}).InitUserRouter(api, handler.NewUserHandler(logic.NewUserService()))
	})
	safe(func() {
		(&router.OrderRouter{}).InitOrderRouter(api, handler.NewOrderHandler(logic.NewOrderService()))
	})
	safe(func() {
		(&router.ArticleRouter{}).InitArticleRouter(api, handler.NewArticleHandler(logic.NewArticleService()))
	})
	safe(func() {
		(&router.ProductRouter{}).InitProductRouter(api, handler.NewProductHandler(logic.NewProductService()))
	})
	safe(func() {
		(&router.FileRouter{}).InitFileRouter(api, handler.NewFileHandler(logic.NewFileService()))
	})
	safe(func() {
		(&router.NotifyRouter{}).InitNotifyRouter(api, handler.NewNotifyHandler(logic.NewNotifyService()))
	})
	safe(func() {
		(&router.AuthRouter{}).InitAuthRouter(api, handler.NewAuthHandler(logic.NewAuthService()))
	})
	safe(func() {
		(&router.ProfileRouter{}).InitProfileRouter(api, handler.NewProfileHandler(logic.NewProfileService()))
	})
	safe(func() {
		(&router.CharacterRouter{}).InitCharacterRouter(api, handler.NewCharacterHandler(logic.NewCharacterService()))
	})
	safe(func() {
		(&router.ThresholdRouter{}).InitThresholdRouter(api, handler.NewThresholdHandler(logic.NewThresholdService()))
	})
	safe(func() {
		(&router.EdgeRouter{}).InitEdgeRouter(api, handler.NewEdgeHandler(logic.NewEdgeService()))
	})
	safe(func() {
		(&router.ArrayRouter{}).InitArrayRouter(api, handler.NewArrayHandler(logic.NewArrayService()))
	})
	safe(func() {
		(&router.NestedRouter{}).InitNestedRouter(api, handler.NewNestedHandler(logic.NewNestedService()))
	})
}
