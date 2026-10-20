---
name: gin-api-yaml
description: 为 go-codegen 编写领域 YAML 接口定义(一个 yaml = 一个领域,一次生成 request/response/service/logic/handler/router/TS 七层)。当用户要新增或修改接口、写 codegen yaml、定义请求响应结构体、排查 yaml 生成报错时使用。触发词:写yaml、加接口、新增api、codegen、生成接口、yaml 报错。
---

# 写 go-codegen 的领域 YAML

一个文件 = 一个领域模块,每个顶层 key 是一个接口,接口里只有 `method` / `path` / `req` / `res`。
命名、结构体、绑定、校验、swagger、TS 全部推导出来 —— 你只写**接口语义**和**注释**。

## 骨架

```yaml
filename: user      # 领域名:决定 user_service.go / user_router.go ... 的文件名
package: system     # swagger @Tags

# 创建用户                          <- 接口注释:进 @Summary/@Description/TS jsdoc
CreateUser:
  method: POST
  path: /users                      # 长说明写这里行尾;path 不带 /api/v1 前缀
  req:
    name: string # 用户名            <- 字段注释:进 Go 字段注释与 @Param 描述
    age: int # 年龄
  res:
    id: int64 # 主键(单个基础字段 → service 直接返回 int64,不生成 Res 结构体)
```

## 硬规则(违反即失败)

| 规则 | 说明 |
| --- | --- |
| `filename` 匹配 `^[a-z][a-zA-Z0-9]*$` | `user_profile` 非法,只能 `userProfile` |
| 接口名匹配 `^[A-Z][a-zA-Z0-9]*$` | `createUser` 非法 |
| 字段名匹配 `^[a-z][a-zA-Z0-9]*$` | `user_name` 非法 |
| 接口里只允许 5 个 key | method / path / req / res / filename,多一个 `desc` 就报错 |
| `method` 只有 5 种 | get/post/put/patch/delete |
| `path` 以 `/` 开头 | 不能含空格与 `?`,query 用 req 表达 |
| 数组只写单元素 | `[int]` 可以;`[]int` 是 YAML 语法错;`[int, string]` 非法 |
| `map` 只能 `map[基础类型]基础类型` | 不能嵌套结构体或切片 |
| `file` 只能出现在 `req` | res 里写 file 直接报错 |
| 接口名 / filename 全局唯一 | 跨文件也不能重名(api/service 是单包) |
| 同配置内路由不得冲突 | 同 method+path 只能一次;同层通配符必须同名 |

## 语义要点(不报错,但结果可能不是你要的)

1. **`res` 只有一个基础字段 → 标量响应**:`res: { id: int64 }` 的签名是 `(int64, error)`,
   不生成 `XxxRes`。要结构体就至少两个字段,或放一个嵌套对象/数组。
2. **路径参数不并入 req**:`/orders/:orderId(int64)` 是独立形参
   `CancelOrder(c *gin.Context, orderId int64, req request.CancelOrderReq)`,别在 req 里重复写。
3. **绑定方式**:GET/DELETE → `ShouldBindQuery`;其余 → `ShouldBindJSON`;req 有**标量** `file` → multipart。
4. **上传必须有标量 `file` 字段**。只写 `[file]` 或把 file 放进对象数组,会退化成 JSON 绑定且漏
   `mime/multipart` import(实测)。多文件时保留一个 `file: file`。
5. **嵌套对象提升为命名结构体**:`<父><字段名>`,对象数组加 `Item`(`res.list[]` → `XxxResListItem`)。
   不能递归,树形只能写有限层。
6. **`binding` 标签自动生成**:`name`/`title`/`userName`/`password`/`phone`/`email`(按末段单词)为
   `required,max=255`,其他 string 为 `omitempty,max=255`,数值 `omitempty,gte=0`,嵌套/切片 `omitempty,dive`。
7. **注释**:接口注释只取一处(上方优先);字段注释会把上方与行尾合并成一行。
   不要用中文弯引号 `“”`,会破坏 swagger 注解配对。
