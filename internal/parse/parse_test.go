package parse

import (
	"os"
	"strings"
	"testing"

	"github.com/hishack/go-codegen/internal/field"
)

const yamlDir = "../../yamlPath"

// TestParseAllYaml 解析全部测试用例,验证核心结构
func TestParseAllYaml(t *testing.T) {
	p, err := NewYamlParse(yamlDir)
	if err != nil {
		t.Fatalf("NewYamlParse: %v", err)
	}
	list, err := p.ParseAllYaml("")
	if err != nil {
		t.Fatalf("ParseAllYaml: %v", err)
	}
	if len(list.YamlList) != 4 {
		t.Fatalf("期望 4 个 yaml 文件,实际 %d", len(list.YamlList))
	}

	// 汇总所有接口,应涵盖 4 个文件的全部接口
	allApis := list.AllApis()
	if len(allApis) < 20 {
		t.Fatalf("接口数量异常: %d", len(allApis))
	}

	// 幂等:两次解析结构体登记数量一致
	list2, err := p.ParseAllYaml("")
	if err != nil {
		t.Fatalf("第二次 ParseAllYaml: %v", err)
	}
	if len(list.AllStructs()) != len(list2.AllStructs()) {
		t.Fatalf("两次解析结构体数量不一致: %d vs %d", len(list.AllStructs()), len(list2.AllStructs()))
	}
}

// TestParseComments 注释是本项目的核心需求,单独验证
func TestParseComments(t *testing.T) {
	src := `filename: demo
package: system

# 创建用户
CreateUser:
  method: POST # 创建动作
  path: /users
  req:
    # 用户账号
    userAccount: string # 账号,4-12位
    age: int
  res:
    id: int64 # 用户ID
`
	tmp := t.TempDir() + "/comment.yaml"
	writeFile(t, tmp, src)
	yy, err := ParseYamlFile(tmp, "")
	if err != nil {
		t.Fatalf("ParseYamlFile: %v", err)
	}
	if len(yy.ApiList) != 1 {
		t.Fatalf("期望 1 个接口,实际 %d", len(yy.ApiList))
	}
	api := yy.ApiList[0]
	if api.ApiName != "CreateUser" {
		t.Fatalf("接口名错误: %s", api.ApiName)
	}
	if api.Comment != "创建用户" {
		t.Errorf("接口注释期望 [创建用户],实际 [%s]", api.Comment)
	}
	if api.Req == nil || len(api.Req.Fields) != 2 {
		t.Fatalf("req 字段数错误: %+v", api.Req)
	}
	f0 := api.Req.Fields[0]
	if f0.Name != "UserAccount" || f0.Tag != "userAccount" {
		t.Errorf("字段名/Tag 错误: %+v", f0)
	}
	// 上方注释与行注释同时存在时合并
	if f0.Comment != "用户账号;账号,4-12位" {
		t.Errorf("注释期望 [用户账号;账号,4-12位],实际 [%s]", f0.Comment)
	}
	f1 := api.Req.Fields[1]
	if f1.Comment != "" {
		t.Errorf("无注释字段期望空,实际 [%s]", f1.Comment)
	}
	if api.Res == nil || api.Res.Fields[0].Comment != "用户ID" {
		t.Errorf("res 字段注释错误: %+v", api.Res)
	}
}

// TestNestedLifting 嵌套对象提升:多层嵌套 + 对象数组
func TestNestedLifting(t *testing.T) {
	src := `filename: nested
package: system

GetOrderDetail:
  method: GET
  path: /orders/:orderId(int64)/detail
  res:
    orderId: int64
    items:
      - productId: int64
        attrs:
          - name: string
            value: string
`
	tmp := t.TempDir() + "/nested.yaml"
	writeFile(t, tmp, src)
	yy, err := ParseYamlFile(tmp, "")
	if err != nil {
		t.Fatalf("ParseYamlFile: %v", err)
	}
	api := yy.ApiList[0]

	// 路径参数
	if len(api.Params) != 1 || api.Params[0].Name != "OrderId" || api.Params[0].Type.GoString() != "int64" {
		t.Fatalf("路径参数错误: %+v", api.Params)
	}

	// 结构体提升:根 Vo -> items 数组元素 Item -> attrs 数组元素 Item
	res := api.Res
	if res.Name != "GetOrderDetailVo" {
		t.Errorf("根结构体名错误: %s", res.Name)
	}
	items := res.Field("Items")
	if items == nil || items.Type.Kind != field.KindSlice {
		t.Fatalf("items 字段错误: %+v", items)
	}
	itemStruct := items.Type.Elem.Struct
	if itemStruct.Name != "GetOrderDetailVoItemsItem" {
		t.Errorf("数组元素结构体名错误: %s", itemStruct.Name)
	}
	attrs := itemStruct.Field("Attrs")
	if attrs == nil || attrs.Type.Elem.Struct == nil {
		t.Fatalf("attrs 字段错误: %+v", attrs)
	}
	if attrs.Type.Elem.Struct.Name != "GetOrderDetailVoItemsItemAttrsItem" {
		t.Errorf("深层结构体名错误: %s", attrs.Type.Elem.Struct.Name)
	}

	// GoString 类型检查
	if got := items.Type.GoString(); got != "[]*GetOrderDetailVoItemsItem" {
		t.Errorf("items 类型期望 []*GetOrderDetailVoItemsItem,实际 %s", got)
	}
}

