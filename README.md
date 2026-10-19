# go-codegen

写 YAML,一键生成 Go Web 脚手架(Gin)。

参考 gen_review_resource 的 model/gen 分层,一个 yaml 文件描述一个领域模块的全部接口,
运行生成器即可**增量**生成 / 更新七层代码,yaml 里写的注释会一路带到生成的代码里。

## 生成的七层代码

以 `yamlPath/user.yaml` 里的 `CreateUser` 接口为例:

| 层 | 文件 | 生成内容 |
| --- | --- | --- |
| model/request | `internal/model/gen/request/user.go` | 请求结构体 `CreateUserReq`(`json`+`form` 双 tag) |
| model/response | `internal/model/gen/response/user.go` | 响应结构体 `CreateUserRes`(嵌套对象自动提升命名) |
| service | `internal/service/user_service.go` | 接口 `IUserService` 及方法签名 |
| logic | `internal/logic/user_logic.go` | 空实现 `UserService`(带 TODO,写业务逻辑的地方) |
| handler | `internal/handler/user_handler.go` | `UserHandler` 胶水方法:ShouldBind -> 调 service -> 包装响应,含 swagger 注解 |
| router | `internal/router/user_router.go` | `UserRouter.InitUserRouter(Router *gin.RouterGroup, ...)` 路由注册 |
| web | `web/src/api/user.ts` | TS 类型 + `UserApi` 请求类(axios) |

## 快速开始

```bash
# 1. 配置 gen.yaml(输入 yaml 目录 + 目标工程各层目录)
# 2. 写 yaml
# 3. 生成
go run ./cmd/gen -config gen.yaml
```

本仓库自带可编译的演示工程 `example/`(独立 go module):

```bash
go run ./cmd/gen -config gen.yaml   # 生成到 example/
cd example && go build ./...        # 直接编译通过
```

也可以走 `example/` 自己的入口(它是以库的方式引用本模块的示例):

```bash
cd example
go run ./cmd/gen -config cmd/gen/gen.yaml   # 注释与多领域示例(默认就是这个)
go run ./cmd/gen -config ../gen.yaml        # 主测试集
```

`example/cmd/gen/yaml/threshold.yaml` 是一份**带完整注释的教学样例**,
演示了注释如何一路带到 Go 字段注释、swagger `@Summary` 与前端 TS jsdoc,
以及接口级 `filename` 如何把一个 yaml 拆成多个领域。
里面的每个字段都写了 `#` 注释,可以直接对照生成结果看效果。

## 在自己的项目里引用

> ⚠️ 本仓库是 **私有仓库**,`GOPROXY`(如 goproxy.cn)无法匿名代理,
> 直接 `go get` 会报 `404 invalid version ... could not read Username for 'https://github.com'`。
> 这不是包的问题,是私有仓库的鉴权问题,按下面两种方式之一处理即可。

**方式一:仓库内 / 联调(推荐,零配置)**

```bash
go mod edit -require=github.com/hishack/go-codegen@v0.0.0-00010101000000-000000000000
go mod edit -replace=github.com/hishack/go-codegen=../   # 指向本地仓库路径
go mod tidy
```

`example/` 就是这么配置的,克隆下来即可 `go build ./...`。

**方式二:私有仓库鉴权拉取**

```bash
go env -w GOPRIVATE=github.com/hishack/*     # 跳过代理与 sumdb 校验
go env -w GONOSUMDB=github.com/hishack/*
go get github.com/hishack/go-codegen@latest
```

或临时使用:

```bash
GOPRIVATE='github.com/hishack/*' GOSUMDB=off GOPROXY=direct \
  go get github.com/hishack/go-codegen@latest
```

**以库的方式调用**

```go
import "github.com/hishack/go-codegen/pkg/codegen"

report, err := codegen.Run("gen.yaml")
if err != nil {
    return err
}
fmt.Println(report) // yaml 文件 4 个,接口 28 个;新增 0,更新 0,签名冲突 0
```

`pkg/codegen` 是本模块唯一的对外入口(`internal/` 下的实现按 Go 规范不对外开放)。
`codegen.Run` 会把配置文件里的相对路径按**配置文件所在目录**解析,因此在任何工作目录下执行结果都一致。

## yaml 语法

```yaml
filename: user        # 领域名(小驼峰,决定生成的文件名)
package: system       # swagger @Tags 等元数据

# 创建用户              <- 接口注释,会进入 @Summary / 文档注释 / TS jsdoc
CreateUser:
  method: POST        # get/post/put/patch/delete,不区分大小写
  path: /users/:orgId(int64)   # 路径参数支持类型标注,不写默认 string
  req:                # 请求体(POST 走 JSON,GET 走 query,含 file 走 multipart)
    name: string # 用户名          <- 字段注释
    age: int
    tags: [string]                 # 数组:yaml 里写 [int] 这种单元素写法
    avatar: file                   # 文件上传 -> *multipart.FileHeader
    extra:                         # 嵌套对象 -> 提升为命名结构体
      key: string
  res:                # 响应(可省略)
    id: int64
    profile:
      nickname: string
```

要点:

- `req`/`res` 都可省略,`req: {}` 等价于没写;
- 字段类型:基础类型白名单(string/int 家族/uint 家族/float/bool/time.Time/time.Duration/any)、
  `map[k]v`、`[元素类型]` 数组、任意深度嵌套对象、`file`(只能用于 req);
