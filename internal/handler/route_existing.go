package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/field"
)

// 跨配置路由冲突检测。
//
// 背景:一个目标工程可能由多份配置先后生成(例如仓库主测试集 + 各自领域示例),
// 每份配置单独跑时互相看不到对方的路由,于是各自都能生成成功,
// 但装配成同一个服务时 gin 仍会 panic。
//
// 因此在生成前还要扫描目标工程里**已存在**的路由文件,
// 把它们与本次要生成的路由一起做冲突检测。

// routeCallRe 匹配生成代码里的路由注册语句,例如:
//
//	userRouter.GET("/users/:userId", userHandler.GetUser)
//	edgeRouter.POST("/logout", edgeHandler.Logout)
var routeCallRe = regexp.MustCompile(
	`(?m)^\s*\w+\.(GET|POST|PUT|PATCH|DELETE)\("([^"]*)"\s*,\s*\w+\.(\w+)\)`)

// groupCallRe 匹配路由文件里的分组声明,例如:
//
//	userRouter := Router.Group("/api/v1")
var groupCallRe = regexp.MustCompile(`(?m)^\s*\w+\s*:=\s*Router\.Group\("([^"]*)"\)`)

// existingRoute 已存在的路由及其来源文件
type existingRoute struct {
	method string
	path   string
	domain string // 来自 {domain}_router.go
	file   string // 完整文件名
	api    string // handler 方法名
}

// collectExistingRoutes 扫描 routerOut 目录下已存在的 *_router.go,提取其中注册的路由。
// 本次将要重新生成的领域会被排除,避免自己跟自己比。
func collectExistingRoutes(routerOut string, skipDomains map[string]bool) ([]existingRoute, error) {
	entries, err := os.ReadDir(routerOut)
	if err != nil {
		// 目录还不存在说明是首次生成,没有历史路由
		return nil, nil
	}
	var out []existingRoute
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_router.go") {
			continue
		}
		domain := strings.TrimSuffix(e.Name(), "_router.go")
		if skipDomains[domain] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(routerOut, e.Name()))
		if err != nil {
			return nil, err
		}
		// 取出该文件的分组前缀,使路径还原成注册到 gin 上的完整路径
		prefix := ""
		if m := groupCallRe.FindStringSubmatch(string(data)); m != nil {
			prefix = config.NormalizeGroup(m[1])
		}
		for _, m := range routeCallRe.FindAllStringSubmatch(string(data), -1) {
			out = append(out, existingRoute{
				method: m[1],
				path:   joinGroup(prefix, m[2]),
				domain: domain,
				file:   e.Name(),
				api:    m[3],
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].method != out[j].method {
			return out[i].method < out[j].method
		}
		return out[i].path < out[j].path
	})
	return out, nil
}

// describeExisting 生成 "GET /orders (order_router.go 的 ListOrders)" 形式的来源描述
func (e existingRoute) describe() string {
	return fmt.Sprintf("%s %s (%s 的 %s)", e.method, e.path, e.domain, e.api)
}

