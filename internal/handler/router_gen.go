package handler

import (
	"path/filepath"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
	"github.com/hishack/go-codegen/internal/merge"
)

// RouterGen 生成 gin 路由文件:{domain}_router.go,结构与参考工程的 router 层一致:
//
//	type UserRouter struct{}
//
//	func (s *UserRouter) InitUserRouter(Router *gin.RouterGroup, userHandler *handler.UserHandler) (R gin.IRoutes) {
//		userRouter := Router.Group("")
//		{
//			userRouter.POST("/users", userHandler.CreateUser)
//		}
//		return userRouter
//	}
//
// 函数体完全由 yaml 推导(生成器全权),路由列表变化时整体替换。
// 不依赖 apiSet/RegisterApi 等项目基建,handler 由参数显式传入,便于手动或 wire 注入
type RouterGen struct {
	cfg *config.CodeGenConfig
}

func NewRouterGen(cfg *config.CodeGenConfig) *RouterGen { return &RouterGen{cfg: cfg} }

func (g *RouterGen) Name() string { return "router" }

func (g *RouterGen) Generate(ctx *GenContext, report *Report) {
	forEachDomain(ctx, report, "router", func(d *domain) error {
		return g.genDomain(d, report)
	})
}

func (g *RouterGen) genDomain(d *domain, report *Report) error {
	path := filepath.Join(g.cfg.RouterOutPath, d.Name+"_router.go")
	f, err := openGoFile(path, g.cfg.Module, pkgNameOfDir(g.cfg.RouterOutPath, "router"), genHeader(g.cfg), report)
	if err != nil {
		return err
	}
	f.NeedImports(g.cfg.Module+"/internal/handler", "github.com/gin-gonic/gin")

	a, err := f.UpsertType("type "+d.routerType()+" struct {}", merge.PolicyIfAbsent)
	if err != nil {
		return err
	}
	report.add(a)
	a, err = f.UpsertFunc(d.routerType(), d.routerInit(), routerFuncText(d), merge.PolicyAlways)
	if err != nil {
		return err
	}
	report.add(a)
	return f.Flush()
}

func routerFuncText(d *domain) string {
	handlerParam := field.LowerFirst(d.pascal()) + "Handler" // userHandler
	groupVar := d.routerGroupVar()                           // userRouter
	var b strings.Builder
	b.WriteString("// " + d.routerInit() + " 注册" + d.pascal() + "模块路由(路由列表由 yaml 生成)\n")
	b.WriteString("func (s *" + d.routerType() + ") " + d.routerInit() +
		"(Router *gin.RouterGroup, " + handlerParam + " *handler." + d.handlerType() + ") (R gin.IRoutes) {\n")
	b.WriteString("\t" + groupVar + " := Router.Group(\"\")\n")
	b.WriteString("\t{\n")
	for _, api := range d.Apis {
		b.WriteString("\t\t" + groupVar + "." + api.Method + "(\"" +
			api.RoutePath() + "\", " + handlerParam + "." + api.ApiName + ")\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("\treturn " + groupVar + "\n")
	b.WriteString("}")
	return b.String()
}
