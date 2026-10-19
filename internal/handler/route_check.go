package handler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hishack/go-codegen/internal/field"
)

// 路由冲突检测。
//
// gin 的路由树有两条硬性约束,而 yaml 是按领域(module)组织的,
// 各领域独立定义 path,生成器若不做全局校验就会产出互相打架的路由:
//
//  1. 同一 METHOD + PATH 只能注册一次,否则 panic:
//     "handlers are already registered for path '/orders'"
//  2. 同一层路径段上的通配符名字必须一致,否则 panic:
//     "':id' in new path '/users/:id/detail' conflicts with existing
//     wildcard ':userId' in existing prefix '/users/:userId'"
//
// 与其在运行时崩溃,不如在生成阶段就报出来,并指明是哪两个接口冲突。

// routeEntry 一条待注册路由及其来源,用于定位问题
type routeEntry struct {
	method string // GET/POST/...
	path   string // 归一化后的路径(去掉类型标注)
	raw    string // 原始 path(保留类型标注)
	api    *field.Api
	domain string // 所属领域(决定生成到哪个 _router.go)
	yaml   string // 来源 yaml 文件名
}

// RouteConflict 一处路由冲突
type RouteConflict struct {
	Kind   string // 冲突类型:duplicate / wildcard
	Detail string // 人类可读的说明
	First  string // 先注册方,格式 "GET /orders (array.yaml 的 GetOrderList)"
	Second string // 后注册方,格式 "GET /orders (order.yaml 的 ListOrders)"
}

// String 输出单条冲突
func (c RouteConflict) String() string {
	return fmt.Sprintf("[%s] %s\n    先: %s\n    后: %s", c.Kind, c.Detail, c.First, c.Second)
}

// collectRoutes 汇总所有领域将要注册的路由
func collectRoutes(yamls *field.YamlList) []routeEntry {
	var out []routeEntry
	for _, d := range groupDomains(yamls) {
		for _, api := range d.Apis {
			out = append(out, routeEntry{
				method: api.Method,
				path:   normalizeRoutePath(api.RoutePath()),
				raw:    api.Path,
				api:    api,
				domain: d.Name,
				yaml:   ownerYaml(yamls, d, api),
			})
		}
	}
	return out
}

// ownerYaml 找到该接口所属的 yaml 文件名,用于报错定位
func ownerYaml(yamls *field.YamlList, d *domain, api *field.Api) string {
	for _, y := range yamls.YamlList {
		for _, a := range y.ApiList {
			if a == api {
				return y.FileName
			}
		}
	}
	return d.Name
}

// normalizeRoutePath 去掉路径参数的类型标注,与 RoutePath 的结果对齐
func normalizeRoutePath(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	return p
}

// describe 生成 "GET /orders (order.yaml 的 ListOrders)" 形式的来源描述
func (e routeEntry) describe() string {
	return fmt.Sprintf("%s %s (%s 的 %s)", e.method, e.raw, e.yaml, e.api.ApiName)
}

// validateRoutes 检测全部路由冲突。
//
// gin 的路由树按"静态前缀"分组,同一前缀下的 :param 必须同名,
// 因此这里既要查 METHOD+PATH 重复,也要查同层通配符命名是否一致。
func validateRoutes(yamls *field.YamlList) []RouteConflict {
	entries := collectRoutes(yamls)
	var conflicts []RouteConflict

	// ---- 规则 1:METHOD + PATH 重复 ----
	byRoute := map[string][]routeEntry{}
	for _, e := range entries {
		key := e.method + " " + e.path
		byRoute[key] = append(byRoute[key], e)
	}
	for key, group := range byRoute {
		if len(group) < 2 {
			continue
		}
		// 同一文件内重复也要报,但优先指出跨文件的情况
		sort.SliceStable(group, func(i, j int) bool { return group[i].yaml < group[j].yaml })
		for i := 1; i < len(group); i++ {
			conflicts = append(conflicts, RouteConflict{
				Kind:   "duplicate",
				Detail: fmt.Sprintf("路由 %s 被定义了 %d 次,gin 只认第一条,其余方法永远不会被调用", key, len(group)),
				First:  group[0].describe(),
				Second: group[i].describe(),
			})
		}
	}

	// ---- 规则 2:同层通配符命名不一致 ----
	// segOwners[静态前缀] = 该层各通配符名 -> 首次出现的路由
	type owner struct {
		name  string
		entry routeEntry
	}
	segOwners := map[string]map[string]owner{}
	for _, e := range entries {
		segs := strings.Split(e.path, "/")
		prefix := ""
		for _, seg := range segs {
			if !strings.HasPrefix(seg, ":") {
				if seg != "" {
					prefix += "/" + seg
				}
				continue
			}
			name := seg[1:]
			owners, ok := segOwners[prefix]
			if !ok {
				owners = map[string]owner{}
				segOwners[prefix] = owners
			}
			if first, exists := owners[name]; exists {
				_ = first // 同名,无冲突
			} else {
				// 该前缀下第一次出现这个通配符名,登记;
				// 若前缀下已有别的名字,下面统一比对
				owners[name] = owner{name: name, entry: e}
			}
		}
	}

	for prefix, owners := range segOwners {
		if len(owners) < 2 {
			continue
		}
		// 同一前缀下出现多个不同的通配符名
		names := make([]string, 0, len(owners))
		for n := range owners {
			names = append(names, n)
		}
		sort.Strings(names)
		for i := 1; i < len(names); i++ {
			conflicts = append(conflicts, RouteConflict{
				Kind: "wildcard",
				Detail: fmt.Sprintf(
					"路径前缀 %q 下同时出现通配符 %s 与 %s,gin 要求同层通配符同名,否则 panic",
					displayPrefix(prefix), ":"+names[0], ":"+names[i]),
				First:  owners[names[0]].entry.describe(),
				Second: owners[names[i]].entry.describe(),
			})
		}
	}

	// 输出稳定,便于测试断言
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

// displayPrefix 把内部累积的前缀("/users/:id" 之类)转成可读形式
func displayPrefix(prefix string) string {
	if prefix == "" {
		return "/"
	}
	return prefix
}

// formatConflicts 把冲突列表拼成一段可直接展示的错误文本
func formatConflicts(conflicts []RouteConflict) string {
	var b strings.Builder
	fmt.Fprintf(&b, "检测到 %d 处路由冲突,已中止生成(修复后重试):", len(conflicts))
	for i, c := range conflicts {
		fmt.Fprintf(&b, "\n\n%d) %s", i+1, c.String())
	}
	b.WriteString("\n\n提示:")
	b.WriteString("\n  - 重复路由:改接口 path,或用接口级 filename 拆到不同领域仍会冲突,请直接改 path")
	b.WriteString("\n  - 通配符同名:把同一前缀下的路径参数改成同名(如 /users/:userId 与 /users/:id/detail -> 都用 :userId)")
	return b.String()
}