// TestPrefix 前缀只作用于请求结构体
func TestPrefix(t *testing.T) {
	src := `filename: user
package: system

CreateUser:
  method: POST
  path: /users
  req:
    name: string
  res:
    id: int64
`
	tmp := t.TempDir() + "/user.yaml"
	writeFile(t, tmp, src)
	yy, err := ParseYamlFile(tmp, "YiKou")
	if err != nil {
		t.Fatalf("ParseYamlFile: %v", err)
	}
	api := yy.ApiList[0]
	if api.Req.Name != "YiKouCreateUserRequest" {
		t.Errorf("req 结构体名期望 YiKouCreateUserRequest,实际 %s", api.Req.Name)
	}
	if api.Res.Name != "CreateUserVo" {
		t.Errorf("res 结构体名期望 CreateUserVo,实际 %s", api.Res.Name)
	}
}

// TestEdgeCases edge.yaml 的全部边界场景
func TestEdgeCases(t *testing.T) {
	p, _ := NewYamlParse(yamlDir)
	list, err := p.ParseAllYaml("")
	if err != nil {
		t.Fatalf("ParseAllYaml: %v", err)
	}
	var edge *field.Yaml
	for _, y := range list.YamlList {
		if y.FileName == "edge" {
			edge = y
		}
	}
	if edge == nil {
		t.Fatal("找不到 edge.yaml")
	}
	apis := map[string]*field.Api{}
	for _, a := range edge.ApiList {
		apis[a.ApiName] = a
	}

	// 无 req 无 res
	if a := apis["HealthCheck"]; a.HasReq() || a.HasRes() {
		t.Errorf("HealthCheck 不应有 req/res")
	}
	// 只有 req
	if a := apis["Logout"]; !a.HasReq() || a.HasRes() {
		t.Errorf("Logout 应只有 req")
	}
	// req: {} 空对象等价于没有
	if a := apis["Ping"]; a.HasReq() {
		t.Errorf("Ping 的空 req 应被视为不存在")
	}
	// GET 带 req
	if a := apis["SearchUser"]; !a.HasReq() || a.Req.Fields[0].Name != "Keyword" {
		t.Errorf("SearchUser req 错误")
	}
	// 多路径参数
	if a := apis["GetUserOrder"]; len(a.Params) != 2 || a.Params[0].Name != "UserId" {
		t.Errorf("GetUserOrder 路径参数错误: %+v", a.Params)
	}
	// 无类型路径参数默认 string
	if a := apis["GetArticle"]; len(a.Params) != 1 || a.Params[0].Type.GoString() != "string" {
		t.Errorf("GetArticle 路径参数错误: %+v", a.Params)
	}
	// 路由路径会去掉类型标注
	if got := apis["GetUserOrder"].RoutePath(); got != "/users/:userId/orders/:orderId" {
		t.Errorf("RoutePath 期望 /users/:userId/orders/:orderId,实际 %s", got)
	}
}

// TestScalarRes 单基础类型字段的 res 直接返回标量
func TestScalarRes(t *testing.T) {
	src := `filename: user
package: system

DeleteUser:
  method: DELETE
  path: /users/:id(int64)
  res:
    deleted: bool
CreateUser:
  method: POST
  path: /users
  res:
    id: int64
    createdAt: time.Time
`
	tmp := t.TempDir() + "/user.yaml"
	writeFile(t, tmp, src)
	yy, err := ParseYamlFile(tmp, "")
	if err != nil {
		t.Fatalf("ParseYamlFile: %v", err)
	}
	del := yy.ApiList[0]
	if got := del.ScalarRes(); got == nil || got.Name != "bool" {
		t.Errorf("DeleteUser 期望标量 bool,实际 %v", got)
	}
	// 多字段 res 不是标量
	create := yy.ApiList[1]
	if create.ScalarRes() != nil {
		t.Errorf("CreateUser 有两个字段,不应是标量")
	}
}