// joinGroup 把分组前缀与组内路径拼成完整路径
func joinGroup(prefix, path string) string {
	if prefix == "" {
		return path
	}
	if path == "" || path == "/" {
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return prefix + path
}

// checkRoutesAgainstExisting 把本次生成的路由与目标工程里已存在的路由做冲突检测。
//
// 这是跨配置场景下的关键防线:两次独立的生成各自都合法,
// 合到一起才冲突,只有把历史路由一起纳入检查才能提前发现。
func checkRoutesAgainstExisting(yamls *field.YamlList, cfg *config.CodeGenConfig) []RouteConflict {
	// 本次会重新生成的领域,其历史路由会被覆盖,无需参与"历史 vs 历史"的比对
	skip := map[string]bool{}
	for _, d := range groupDomains(yamls) {
		skip[d.Name] = true
	}

	all, err := collectExistingRoutes(cfg.RouterOutPath, nil)
	if err != nil {
		return nil
	}
	// 历史路由里,属于本次会重新生成的领域先摘掉
	var existing []existingRoute
	for _, e := range all {
		if !skip[e.domain] {
			existing = append(existing, e)
		}
	}

	incoming := collectRoutes(yamls, cfg)
	var conflicts []RouteConflict

	// ---- 规则 0:历史路由自身就有冲突 ----
	// 这些冲突早已存在(比如上一版生成时就撞了),本次生成若不检查就会一直带病。
	conflicts = append(conflicts, conflictsAmong(existing)...)

	if len(existing) == 0 {
		return conflicts
	}

	// ---- 规则 1:METHOD + PATH 重复(本次 vs 已存在)----
	byRoute := map[string][]existingRoute{}
	for _, e := range existing {
		byRoute[e.method+" "+e.path] = append(byRoute[e.method+" "+e.path], e)
	}
	for _, in := range incoming {
		key := in.method + " " + in.path
		for _, ex := range byRoute[key] {
			conflicts = append(conflicts, RouteConflict{
				Kind: "duplicate",
				Detail: fmt.Sprintf(
					"路由 %s 与目标工程中已存在的路由重复(来自 %s),gin 会 panic:"+
						"handlers are already registered for path '%s'",
					key, ex.file, in.path),
				First:  ex.describe(),
				Second: in.describe(),
			})
		}
	}

	// ---- 规则 2:同层通配符命名不一致(本次 vs 已存在)----
	type owner struct {
		name  string
		entry existingRoute
	}
	existingOwners := map[string]map[string]owner{}
	firstOf := map[string]existingRoute{} // 记录每个前缀首次出现的路由,便于报错
	for _, e := range existing {
		for _, seg := range walkParams(e.path) {
			prefix, name := seg[0], seg[1]
			owners, ok := existingOwners[prefix]
			if !ok {
				owners = map[string]owner{}
				existingOwners[prefix] = owners
			}
			if _, dup := owners[name]; !dup {
				owners[name] = owner{name: name, entry: e}
			}
			if _, ok := firstOf[prefix]; !ok {
				firstOf[prefix] = e
			}
		}
	}

	for _, in := range incoming {
		for _, seg := range walkParams(in.path) {
			prefix, name := seg[0], seg[1]
			owners := existingOwners[prefix]
			if len(owners) == 0 {
				continue
			}
			for otherName, other := range owners {
				if otherName == name {
					continue
				}
				conflicts = append(conflicts, RouteConflict{
					Kind: "wildcard",
					Detail: fmt.Sprintf(
						"路径前缀 %q 下已有通配符 :%s,本次又用了 :%s,"+
							"gin 要求同层通配符同名,否则 panic",
						displayPrefix(prefix), otherName, name),
					First:  other.entry.describe(),
					Second: in.describe(),
				})
			}
		}
	}

	return conflicts
}

// walkParams 逐段扫描路径,返回 [静态前缀, 通配符名] 的列表
func walkParams(p string) [][2]string {
	var out [][2]string
	prefix := ""
	for _, seg := range strings.Split(p, "/") {
		if !strings.HasPrefix(seg, ":") {
			if seg != "" {
				prefix += "/" + seg
			}
			continue
		}
		out = append(out, [2]string{prefix, seg[1:]})
	}
	return out
}

// conflictsAmong 检测一组已存在路由内部的冲突
func conflictsAmong(existing []existingRoute) []RouteConflict {
	var conflicts []RouteConflict

	// METHOD + PATH 重复
	byRoute := map[string][]existingRoute{}
	for _, e := range existing {
		byRoute[e.method+" "+e.path] = append(byRoute[e.method+" "+e.path], e)
	}
	for key, group := range byRoute {
		if len(group) < 2 {
			continue
		}
		for i := 1; i < len(group); i++ {
			conflicts = append(conflicts, RouteConflict{
				Kind:   "duplicate",
				Detail: fmt.Sprintf("路由 %s 被定义了 %d 次", key, len(group)),
				First:  group[0].describe(),
				Second: group[i].describe(),
			})
		}
	}

	// 同层通配符命名不一致
	type owner struct {
		name  string
		entry existingRoute
	}
	segOwners := map[string]map[string]owner{}
	for _, e := range existing {
		for _, seg := range walkParams(e.path) {
			prefix, name := seg[0], seg[1]
			owners, ok := segOwners[prefix]
			if !ok {
				owners = map[string]owner{}
				segOwners[prefix] = owners
			}
			if _, dup := owners[name]; !dup {
				owners[name] = owner{name: name, entry: e}
			}
		}
	}
	for prefix, owners := range segOwners {
		if len(owners) < 2 {
			continue
		}
		names := make([]string, 0, len(owners))
		for n := range owners {
			names = append(names, n)
		}
		sort.Strings(names)
		for i := 1; i < len(names); i++ {
			conflicts = append(conflicts, RouteConflict{
				Kind: "wildcard",
				Detail: fmt.Sprintf("路径前缀 %q 下同时出现通配符 :%s 与 :%s",
					displayPrefix(prefix), names[0], names[i]),
				First:  owners[names[0]].entry.describe(),
				Second: owners[names[i]].entry.describe(),
			})
		}
	}

	sort.SliceStable(conflicts, func(i, j int) bool {
		if conflicts[i].Kind != conflicts[j].Kind {
			return conflicts[i].Kind < conflicts[j].Kind
		}
		if conflicts[i].Detail != conflicts[j].Detail {
			return conflicts[i].Detail < conflicts[j].Detail
		}
		return conflicts[i].Second < conflicts[j].Second
	})
	return conflicts
}

// ScanRouterOutForConflicts 扫描已生成的路由文件,返回其中存在的冲突。
// 供装配层(示例工程的 main.go)在启动前自检使用。
func ScanRouterOutForConflicts(routerOut string) ([]RouteConflict, error) {
	existing, err := collectExistingRoutes(routerOut, nil)
	if err != nil {
		return nil, err
	}
	return conflictsAmong(existing), nil
}