- `res` 只有**一个基础类型字段**时不生成结构体,service 直接返回该类型(如 `(bool, error)`);
- 接口名全局唯一(跨 yaml 文件也不允许重名,api/service 是单包)。

## 绑定与文件传输

handler 胶水按请求形态自动选择 gin 绑定方式:

| req 形态 | 绑定 |
| --- | --- |
| 含 `file` 字段 | `c.ShouldBind(&req)`(multipart 表单) |
| GET / DELETE | `c.ShouldBindQuery(&req)`(form 标签) |
| POST / PUT / PATCH | `c.ShouldBindJSON(&req)` |

文件下载是纯响应行为(不产生结构体),按 gin 惯例在 logic 返回内容、
handler 里用 `c.File` / `c.Data` 输出即可,生成器不干预。

## 自动生成的校验与状态码

请求结构体会按字段类型自动带上 `binding` 标签,无需在 yaml 里声明:

| 字段形态 | 生成的标签 |
| --- | --- |
| 关键标识(`name` / `title` / `userName` / `password` / `phone` / `email`) | `required,max=255` |
| 其他 string | `omitempty,max=255` |
| 整型 / 浮点 | `omitempty,gte=0` |
| bool | `omitempty` |
| 嵌套对象 / 切片 | `omitempty,dive` |
| 时间 / file | `omitempty` |

handler 按错误来源返回不同状态码,不再一律 200:

```go
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ...)   // 参数/绑定错误 -> 400
    return
}
data, err := u.userService.CreateUser(c, req)
if err != nil {
    c.JSON(errorutil.HTTPStatusOf(err), ...)   // 业务失败 -> 422
    return
}
c.JSON(http.StatusOK, ...)                // 成功 -> 200
```

`errorutil` 需要提供 `HTTPStatusOf(err) int`,参考
`example/pkg/errorutil/business_error.go`。

## 命名约定

以 user.yaml + 接口 CreateUser:

```
request.CreateUserReq / response.CreateUserRes     (model/gen 两分包)
IUserService.CreateUser                            UserService.CreateUser(logic)
UserHandler.CreateUser                             UserRouter.InitUserRouter
UserApi.createUser(TS)
```

response 包在 handler/service/logic 里统一以 `res` 别名导入,
与 `pkg/response`(统一响应封装)不冲突。

## 增量生成(核心特性)

写入不是整体覆盖,而是 `go/ast` 定位 + 文本拼接(`internal/merge`),
没碰到的用户代码一字节不变,重复运行输出完全一致(幂等):

| 场景 | 行为 |
| --- | --- |
| yaml 新增接口 | 所有层自动补齐 |
| yaml 改 req/res 字段 | request/response 结构体自动同步;service 签名不变时 logic 不动 |
| yaml 改接口签名 | 接口签名更新;handler/路由重写;**logic 保留旧实现并打印告警** |
| yaml 删除接口 | 不删代码,打印告警提醒手动清理 |
| 用户手写的 logic 业务逻辑、结构体字段、自定义函数/类型 | **永不触碰** |
| handler 方法体 | 由生成器管理(纯胶水层),自定义逻辑请写在 logic 层 |

以上场景全部有自动化测试覆盖:`internal/handler/scenario_test.go`。

## 扩展

生成流程基于接口组装,新增产出物(如 CRUD 实现、middleware、文档)实现
`handler.Generator` 接口并 `RegisterGenerator` 注册即可,无需改动既有生成器:

```go
type Generator interface {
    Name() string
    Generate(ctx *GenContext, report *Report)
}
RegisterGenerator(func(cfg *config.CodeGenConfig) Generator { return NewMyGen(cfg) })
```

## 配置项(gen.yaml)

| 字段 | 说明 |
| --- | --- |
| `yaml` | 输入的 yaml 目录 |
| `module` | 目标工程 go module 名(拼 import 路径) |
| `group` | 路由分组前缀(可选,如 `/api/v1`),生成路由、swagger、前端 url 时统一加上 |
| `serviceOut` / `modelOut` / `logicOut` / `handlerOut` / `routerOut` | 目标工程各层目录(modelOut 下自动分 request/ response/ 子包) |
| `webOut` | 前端 ts 目录(留空跳过) |
| `author` / `github` | 写入生成文件头 |

`group` 的写法很自由,`api/v1`、`/api/v1/`、`/api/v1` 都会被归一化成 `/api/v1`。
同一目标工程的各份配置应使用一致的前缀,否则跨配置冲突检测会失效。

## 目录结构

```
cmd/gen/            命令行入口
config/             配置加载
internal/field/     类型模型(Type/Field/Struct/Api/Registry)
internal/parse/     yaml 解析:注释提取、类型校验、嵌套提升
internal/merge/     AST 增量写入引擎(文本拼接 + import 修补,支持别名 import)
internal/handler/   生成器注册表 + 六个生成器 + 命名约定 + 场景化 e2e 测试
yamlPath/           测试用例(4 个 yaml:基础类型、数组、深嵌套、边界、文件上传)
example/            演示目标工程(独立 module,可编译)
docs/               设计笔记与模板参考
```

## 测试

```bash
go test ./...        # 解析器 + 合并引擎单元测试 + 增量场景 e2e
```