8. **删接口不删代码**,只告警;改签名后 logic 保留旧实现,需手工对齐。
9. **接口级 `filename` 只用于拆分领域**,不解决重名,也不能与已有 yaml 文件名相同。

## 类型

| yaml | Go | TS |
| --- | --- | --- |
| `string` `int` `int64` `uint8` `float64` `bool` `byte` `rune` | 原样 | `string` / `number` / `boolean` |
| `time.Time` / `time.Duration` | 原样(自动 import time) | `string` / `number` |
| `any` | `any` | `unknown` |
| `map[string]string` | 同 | `Record<string,string>` |
| `[int]` | `[]int` | `number[]` |
| `- id: int64` | `[]*XxxItem` | `XxxItem[]` |
| 嵌套 mapping | `*Xxx` | interface |
| `file` | `*multipart.FileHeader` | `File` |

路径参数字段类型只能 `string/bool/int*/uint*/float32/float64`(不支持 time.Time、any、byte、rune),
`ctx/c/err/req/res` 不能当参数名。

## 生成与报错

```bash
go run ./cmd/gen -config gen.yaml
# ✔ 生成完成:yaml 文件 1 个,接口 8 个;新增 33,更新 0,签名冲突 0
```

重复跑是幂等的(`新增 0,更新 0`);解析失败会中止,不会写坏文件。
报错格式 `磁盘文件名: 接口名.位置: 原因`,常见几类:

| 报错 | 修法 |
| --- | --- |
| `接口名 "createUser" 非法:要求大驼峰` / `字段名 "user_name" 非法` | 改驼峰 |
| `filename "user_profile" 非法:要求小驼峰单词条` | 改 `userProfile` |
| `字段 "desc": 不支持的字段类型 "说明文字"` | 补类型:`desc: string # 说明文字` |
| `数组写法只支持单元素` / `数组元素必须是类型或对象定义` | 去掉多元素或二维数组 |
| `file 类型只能用于 req 字段` | 下载走 `c.File`/`c.Data`,不进结构体 |
| `路径参数 :id(time.Time) 类型不支持` | 改成 `int64` |
| `路径参数 "req" 是保留变量名` | 换名,如 `userId` |
| `接口名 Xxx 重复(a 与 b)` | 接口名全局唯一,改名 |
| `[duplicate] 路由 GET /same 被定义了 2 次` | 同 method+path 只留一处 |
| `[wildcard] ... :id 与 :userId` | 同前缀通配符统一命名(推荐 `:xxxId`) |
| `did not find expected key` | 多半写了 `[]int`,改 `[int]`;或 Tab 混用 |
| `undefined: multipart` | 上传接口漏了标量 `file` 字段 |

## 常用写法

```yaml
# 分页列表:req 用 page/pageSize,res 用 total/list
ListItems:
  method: GET
  path: /items
  req:
    keyword: string # 搜索关键字
    page: int # 页码,从 1 开始
    pageSize: int # 每页条数
  res:
    total: int64 # 总条数
    list:
      - id: int64 # 商品 id
        title: string # 标题

# 动作型:不需要返回值就省略 res
LogoutAccount:
  method: POST
  path: /auth/logout
  req:
    token: string # 令牌

# 上传:保留一个标量 file 字段
UploadItemImage:
  method: POST
  path: /items/images
  req:
    itemId: int64 # 商品 id
    file: file # 图片文件
  res:
    url: string # 访问地址
```

## 交付前自检

- [ ] filename / 接口名 / 字段名都是驼峰,接口名与既有接口不重复
- [ ] path 不带 group 前缀;同层通配符与同目录既有接口同名
- [ ] 路径参数写了类型标注,且没在 req 里重复声明
- [ ] 列表统一 `page`/`pageSize`/`total`/`list`
- [ ] `res` 单字段确认要的是标量;上传接口有标量 `file`
- [ ] 字段都有行尾注释;跑过生成并看过 service 签名

更完整的样例可对照 `example/cmd/gen/yaml/threshold.yaml`(注释透传)与 `yamlPath/*.yaml`(边界写法)。
