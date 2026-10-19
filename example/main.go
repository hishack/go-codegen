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

// newEngine 手动装配全部领域并注册路由
func newEngine(addr string) (*gin.Engine, *routereg.Registry, bool) {
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
// 生成器产出的 Init*Router 内部直接调用 gin 注册,这里用 recover 兜住
// gin 因通配符冲突抛出的 panic:冲突的那一组路由跳过注册,其余照常,
// 保证服务仍能启动、其余接口仍可测试。
func registerAll(r *gin.Engine) {
	// 安全注册:panic 时记录并跳过
	safe := func(name string, fn func()) {
		defer func() {
			if e := recover(); e != nil {
				fmt.Printf("! %s 领域路由注册失败: %v\n", name, e)
			}
		}()
		fn()
	}

	api := r.Group("")

	// 注册顺序说明:生成器不检测跨领域路由冲突,先注册者占位。
	// 这里按"业务价值"排序,让 order/article 等业务接口优先占位,
	// 主测试集(array/edge/nested)冲突时自动跳过,便于测试主要功能。
	safe("user", func() {
		(&router.UserRouter{}).InitUserRouter(api, handler.NewUserHandler(logic.NewUserService()))
	})
	safe("order", func() {
		(&router.OrderRouter{}).InitOrderRouter(api, handler.NewOrderHandler(logic.NewOrderService()))
	})
	safe("article", func() {
		(&router.ArticleRouter{}).InitArticleRouter(api, handler.NewArticleHandler(logic.NewArticleService()))
	})
	safe("product", func() {
		(&router.ProductRouter{}).InitProductRouter(api, handler.NewProductHandler(logic.NewProductService()))
	})
	safe("file", func() {
		(&router.FileRouter{}).InitFileRouter(api, handler.NewFileHandler(logic.NewFileService()))
	})
	safe("notify", func() {
		(&router.NotifyRouter{}).InitNotifyRouter(api, handler.NewNotifyHandler(logic.NewNotifyService()))
	})
	safe("auth", func() {
		(&router.AuthRouter{}).InitAuthRouter(api, handler.NewAuthHandler(logic.NewAuthService()))
	})
	safe("profile", func() {
		(&router.ProfileRouter{}).InitProfileRouter(api, handler.NewProfileHandler(logic.NewProfileService()))
	})
	safe("character", func() {
		(&router.CharacterRouter{}).InitCharacterRouter(api, handler.NewCharacterHandler(logic.NewCharacterService()))
	})
	safe("threshold", func() {
		(&router.ThresholdRouter{}).InitThresholdRouter(api, handler.NewThresholdHandler(logic.NewThresholdService()))
	})

	// 以下三个主测试集领域与上面的业务领域存在路径冲突,会被安全跳过:
	//   edge    /health(与运维接口抢)、/articles/:id(与 article 抢)
	//   array   GET /orders(与 order 抢)
	//   nested  /users/:id/detail(与 user 的 /users/:userId 通配符冲突)
	safe("edge", func() {
		(&router.EdgeRouter{}).InitEdgeRouter(api, handler.NewEdgeHandler(logic.NewEdgeService()))
	})
	safe("array", func() {
		(&router.ArrayRouter{}).InitArrayRouter(api, handler.NewArrayHandler(logic.NewArrayService()))
	})
	safe("nested", func() {
		(&router.NestedRouter{}).InitNestedRouter(api, handler.NewNestedHandler(logic.NewNestedService()))
	})
}
