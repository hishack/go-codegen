// example 是 go-codegen 生成结果的演示工程。
//
// 本文件是手写的手动装配层:把生成出来的 logic / handler / router
// 串成一个可运行的 Gin 服务,用来验证"生成的脚手架能直接跑起来"。
//
// 目前只装配了 user 领域的完整链路(user_logic.go 里补了伪实现);
// 其余领域使用生成器产出的空实现,能编译但返回零值。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"codegen-example/internal/handler"
	"codegen-example/internal/logic"
	"codegen-example/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := flag.String("addr", ":8080", "监听地址")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := newEngine()
	fmt.Printf("example 服务已启动: http://127.0.0.1%s\n", *addr)
	fmt.Println("可试:curl http://127.0.0.1" + *addr + "/users")
	if err := r.Run(*addr); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}

// newEngine 手动装配依赖并注册路由
func newEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// logic -> handler -> router 的标准三段装配
	userService := logic.NewUserService()
	userHandler := handler.NewUserHandler(userService)
	(&router.UserRouter{}).InitUserRouter(r.Group(""), userHandler)

	// 健康检查,便于确认服务是否存活
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
	})

	return r
}
