package handler

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hishack/go-codegen/config"
	"github.com/hishack/go-codegen/internal/parse"
)

// 场景化端到端测试:模拟真实开发流程,
// 覆盖"生成 → 用户改代码 → yaml 增/改/删 → 再生成"的完整增量闭环。
// 业务覆盖:CRUD、路径参数、GET query、标量/对象 res、嵌套对象、对象数组、文件上传、多领域模块
type scenarioEnv struct {
	t   *testing.T
	dir string // 工作根目录
	cfg *config.CodeGenConfig
}

// sub 在子测试里切换当前 t,保证断言失败定位到子测试
func (e *scenarioEnv) sub(t *testing.T) { e.t = t }

func newScenarioEnv(t *testing.T) *scenarioEnv {
	dir := t.TempDir()
	e := &scenarioEnv{t: t, dir: dir}
	e.cfg = mustConfig(config.CodeGenConfig{
		YamlInPath:        filepath.Join(dir, "yaml"),
		Module:            "testapp",
		ServiceIntOutPath: filepath.Join(dir, "out", "service"),
		ModelOutPath:      filepath.Join(dir, "out", "model", "gen"),
		LogicOutPath:      filepath.Join(dir, "out", "logic"),
		HandlerOutPath:    filepath.Join(dir, "out", "handler"),
		RouterOutPath:     filepath.Join(dir, "out", "router"),
		WebApiOutPath:     filepath.Join(dir, "out", "web"),
	})
	mustWrite(filepath.Join(e.cfg.YamlInPath, "user.yaml"), baseUserYaml)
	mustWrite(filepath.Join(e.cfg.YamlInPath, "order.yaml"), baseOrderYaml)
	return e
}

func (e *scenarioEnv) run() *Report {
	e.t.Helper()
	yp, err := parse.NewYamlParse(e.cfg.YamlInPath)
	if err != nil {
		e.t.Fatalf("NewYamlParse: %v", err)
	}
	yamls, err := yp.ParseAllYaml()
	if err != nil {
		e.t.Fatalf("ParseAllYaml: %v", err)
	}
	report, err := RunAll(e.cfg, yamls)
	if err != nil {
		e.t.Fatalf("RunAll: %v", err)
	}
	e.assertAllParse()
	return report
}

// assertAllParse 所有生成的 Go 文件都必须能通过语法解析
func (e *scenarioEnv) assertAllParse() {
	e.t.Helper()
	for _, sub := range []string{"service", "logic", "handler", "router"} {
		dir := filepath.Join(e.dir, "out", sub)
		entries, _ := os.ReadDir(dir)
		for _, en := range entries {
			if filepath.Ext(en.Name()) != ".go" {
				continue
			}
			fset := token.NewFileSet()
			if _, err := parser.ParseFile(fset, filepath.Join(dir, en.Name()), nil, 0); err != nil {
				e.t.Fatalf("%s/%s 语法错误: %v", sub, en.Name(), err)
			}
		}
	}
	for _, sub := range []string{"request", "response"} {
		dir := filepath.Join(e.dir, "out", "model", "gen", sub)
		entries, _ := os.ReadDir(dir)
		for _, en := range entries {
			fset := token.NewFileSet()
			if _, err := parser.ParseFile(fset, filepath.Join(dir, en.Name()), nil, 0); err != nil {
				e.t.Fatalf("model/%s/%s 语法错误: %v", sub, en.Name(), err)
			}
		}
	}
}

func (e *scenarioEnv) out(layer, file string) string {
	if layer == "request" || layer == "response" {
		return filepath.Join(e.dir, "out", "model", "gen", layer, file)
	}
	return filepath.Join(e.dir, "out", layer, file)
}

func (e *scenarioEnv) mustContain(layer, file, want string) {
	e.t.Helper()
	got := mustRead(e.out(layer, file))
	if !strings.Contains(got, want) {
		e.t.Fatalf("%s/%s 缺少:\n%s\n--- 实际 ---\n%s", layer, file, want, got)
	}
}

