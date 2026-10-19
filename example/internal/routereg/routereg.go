// Package routereg 提供路由注册的统一入口。
//
// 背景:gin 的路由树要求"同一层路径段上的通配符名字必须一致"。
// 而代码生成器允许不同领域各自定义路径参数名,例如:
//
//	user_router.go    注册 /users/:userId
//	nested_router.go  注册 /users/:id/detail
//
// 两者在同一层混用不同名字,gin 在启动时会直接 panic:
//
//	':id' in new path '/users/:id/detail' conflicts with existing wildcard ':userId'
//
// 本包在注册前把同层通配符统一改名,并把真实的参数名记录下来,
// 使生成的多领域路由可以共存。冲突详情可通过 Conflicts() 查询。
package routereg

import (
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// Conflict 记录一次路径段命名冲突
type Conflict struct {
	Path      string // 已注册的路径
	NewPath   string // 引发冲突的新路径
	Existing  string // 已存在的参数名
	Introduce string // 引发冲突的参数名
}

// Registry 路由注册器:归一化参数名后再交给 gin
type Registry struct {
	engine *gin.Engine
	group  *gin.RouterGroup
	// segNames[绝对路径的静态前缀] = 该层已使用的参数名
	segNames map[string]string
	// rewrites 记录 "注册时路径 -> 实际注册到 gin 的路径"
	rewrites  []string
	conflicts []Conflict
}

// paramRe 匹配路径段里的 :param
var paramRe = regexp.MustCompile(`^:([A-Za-z_][A-Za-z0-9_]*)$`)

// New 创建注册器
func New(engine *gin.Engine) *Registry {
	return &Registry{
		engine:   engine,
		group:    engine.Group(""),
		segNames: map[string]string{},
	}
}

// Group 取一个子分组(本示例未使用,保留给后续按前缀分组)
func (r *Registry) Group(prefix string) *gin.RouterGroup {
	return r.engine.Group(prefix)
}

// Handle 按方法注册一个 handler
func (r *Registry) Handle(method, path string, h gin.HandlerFunc) {
	norm := r.normalize(path)
	switch strings.ToUpper(method) {
	case "GET":
		r.group.GET(norm, h)
	case "POST":
		r.group.POST(norm, h)
	case "PUT":
		r.group.PUT(norm, h)
	case "PATCH":
		r.group.PATCH(norm, h)
	case "DELETE":
		r.group.DELETE(norm, h)
	default:
		r.group.Handle(strings.ToUpper(method), norm, h)
	}
	if norm != path {
		r.rewrites = append(r.rewrites, path+"  ->  "+norm)
	}
}

// normalize 逐段扫描路径,把通配符统一改名为该层的第一个名字。
// gin 的路由树按"静态前缀"分组,同一前缀下的 :param 必须同名,
// 否则启动即 panic。
func (r *Registry) normalize(path string) string {
	segs := strings.Split(path, "/")
	prefix := "" // 当前累计的静态前缀
	for i, s := range segs {
		m := paramRe.FindStringSubmatch(s)
		if m == nil {
			// 静态段:并入前缀。注意 gin 按整段匹配,这里用该段自身作为 key
			if s != "" {
				prefix += "/" + s
			}
			continue
		}
		name := m[1]
		if exist, ok := r.segNames[prefix]; ok {
			if exist != name {
				r.conflicts = append(r.conflicts, Conflict{
					Path: path, NewPath: path, Existing: exist, Introduce: name,
				})
				segs[i] = ":" + exist
			}
			continue
		}
		r.segNames[prefix] = name
	}
	return strings.Join(segs, "/")
}

// Conflicts 返回已记录的命名冲突
func (r *Registry) Conflicts() []Conflict {
	out := append([]Conflict(nil), r.conflicts...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Rewrites 返回因归一化而被改写的路径清单
func (r *Registry) Rewrites() []string {
	return append([]string(nil), r.rewrites...)
}
