# go-codegen

写 YAML,一键生成 Go Web 脚手架(Gin)。

一个 yaml 文件描述一个领域模块的全部接口,运行生成器即可**增量**生成 / 更新七层代码。
yaml 里写的注释会一路带到 Go 字段注释、swagger 注解和前端 TS jsdoc —— **接口文档只写一遍**。

- **七层同步**:request / response / service / logic / handler / router / TS api 一次生成
- **增量而非覆盖**:基于 `go/ast` 定位 + 文本拼接,用户手写的业务代码一字节不动,重复运行幂等
- **注释透传**:接口注释进 `@Summary` 与 TS jsdoc,字段注释进结构体字段,不用再抄一遍
- **生成期自检**:路由冲突、结构体撞名、非法类型都在生成时中止并给出文件与接口名,而不是等到运行期 panic
- **两种用法**:命令行 `gen`,或 `import "github.com/hishack/go-codegen/pkg/codegen"` 嵌进自己的构建流程

---

## 环境要求

| 项 | 要求 |
| --- | --- |
| Go | 1.25 及以上(见 `go.mod`) |
| 目标工程 | 使用 [gin](https://github.com/gin-gonic/gin) 的 Web 工程 |
| 目标工程必需包 | `pkg/response`;`pkg/errorutil`(有接口需要参数校验时才会 import) |

## 安装

```bash
go install github.com/hishack/go-codegen/cmd/gen@latest
```

装到 `$GOBIN`(默认 `$GOPATH/bin`),确认可用:

```bash
gen -h
```

不想装也可以直接跑:

```bash
go run github.com/hishack/go-codegen/cmd/gen@latest -config gen.yaml
```

---

## 五分钟上手

### 1. 准备目标工程

生成器只负责写业务代码,不生成 Web 框架本身。你的工程需要先具备两个基础包,
生成出来的 handler 会 import 它们(完整实现见 [`example/pkg/`](example/pkg),可直接复制):

| 包 | 需要提供 | 用途 |
| --- | --- | --- |
| `pkg/response` | `BaseResponse[T]`、`NewSuccessResponse[T](data T)`、`NewErrorResponse[T](err error)` | handler 统一包装响应体 |
| `pkg/errorutil` | `BusinessError` 类型、`ParamsError` 变量 | 参数/绑定失败时构造 400 的错误体 |

> 这两个包的 import 前缀是 `<module>/pkg/response` 和 `<module>/pkg/errorutil`,
> 所以必须放在目标工程根目录下的 `pkg/` 里,`module` 也要配成真实的 go module 名。

### 2. 写 `gen.yaml`

放在工程根目录,描述"yaml 从哪读、代码往哪写":

```yaml
yaml: ./codegen/yaml          # 输入:领域 yaml 目录
module: myapp                 # 目标工程 go module 名(拼 import 路径)

serviceOut: ./internal/service
modelOut:   ./internal/model/gen    # 其下自动分 request/ response/ 子包
logicOut:   ./internal/logic
handlerOut: ./internal/handler
routerOut:  ./internal/router

webOut:  ./web/src/api        # 前端 api 目录,留空则跳过前端生成
group:   /api/v1              # 路由统一前缀,可省略
author:  your-name
github:  https://github.com/your-name
```

> 所有相对路径都相对**本配置文件所在目录**解析,因此 `gen` 在哪个工作目录下执行,结果都一致。

### 3. 写一个领域 yaml

`codegen/yaml/user.yaml`:

```yaml
filename: user      # 领域名,决定生成的文件名
package: system     # swagger @Tags

# 创建用户                      <- 接口注释
CreateUser:
  method: POST
  path: /users
  req:
    name: string # 用户名        <- 字段注释
    age: int
  res:
    id: int64
```

### 4. 生成

```bash
gen -config gen.yaml
```

```
✔ 生成完成:yaml 文件 1 个,接口 1 个;新增 1,更新 0,签名冲突 0
```

产物(以 `module: myapp` 为例):

```
internal/model/gen/request/user.go    CreateUserReq
internal/model/gen/response/user.go   CreateUserRes
internal/service/user_service.go      IUserService
internal/logic/user_logic.go          UserService(空实现,写业务逻辑的地方)
internal/handler/user_handler.go      UserHandler
internal/router/user_router.go        UserRouter
web/src/api/user.ts                   UserApi
```

### 5. 挂到 gin

路由是用代码装配的,生成器只负责生成 `InitXxxRouter`:

```go
api := r.Group("")   // 前缀由 gen.yaml 的 group 决定,这里传空组即可

(&router.UserRouter{}).InitUserRouter(
    api,
    handler.NewUserHandler(logic.NewUserService()),
)
```

然后去 `internal/logic/user_logic.go` 填业务逻辑 —— 那个文件里的方法体**永远不会被生成器覆盖**。

---

## 看完整示例

仓库自带一个可编译的演示工程 [`example/`](example)(独立 go module,以库的方式引用本模块):

```bash
# 主测试集(根 gen.yaml,输出到 example/)
go run ./cmd/gen -config gen.yaml
cd example && go build ./...

# 或走 example 自己的配置(带注释的教学样例 + 多领域)
cd example
go run ./cmd/gen -config cmd/gen/gen.yaml
go run ./cmd/gen -config ../gen.yaml
```

[`example/cmd/gen/yaml/threshold.yaml`](example/cmd/gen/yaml/threshold.yaml) 是一份**逐字段带注释**的教学样例,
可以直接对照生成结果,看注释如何落到 Go 注释、swagger 注解和 TS jsdoc 上。

---

## yaml 语法

### 文件级字段

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `filename` | 是 | 领域名,小驼峰,决定生成的文件名(`user` → `user_service.go`) |
| `package` | 是 | swagger `@Tags` 等元数据 |

文件里其余的顶层 key 都是接口名,每个接口是一个对象。

### 接口级字段

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `method` | 是 | `get` / `post` / `put` / `patch` / `delete`,不区分大小写 |
| `path` | 是 | 路由路径,支持路径参数 `:userId` 与类型标注 `:userId(int64)` |
| `req` | 否 | 请求体;`req: {}` 等价于不写 |
| `res` | 否 | 响应体,可省略 |
| `filename` | 否 | 仅本接口生效,把该接口输出到另一个领域文件(`user` → 拆出 `profile`/`auth`) |

接口对象里**只允许**这五个 key,写错会直接报错并指出接口名。

### 类型系统

| yaml 写法 | 生成的 Go 类型 | 备注 |
| --- | --- | --- |
| `string` / `int` / `int64` / `uint8` … | 原样 | 整型、无符号整型家族全支持 |
| `float32` / `float64` / `bool` | 原样 | |
| `time.Time` | `time.Time` | swagger 侧自动处理 |
| `time.Duration` | `time.Duration` | 自动加 `swaggertype:"integer"`,避免 swag 解析失败 |
| `any` | `any` | |
| `map[string]string` | `map[string]string` | 写 `map[k]v` 形式 |
| `[int]` / `[string]` | `[]int` / `[]string` | 数组只支持**单元素**写法 |
| `- id: int64` | `[]XxxItem` | 对象数组用一段只有一个元素的列表 |
| `file` | `*multipart.FileHeader` | **只能用于 req** |
| 嵌套对象(mapping) | 提升为独立命名结构体 | 命名规则见下方 |

路径参数会自动并入该接口的请求结构体,handler 从 `c.Param` 取值后转成对应类型:

```yaml
GetUser:
  method: GET
  path: /users/:userId(int64)   # 不写 (int64) 则按 string 处理
  res:
    id: int64
```

### 注释怎么透传

```yaml
# 创建用户                                  <- 接口上方注释
CreateUser:
  method: POST
  path: /users          # 新增用户         <- path 行尾注释,与上方注释合并
  req:
    name: string # 用户名                    <- 字段行尾注释
```

注释会原样进入三处:Go 文档注释与字段注释、swagger `@Summary`/`@Description`/`@Param`、前端 TS jsdoc。
上方注释与行尾注释同时存在时用 `;` 连接。

> **建议**:说明性的长文字写在 `path:` 行尾,不要写在接口上方 —— 接口上方的注释会被拼进
> swagger `@Success` 描述,多行文字会污染文档。
> 另外注释里的引号会进入 `@Param`/`@Success`,**不要用中文弯引号 `“”`**,会破坏注解配对。

### 约束与常见报错

| 约束 | 报错信息 |
| --- | --- |
| 接口名全局唯一(跨 yaml 文件也不能重名,api/service 是单包) | `redeclared in this block` |
| 字段名必须是**小驼峰**,小写字母开头,仅字母数字 | `字段名 "Xxx" 非法` |
| 结构体同名但结构不同 | `结构体名 "Xxx" 冲突:...请修改其中一个接口名` |
| 数组只能写单元素(`[int]`,不能 `[int, string]`) | `数组写法只支持单元素` |
| `file` 不能出现在 `res` | `file 类型只能用于 req 字段` |
| `path` 参数只支持基础类型 | `路径参数类型不支持` |

---

## 生成产物

以 `user.yaml` 里的 `CreateUser` 接口为例:

| 层 | 文件 | 生成内容 |
| --- | --- | --- |
| model/request | `internal/model/gen/request/user.go` | `CreateUserReq`(`json` + `form` 双 tag) |
| model/response | `internal/model/gen/response/user.go` | `CreateUserRes`(嵌套对象自动提升命名) |
| service | `internal/service/user_service.go` | 接口 `IUserService` 及方法签名 |
| logic | `internal/logic/user_logic.go` | 空实现 `UserService`(带 TODO) |
| handler | `internal/handler/user_handler.go` | 绑定 → 调 service → 包装响应,含 swagger 注解 |
| router | `internal/router/user_router.go` | `InitUserRouter(Router *gin.RouterGroup, ...)` |
| web | `web/src/api/user.ts` | TS 类型 + `UserApi` 请求类(axios) |

### 命名约定

```
request.CreateUserReq / response.CreateUserRes     模型(model/gen 两分包)
IUserService.CreateUser                            接口(service)
UserService.CreateUser                             实现(logic)
UserHandler.CreateUser                             胶水(handler)
UserRouter.InitUserRouter                          路由(router)
UserApi.createUser                                 TS 方法
```

嵌套结构体命名为 `<父结构体><字段大驼峰>`,对象数组再加 `Item`,
例如 `GetCharacterSkillsResListItem`。

response 包在 handler/service/logic 里统一以 `res` 别名导入,与 `pkg/response` 不冲突。

### 绑定方式

handler 按请求形态自动选择 gin 的绑定方法:

| req 形态 | 绑定 |
| --- | --- |
| 含 `file` 字段 | `c.ShouldBind(&req)`(multipart 表单) |
| GET / DELETE | `c.ShouldBindQuery(&req)`(form 标签) |
| POST / PUT / PATCH | `c.ShouldBindJSON(&req)` |

文件**下载**属于纯响应行为(不生成结构体),在 logic 里返回内容、在 handler 里用 `c.File` / `c.Data` 输出即可。

### 自动生成的校验与状态码

请求结构体会按字段类型自动带 `binding` 标签,无需在 yaml 里声明:

| 字段形态 | 生成的标签 |
| --- | --- |
| 关键标识(`name` / `title` / `userName` / `password` / `phone` / `email`) | `required,max=255` |
| 其他 string | `omitempty,max=255` |
| 整型 / 浮点 | `omitempty,gte=0` |
| bool / 时间 / file | `omitempty` |
| 嵌套对象 / 切片 | `omitempty,dive` |

> 切片与嵌套对象**不设 required** —— yaml 语法无法区分"必填数组"和"可选数组",
> 一律 required 会误拒合法请求,因此只校验"传了之后是否合法"。

handler 按错误来源返回不同状态码:

| 情况 | 状态码 |
| --- | --- |
| 绑定 / 参数解析失败(请求本身不合法) | `400` |
| 业务失败(`logic` 返回 error) | `422`(固定 `http.StatusUnprocessableEntity`) |
| 成功 | `200` |

业务失败**统一**是 422,不再按错误类型二次映射:请求能解析、但业务上处理不了,语义上正是
"格式没问题、内容无法处理"。具体是哪一类失败,由响应体里的业务错误码表达 —— 这样前端只需
判断 HTTP 状态,业务分支交给 code,不必为每种错误维护一张状态码映射表。

---

## 配置项(`gen.yaml`)

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `yaml` | 是 | 输入的 yaml 目录 |
| `module` | 是 | 目标工程 go module 名,用于拼 import 路径 |
| `serviceOut` | 是 | service 接口目录 |
| `modelOut` | 是 | 模型父目录,其下自动分 `request/`、`response/` 子包 |
| `logicOut` | 是 | logic 空实现目录 |
| `handlerOut` | 是 | handler 目录 |
| `routerOut` | 是 | 路由目录 |
| `webOut` | 否 | 前端 TS api 目录,留空跳过前端生成 |
| `group` | 否 | 路由分组前缀,如 `/api/v1`,同时写进 router、swagger `@Router`、TS url |
| `author` | 否 | 写入生成文件头的作者 |
| `github` | 否 | 写入生成文件头的 GitHub 地址 |

`group` 的写法很宽松,`api/v1`、`/api/v1/`、`/api/v1` 都会被归一化成 `/api/v1`。
**同一目标工程的各份配置要用一致的前缀**,否则跨配置的路由冲突检测会失效。

目录不存在时会自动创建。

---

## 增量生成(核心特性)

写入不是整体覆盖,而是 `go/ast` 定位 + 文本拼接(`internal/merge`):
没碰到的用户代码一字节不变,重复运行输出完全一致(幂等)。

| 场景 | 行为 |
| --- | --- |
| yaml 新增接口 | 所有层自动补齐 |
| yaml 改 req/res 字段 | request/response 结构体自动同步;service 签名不变时 logic 不动 |
| yaml 改接口签名 | 接口签名更新;handler/路由重写;**logic 保留旧实现并打印告警** |
| yaml 删除接口 | 不删代码,打印告警提醒手动清理 |
| 用户手写的 logic 业务逻辑、结构体字段、自定义函数/类型 | **永不触碰** |
| handler 方法体 | 由生成器管理(纯胶水层),不要在这里写业务逻辑 |

也就是说:**业务代码写在 `logic/`,任何你想自己维护的结构体都别放进 `model/gen/`**。

## 路由冲突检测

gin 有两条硬约束,生成器在**写文件之前**就会校验:

1. 同一 METHOD + PATH 只能注册一次
2. 同一层路径段上的通配符**名字必须一致**(`/users/:userId` 与 `/users/:id/detail` 不能共存)

本次配置内部的冲突会直接中止生成并列出先/后注册方。
一个工程常由多份配置先后生成,各自单独跑都合法、合到一起才冲突 —— 这种情况在装配层追加一次自检:

```go
import "github.com/hishack/go-codegen/pkg/codegen"

if cs, err := codegen.ScanRouteConflicts("internal/router"); err != nil {
    log.Fatal(err)
} else if len(cs) > 0 {
    log.Fatal(codegen.FormatRouteConflicts(cs))
}
```

---

## 作为库使用

```bash
go get github.com/hishack/go-codegen/pkg/codegen
```

```go
import "github.com/hishack/go-codegen/pkg/codegen"

report, err := codegen.Run("gen.yaml")
if err != nil {
    return err
}
fmt.Println(report)   // yaml 文件 4 个,接口 28 个;新增 0,更新 0,签名冲突 0

if report.HasWarnings() {
    for _, w := range report.Warnings {
        log.Println("[codegen]", w)
    }
}
```

`pkg/codegen` 是本模块**唯一**的对外入口(`internal/` 下的实现按 Go 规范不对外开放):

| API | 说明 |
| --- | --- |
| `Run(cfgPath string) (*Report, error)` | 读配置 + 生成,最常用 |
| `LoadConfig(cfgPath string) (*Config, error)` | 只加载配置,相对路径已按配置文件目录展开 |
| `NewConfig(cfg Config, baseDir string) (*Config, error)` | 用代码构造配置,便于从环境变量拼装 |
| `Generate(cfg *Config) (*Report, error)` | 用已加载的配置生成,适合 `go:generate` / CI |
| `ScanRouteConflicts(routerDir string) ([]RouteConflictInfo, error)` | 扫描已生成的路由文件,返回冲突 |
| `FormatRouteConflicts(cs []RouteConflictInfo) string` | 把冲突格式化成可直接打印的多行文本 |
| `Report.String()` / `Report.HasWarnings()` | 摘要 / 是否需人工关注 |

配合 `go:generate`:

```go
//go:generate go run github.com/hishack/go-codegen/cmd/gen@latest -config gen.yaml
```

---

## 常见问题

**Q:`go get` 报 404,提示 `could not read Username for 'https://github.com'`**

仓库是公开的,`could not read Username` 不等于私有仓库 —— 更常见的是代理侧缓存了旧的失败结果。
先匿名确认一次:

```bash
git ls-remote https://github.com/hishack/go-codegen   # 能列出 ref 就是公开的
```

然后换一个代理(逗号表示按序回退,前一个返回 404/410 会自动试下一个):

```bash
go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
go get github.com/hishack/go-codegen/pkg/codegen
```

**Q:生成成功,但 `go build` 编译不过,提示找不到 `pkg/response` / `pkg/errorutil`**

目标工程缺这两个包,或者 `gen.yaml` 里的 `module` 写错(import 前缀是拿它拼的)。
把 [`example/pkg/`](example/pkg) 下的两个目录复制过去即可,若你的 module 不叫 `myapp` 记得改 `module` 配置。

**Q:前端 TS 报找不到 `@/utils/request` 或 `@/types/api`**

生成的 `web/src/api/*.ts` 默认用这两处:前者是 axios 实例,后者需导出 `ApiResp<T>` 泛型。
换路径的话,生成后统一替换一下 import 即可(`webOut` 留空可完全跳过前端生成)。

**Q:重复运行会不会把代码插两遍?**

不会。写入是 `go/ast` 定位 + 文本拼接,已存在的声明会被原地更新,重复运行输出一致(幂等),
第二次跑会显示 `新增 0,更新 0`。

**Q:yaml 里删掉一个接口,代码为什么不删?**

删除是不可逆操作,生成器只打印告警让你手动清理。这是有意的安全策略。

**Q:启动时 gin panic `handlers are already registered for path ...`**

手工改过路由,或不同配置的 `group` 前缀不一致。用 `codegen.ScanRouteConflicts` 查一遍(见上一节)。

**Q:能不能只生成某一层?**

不能。七层是同一次生成的产物,增量机制保证只改动有变化的部分。

---

## 开发

```bash
go test ./...        # 解析器 + 合并引擎单元测试 + 增量场景 e2e
```

目录结构:

```
cmd/gen/            命令行入口(只是 pkg/codegen 的薄封装)
config/             配置加载与校验
pkg/codegen/        对外公开 API 门面
internal/field/     类型模型(Type/Field/Struct/Api/Registry)
internal/parse/     yaml 解析:注释提取、类型校验、嵌套提升
internal/merge/     AST 增量写入引擎(文本拼接 + import 修补,支持别名 import)
internal/handler/   生成器注册表 + 六个生成器 + 命名约定 + 场景化 e2e 测试
yamlPath/           测试用例(基础类型、数组、深嵌套、边界、文件上传)
example/            演示目标工程(独立 module,可编译)
docs/               设计笔记与模板参考
```

扩展新的产出物(CRUD 实现、middleware、文档等)只要实现 `handler.Generator` 并注册,
不必改动既有生成器:

```go
type Generator interface {
    Name() string
    Generate(ctx *GenContext, report *Report)
}

RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewMyGen(cfg) })
```

---

## 许可

本仓库暂未附带 LICENSE 文件。如需在团队或公司内使用,请先与作者确认授权方式。