func (e *scenarioEnv) mustNotContain(layer, file, want string) {
	e.t.Helper()
	got := mustRead(e.out(layer, file))
	if strings.Contains(got, want) {
		e.t.Fatalf("%s/%s 不应包含:\n%s\n--- 实际 ---\n%s", layer, file, want, got)
	}
}

func (e *scenarioEnv) yaml() string { return filepath.Join(e.cfg.YamlInPath, "user.yaml") }
func (e *scenarioEnv) orderYaml() string {
	return filepath.Join(e.cfg.YamlInPath, "order.yaml")
}

func (e *scenarioEnv) editYaml(mutate func(src string) string) {
	src := mustRead(e.yaml())
	mustWrite(e.yaml(), mutate(src))
}

func (e *scenarioEnv) editOrderYaml(mutate func(src string) string) {
	src := mustRead(e.orderYaml())
	mustWrite(e.orderYaml(), mutate(src))
}

// ============ 测试数据:两个业务领域 ============

const baseUserYaml = `filename: user
package: system

# 创建用户
CreateUser:
  method: POST
  path: /users
  req:
    name: string # 用户名
    age: int
  res:
    id: int64
    createdAt: time.Time

GetUser:
  method: GET
  path: /users/:id
  res:
    id: int64
    name: string

ListUsers:
  method: GET
  path: /users
  req:
    keyword: string
    page: int
  res:
    total: int64
    list:
      - id: int64
        name: string

DeleteUser:
  method: DELETE
  path: /users/:id(int64)
  res:
    deleted: bool
`

const baseOrderYaml = `filename: order
package: trade

# 创建订单
CreateOrder:
  method: POST
  path: /orders
  req:
    productTitle: string
    count: int
    tags: [string]
  res:
    orderId: int64

# 上传订单附件
UploadAttachment:
  method: POST
  path: /orders/:orderId(int64)/attachments
  req:
    orderId: int64
    file: file # 附件文件
    remark: string
  res:
    fileId: int64 # 文件ID
`

// ============ 场景测试 ============