// TestValidationErrors 各种非法输入要有清晰报错
func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"method非法", "filename: a\npackage: system\nX:\n  method: GETT\n  path: /x\n", "method"},
		{"path不以/开头", "filename: a\npackage: system\nX:\n  method: GET\n  path: x\n", "path"},
		{"缺method", "filename: a\npackage: system\nX:\n  path: /x\n", "缺少 method"},
		{"类型不支持", "filename: a\npackage: system\nX:\n  method: GET\n  path: /x\n  res:\n    a: long\n", "不支持的字段类型"},
		{"接口名小写", "filename: a\npackage: system\ngetUser:\n  method: GET\n  path: /x\n", "接口名"},
		{"字段名大写开头", "filename: a\npackage: system\nX:\n  method: GET\n  path: /x\n  res:\n    Name: string\n", "字段名"},
		{"map格式错", "filename: a\npackage: system\nX:\n  method: GET\n  path: /x\n  res:\n    a: map[string]\n", "map 类型格式非法"},
		{"多元素数组", "filename: a\npackage: system\nX:\n  method: GET\n  path: /x\n  res:\n    a: [int, string]\n", "单元素"},
		{"路径参数类型不支持", "filename: a\npackage: system\nX:\n  method: GET\n  path: /x/:id(time.Time)\n", "类型不支持"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmp := t.TempDir() + "/bad.yaml"
			writeFile(t, tmp, c.src)
			_, err := ParseYamlFile(tmp, "")
			if err == nil {
				t.Fatalf("期望报错包含 [%s],实际没有报错", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("期望报错包含 [%s],实际 [%s]", c.want, err.Error())
			}
		})
	}
}

// TestGlobalDuplicate 跨文件接口名/文件名冲突
func TestGlobalDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/a.yaml", "filename: user\npackage: system\nGetUser:\n  method: GET\n  path: /a\n")
	writeFile(t, dir+"/b.yaml", "filename: user\npackage: system\nGetOrder:\n  method: GET\n  path: /b\n")
	p, _ := NewYamlParse(dir)
	_, err := p.ParseAllYaml("")
	if err == nil || !strings.Contains(err.Error(), "filename") {
		t.Fatalf("期望 filename 冲突报错,实际 %v", err)
	}

	dir2 := t.TempDir()
	writeFile(t, dir2+"/a.yaml", "filename: user\npackage: system\nGetUser:\n  method: GET\n  path: /a\n")
	writeFile(t, dir2+"/b.yaml", "filename: order\npackage: system\nGetUser:\n  method: GET\n  path: /b\n")
	p2, _ := NewYamlParse(dir2)
	_, err = p2.ParseAllYaml("")
	if err == nil || !strings.Contains(err.Error(), "接口名") {
		t.Fatalf("期望接口名冲突报错,实际 %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}
}

// TestStructNameCollision 同名不同结构的结构体必须报错,而不是静默复用
func TestStructNameCollision(t *testing.T) {
	// GetUser 的 res 名是 GetUserVo;再来一个名为 GetUserVo 的接口,res 名也是
	// GetUserVo,结构不同 → 必须报错
	src := `filename: user
package: system

GetUser:
  method: GET
  path: /users/:id(int64)
  res:
    id: int64
    name: string
    age: int
GetUserVo:
  method: GET
  path: /users/vo/:id(int64)
  res:
    id: int64
    profile:
      nickname: string
`
	tmp := t.TempDir() + "/user.yaml"
	writeFile(t, tmp, src)
	_, err := ParseYamlFile(tmp, "")
	if err == nil {
		t.Fatal("期望结构体名冲突报错,实际成功")
	}
	if !strings.Contains(err.Error(), "GetUserVo") || !strings.Contains(err.Error(), "冲突") {
		t.Fatalf("报错信息不明确: %v", err)
	}

	// 同名同结构:合法(不同接口的响应结构完全一样时复用)
	ok := `filename: user
package: system

GetA:
  method: GET
  path: /a
  res:
    same: string
GetB:
  method: GET
  path: /b
  res:
    same: string
`
	tmp2 := t.TempDir() + "/user2.yaml"
	writeFile(t, tmp2, ok)
	yy, err := ParseYamlFile(tmp2, "")
	if err != nil {
		t.Fatalf("不同接口各自的结构体不应报错: %v", err)
	}
	if len(yy.ApiList[0].Res.Fields) != 1 || len(yy.ApiList[1].Res.Fields) != 1 {
		t.Fatal("两个接口的 res 字段数错误")
	}
}