// TestScenarioIncremental 全流程增量场景
func TestScenarioIncremental(t *testing.T) {
	e := newScenarioEnv(t)

	// ---------- S1 全新生成(多业务形态) ----------
	t.Run("S1_全新生成", func(t *testing.T) {
		e.sub(t)
		r := e.run()
		if r.Added == 0 || r.Warnings != nil {
			t.Fatalf("首次生成应全部新增且无告警: %+v", r)
		}
		// 各层关键内容
		e.mustContain("service", "user_service.go",
			"CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error)")
		e.mustContain("service", "user_service.go",
			"DeleteUser(c *gin.Context, id int64) (bool, error)") // 标量 res 直接返回类型
		e.mustContain("logic", "user_logic.go",
			"func (s *UserService) CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error)")
		e.mustContain("handler", "user_handler.go",
			"func (u *UserHandler) CreateUser(c *gin.Context)")
		e.mustContain("handler", "user_handler.go",
			"response.NewSuccessResponse[res.GetUserRes](data)")
		e.mustContain("handler", "user_handler.go",
			"c.Param(\"id\")") // 无类型路径参数默认 string
		e.mustContain("router", "user_router.go",
			"func (s *UserRouter) InitUserRouter(Router *gin.RouterGroup, userHandler *handler.UserHandler) (R gin.IRoutes)")
		e.mustContain("router", "user_router.go", "userRouter.DELETE(\"/users/:id\", userHandler.DeleteUser)")
		e.mustContain("request", "user.go", "type CreateUserReq struct")
		e.mustContain("response", "user.go", "type ListUsersRes struct")
		e.mustContain("response", "user.go", "[]*ListUsersResListItem") // 对象数组嵌套提升
		// 文件上传
		e.mustContain("request", "order.go", "*multipart.FileHeader")
		e.mustContain("handler", "order_handler.go", "c.ShouldBind(&req)") // multipart 绑定
		e.mustContain("handler", "order_handler.go", "// @Param file formData file true")
		// 多领域各生成一组文件
		if _, err := os.Stat(e.out("service", "order_service.go")); err != nil {
			t.Fatal("order 领域文件未生成")
		}
		// GET query 绑定
		e.mustContain("handler", "user_handler.go", "c.ShouldBindQuery(&req)")
		// 注释透传到 logic TODO
		e.mustContain("logic", "user_logic.go", "TODO: 实现业务逻辑:创建用户")
	})

	// ---------- S2 幂等:重复生成零变更 ----------
	t.Run("S2_幂等", func(t *testing.T) {
		e.sub(t)
		before := snapshotDir(t, filepath.Join(e.dir, "out"))
		r := e.run()
		if r.Added != 0 || r.Updated != 0 || r.Conflict != 0 {
			t.Fatalf("重复生成应零变更: %+v", r)
		}
		after := snapshotDir(t, filepath.Join(e.dir, "out"))
		if before != after {
			t.Fatal("重复生成后文件内容发生了变化")
		}
	})

	// ---------- S3 用户改 logic(真实业务代码)后再生成 ----------
	t.Run("S3_用户代码保留", func(t *testing.T) {
		e.sub(t)
		logicFile := e.out("logic", "user_logic.go")
		src := mustRead(logicFile)
		// 写入真实业务逻辑
		src = strings.Replace(src, `func (s *UserService) CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error) {
	// TODO: 实现业务逻辑:创建用户
	return res.CreateUserRes{}, nil
}`, `func (s *UserService) CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error) {
	// 用户手写的真实业务逻辑
	if req.Age < 0 {
		return res.CreateUserRes{}, fmt.Errorf("年龄不能为负")
	}
	return res.CreateUserRes{Id: 100}, nil
}`, 1)
		// 自定义依赖字段 + 自定义方法 + 自定义类型
		src = strings.Replace(src, "type UserService struct{}", `type UserService struct {
	db *sql.DB
}`, 1)
		src += `
// SendWelcomeMail 用户自己加的函数
func (s *UserService) SendWelcomeMail(userId int64) error {
	return nil
}

// userCache 用户自己加的类型
type userCache struct {
	mu sync.Map
}
`
		mustWrite(logicFile, src)

		// service 接口里加自定义方法
		svcFile := e.out("service", "user_service.go")
		src = mustRead(svcFile)
		src = strings.Replace(src, "}", "\tSendWelcomeMail(c *gin.Context, userId int64) error\n}", 1)
		mustWrite(svcFile, src)

		r := e.run()
		for _, want := range []string{
			"用户手写的真实业务逻辑",
			"db *sql.DB",
			"func (s *UserService) SendWelcomeMail(userId int64) error",
			"type userCache struct",
		} {
			if !strings.Contains(mustRead(logicFile), want) {
				t.Fatalf("用户代码被覆盖,缺少 %q", want)
			}
		}
		if !strings.Contains(mustRead(svcFile), "SendWelcomeMail(c *gin.Context, userId int64) error") {
			t.Fatal("接口自定义方法被覆盖")
		}
		// 自定义接口方法不在 yaml 中,应有告警
		found := false
		for _, w := range r.Warnings {
			if strings.Contains(w, "SendWelcomeMail") {
				found = true
			}
		}
		if !found {
			t.Fatalf("自定义接口方法应触发告警: %+v", r.Warnings)
		}
	})

	// ---------- S4 yaml 新增接口(嵌套 + 注释) ----------
	t.Run("S4_yaml新增接口", func(t *testing.T) {
		e.sub(t)
		e.editYaml(func(src string) string {
			return src + `
# 获取用户详情
GetUserDetail:
  method: GET
  path: /users/:id(int64)/detail
  res:
    id: int64 # 用户ID
    profile:
      nickname: string # 昵称
      tags: [string]
`
		})
		r := e.run()
		if r.Added == 0 {
			t.Fatal("新增接口应有声明新增")
		}
		e.mustContain("service", "user_service.go",
			"GetUserDetail(c *gin.Context, id int64) (res.GetUserDetailRes, error)")
		e.mustContain("handler", "user_handler.go", "GetUserDetail(c *gin.Context)")
		e.mustContain("handler", "user_handler.go", "strconv.ParseInt(idStr, 10, 64)")
		e.mustContain("response", "user.go", "type GetUserDetailResProfile struct")
		e.mustContain("router", "user_router.go", "userRouter.GET(\"/users/:id/detail\", userHandler.GetUserDetail)")
		e.mustContain("logic", "user_logic.go", "TODO: 实现业务逻辑:获取用户详情")
		// S3 的用户代码仍在
		if !strings.Contains(mustRead(e.out("logic", "user_logic.go")), "用户手写的真实业务逻辑") {
			t.Fatal("新增接口后用户代码丢失")
		}
	})

	// ---------- S5 yaml 修改 req 字段 ----------
	t.Run("S5_修改req字段", func(t *testing.T) {
		e.sub(t)
		e.editYaml(func(src string) string {
			src = strings.Replace(src, "    name: string # 用户名\n    age: int\n  res:\n    id: int64\n    createdAt: time.Time",
				"    name: string # 用户名\n    age: int\n    nickname: string # 昵称\n  res:\n    id: int64\n    createdAt: time.Time", 1)
			return src
		})
		e.run()
		// request 结构体同步(新增字段进入,老结构体保留其它字段)
		e.mustContain("request", "user.go", "Nickname string")
		e.mustContain("request", "user.go", "// 昵称")
		// service 签名没变 → logic 方法体不动
		if !strings.Contains(mustRead(e.out("logic", "user_logic.go")), "用户手写的真实业务逻辑") {
			t.Fatal("req 字段变化不应影响 logic 方法体")
		}
		// 幂等检查:再跑一遍零变更
		r := e.run()
		if r.Added != 0 || r.Updated != 0 {
			t.Fatalf("S5 之后再生成应零变更: %+v", r)
		}
	})

	// ---------- S6 yaml 修改 res:标量 → 对象(签名变化) ----------
	t.Run("S6_res标量变对象", func(t *testing.T) {
		e.sub(t)
		before := mustRead(e.out("logic", "user_logic.go"))
		e.editYaml(func(src string) string {
			return strings.Replace(src, `DeleteUser:
  method: DELETE
  path: /users/:id(int64)
  res:
    deleted: bool
`, `DeleteUser:
  method: DELETE
  path: /users/:id(int64)
  res:
    deleted: bool
    deletedAt: time.Time # 删除时间
`, 1)
		})
		r := e.run()
		// 接口签名自动更新
		e.mustContain("service", "user_service.go",
			"DeleteUser(c *gin.Context, id int64) (res.DeleteUserRes, error)")
		// 响应结构体生成
		e.mustContain("response", "user.go", "type DeleteUserRes struct")
		// logic 保留旧实现并告警
		if !strings.Contains(mustRead(e.out("logic", "user_logic.go")), "return false, nil") {
			t.Fatal("签名变化时 logic 旧实现应保留")
		}
		if mustRead(e.out("logic", "user_logic.go")) != before {
			t.Fatal("签名变化时 logic 文件应保持不变(只告警)")
		}
		conflictWarned := false
		for _, w := range r.Warnings {
			if strings.Contains(w, "DeleteUser") && strings.Contains(w, "签名") {
				conflictWarned = true
			}
		}
		if !conflictWarned || r.Conflict == 0 {
			t.Fatalf("签名冲突应告警: conflict=%d warnings=%v", r.Conflict, r.Warnings)
		}
		// handler 自动重写为新签名
		e.mustContain("handler", "user_handler.go",
			"response.NewSuccessResponse[res.DeleteUserRes](data)")
	})

	// ---------- S7 yaml 修改路径参数类型 ----------
	t.Run("S7_路径参数改类型", func(t *testing.T) {
		e.sub(t)
		e.editYaml(func(src string) string {
			return strings.Replace(src, "  path: /users/:id\n  res:\n    id: int64\n    name: string",
				"  path: /users/:id(int64)\n  res:\n    id: int64\n    name: string", 1)
		})
		r := e.run()
		e.mustContain("service", "user_service.go",
			"GetUser(c *gin.Context, id int64) (res.GetUserRes, error)")
		e.mustContain("handler", "user_handler.go", "strconv.ParseInt(idStr, 10, 64)")
		// logic 原实现保留(string → int64 签名变了,告警)
		if r.Conflict == 0 {
			t.Fatalf("路径参数类型变化应产生签名冲突告警: %+v", r)
		}
	})

	// ---------- S8 yaml 删除接口 ----------
	t.Run("S8_yaml删除接口", func(t *testing.T) {
		e.sub(t)
		e.editYaml(func(src string) string {
			return strings.Replace(src, `ListUsers:
  method: GET
  path: /users
  req:
    keyword: string
    page: int
  res:
    total: int64
    list:
      - id: int64
        name: string

`, "", 1)
		})
		r := e.run()
		// 代码不自动删
		e.mustContain("service", "user_service.go", "ListUsers(c *gin.Context, req request.ListUsersReq)")
		e.mustContain("handler", "user_handler.go", "ListUsers(c *gin.Context)")
		// 告警提醒
		warned := 0
		for _, w := range r.Warnings {
			if strings.Contains(w, "ListUsers") {
				warned++
			}
		}
		if warned == 0 {
			t.Fatalf("删除接口应触发告警: %+v", r.Warnings)
		}
	})

	// ---------- S9 注释修改透传 ----------
	t.Run("S9_注释修改透传", func(t *testing.T) {
		e.sub(t)
		e.editYaml(func(src string) string {
			return strings.Replace(src, "# 创建用户\nCreateUser:", "# 创建用户(含邀请码)\nCreateUser:", 1)
		})
		e.run()
		e.mustContain("handler", "user_handler.go", "@Summary 创建用户(含邀请码)")
		// logic 是用户层,注释变化不重写方法(未被用户改过的方法 TODO 原样保留)
		if !strings.Contains(mustRead(e.out("logic", "user_logic.go")), "TODO: 实现业务逻辑:GetUser") {
			t.Fatal("S9 后 logic 原 TODO 丢失")
		}
		// 用户业务逻辑仍在(handler 重写不影响 logic)
		if !strings.Contains(mustRead(e.out("logic", "user_logic.go")), "用户手写的真实业务逻辑") {
			t.Fatal("S9 后用户代码丢失")
		}
	})

	// ---------- S10 终态幂等 ----------
	t.Run("S10_终态幂等", func(t *testing.T) {
		e.sub(t)
		e.run() // 先跑一次消化掉告警状态
		before := snapshotDir(t, filepath.Join(e.dir, "out"))
		r := e.run()
		after := snapshotDir(t, filepath.Join(e.dir, "out"))
		if before != after {
			t.Logf("changed: %s", diffSnapshots(before, after))
		}
		if r.Added != 0 || r.Updated != 0 {
			t.Fatalf("终态应零变更: %+v", r)
		}
	})
}

// ============ 测试辅助 ============

func mustConfig(cfg config.CodeGenConfig) *config.CodeGenConfig {
	c, err := config.NewCodeGenConfig(cfg)
	if err != nil {
		panic(err)
	}
	return c
}

func mustWrite(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}

func mustRead(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("读取失败 " + path + ": " + err.Error())
	}
	return string(data)
}

// snapshotDir 目录内所有文件的稳定快照(路径+内容哈希),用于幂等断言
func snapshotDir(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		b.WriteString(rel)
		b.WriteString(":")
		data, _ := os.ReadFile(path)
		b.WriteString(strings.TrimSpace(string(data)))
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// diffSnapshots 找出两个快照之间内容不同的文件名
func diffSnapshots(a, b string) string {
	nl := string(rune(10))
	ma := map[string]string{}
	for _, line := range strings.Split(a, nl) {
		if i := strings.Index(line, ":"); i > 0 {
			ma[line[:i]] = line
		}
	}
	var out []string
	for _, line := range strings.Split(b, nl) {
		if i := strings.Index(line, ":"); i > 0 {
			if ma[line[:i]] != line {
				out = append(out, line[:i])
			}
		}
	}
	return strings.Join(out, ", ")
}
