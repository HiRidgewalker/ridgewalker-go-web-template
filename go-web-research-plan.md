# Go Web 技术路线与项目架构调研 / 落地计划

- 调研时间：2026-10-18
- 项目定位：`learning-go-web-notebook`，一个用于学习 Go Web 开发的笔记类 REST API 服务
- 调研目标：选择 **Go 技术圈主流、维护活跃、层次清晰** 的技术栈与目录布局，避免职能代码混杂

---

## 1. 结论摘要（TL;DR）

| 维度 | 选型 | 一句话理由 |
|---|---|---|
| Go 版本 | Go 1.27.x（当前稳定版 1.27.2） | 官方最新稳定线；`go.mod` 声明版本，`GOTOOLCHAIN=auto` 按需切换 |
| 路由 / HTTP | `net/http` + `go-chi/chi v5` | 完全兼容标准库 `http.Handler`，零外部依赖，中间件生态成熟 |
| 数据库 | PostgreSQL + `jackc/pgx v5` | Go 生态事实标准的 PostgreSQL 驱动与连接池 |
| 数据访问 | `sqlc`（SQL 优先，代码生成） | 类型安全、SQL 可见、无 ORM 隐式行为，契合分层架构 |
| 迁移 | `pressly/goose` | 单文件 up/down、默认事务执行、支持 `go:embed` |
| 配置 | 环境变量 + `caarlos0/env` | 12-Factor，无配置框架负担；比 Viper 更轻 |
| 日志 | 标准库 `log/slog` | Go 1.21+ 官方结构化日志，无第三方依赖 |
| 校验 | `go-playground/validator v10`（仅传输层） | 业内使用最广；领域不变量在 domain 内手写校验 |
| 测试 | `testing` + `httptest` + `testcontainers-go` | 标准库为主，集成测试跑真实 PostgreSQL |
| 架构布局 | 官方 `cmd/` + `internal/`，内部分层（Ports & Adapters） | `cmd` / `internal` 是官方推荐；分层保证依赖单向 |
| 依赖注入 | 构造函数手工装配 | Go 社区主流做法，显式、可读；不引入 DI 框架 |
| Lint | `golangci-lint` + `depguard` | 用规则机器化约束分层依赖 |
| 热重载 | `air` | 开发期主流工具 |

**明确不采用**：`pkg/` 目录（非官方、易变成杂物间）、`models/` 大杂烩、handler 直连数据库、框架式 DI、Fiber（基于 fasthttp，脱离标准库生态）。

---

## 2. 调研方法与判定标准

本次调研以「主流度」为第一优先级，判定依据按权重排序：

1. **官方文档与标准库能力**：go.dev 官方指引优先于社区模板
2. **开发者调查数据**：Go Developer Survey、JetBrains Go 生态报告
3. **生产环境采用情况**：知名项目 / 公司公开使用记录
4. **维护活跃度**：提交频率、issue 响应、版本发布节奏
5. **与标准库的兼容性**：能否使用 `http.Handler`、`context.Context` 等通用抽象
6. **学习价值**：是否帮助理解 Go 的惯用法，而不是用框架遮蔽语言特性

> 说明：GitHub Star 只能作参考，不能单独作为主流度依据（不同工具检索时间点不同，且存在刷星/历史积累偏差）。

---

## 3. Web 框架 / 路由调研

### 3.1 采用数据

JetBrains《The Go Ecosystem in 2025》（2025-11）针对以 Go 为主语言的开发者、多选统计：

| 方案 | 采用率 |
|---|---|
| **标准库 `net/http`** | 最常用的路由选择 |
| Gin | 48% |
| Echo | 16% |
| Chi | ~12% |
| Fiber | 11% |

Go 1.22（2024-02）为 `net/http` 增加了方法匹配、路径通配符与 `Request.PathValue`，大量项目已不再需要第三方路由。

### 3.2 候选对比

| 方案 | 模型 | 优点 | 代价 / 风险 |
|---|---|---|---|
| `net/http`（1.22+） | 标准库 | 零依赖、稳定、知识可迁移 | 中间件、路由分组需自己组织 |
| **chi** | 标准库 `http.Handler` | 零依赖、100% 标准库兼容、分组/中间件清晰 | 功能克制，无内建参数绑定 |
| Gin | 自有 `*gin.Context` | 使用最广、教程多、JSON 绑定/渲染齐全 | 上下文非标准，容易向 handler 外泄漏 |
| Echo | 自有 Context，兼容 `http.Handler` | 功能全、错误处理集中 | 生态小于 Gin |
| Fiber | fasthttp | 基准测试快 | 不兼容 `net/http` 生态；Context 复用有生命周期陷阱 |

### 3.3 决策

**采用 `chi v5`，构建在标准库 `net/http` 之上。**

理由：

1. handler 是标准 `http.HandlerFunc`，测试用 `httptest` 即可，无框架耦合；
2. 中间件就是标准 `func(http.Handler) http.Handler`，第三方标准中间件可直接用；
3. 路由分组、URL 参数、超时/恢复中间件齐全，足够支撑本项目；
4. 若将来要去掉 chi，替换成本极低。

**备选**：如果更看重「教程数量 / 生态 / 招聘关键词」，改用 Gin。Gin 使用率最高，但需接受 `*gin.Context` 只存在于 adapter 层，禁止传入 usecase。

**不采用 Fiber**：它基于 fasthttp，无法复用标准库中间件与 `http.Request` 工具链，且 Context 对象会被复用，容易踩坑。

---

## 4. 数据库与数据访问调研

### 4.1 驱动

**采用 `jackc/pgx v5` + `pgxpool`。** 这是 Go 访问 PostgreSQL 的事实标准，性能与类型支持优于 `database/sql` 通用驱动。GORM 内部也使用 pgx。

### 4.2 数据访问方式

| 方案 | 定位 | 优点 | 代价 |
|---|---|---|---|
| **sqlc** | 写 SQL，生成类型安全 Go 代码 | SQL 完全可见、编译期类型检查、无运行时魔法 | 动态查询/查询组合较弱，需代码生成步骤 |
| GORM | 全功能 ORM | 使用最广、上手快、CRUD 自动化 | 复杂查询抽象泄漏、隐式行为多、性能不可控 |
| ent | 图式 ORM + 代码生成 | schema 即代码、类型安全 | 学习曲线陡、框架绑定强 |

调研结论：GORM 是「默认大家都会用」的 ORM；pgx 是 PostgreSQL 的标准驱动；sqlc 在生产环境中被大量采用，但更偏「SQL 优先」的团队，属于主流而非小众。

### 4.3 决策

**采用 sqlc。** 理由：

- 与分层架构天然契合：生成的代码只属于 `adapter/postgres`，usecase 只依赖 `domain.Repository` 接口；
- 学习价值高：SQL、事务、索引、N+1 等问题不会被 ORM 隐藏；
- 编译期类型安全，重构时错误暴露更早。

**备选**：若你希望先快速做出 CRUD、减少代码生成步骤，可改用 GORM；但 Repository 接口与分层边界保持不变，未来可替换。

### 4.4 迁移工具

| 工具 | 定位 | 备注 |
|---|---|---|
| **goose** | CLI + Go 库，SQL/Go 迁移 | 默认事务执行、单文件 up/down、支持 embed |
| golang-migrate | CLI + Go 库，up/down 文件 | Star 更多、数据库支持最广；失败可能留下 dirty 版本 |
| Atlas | 声明式 schema 迁移 | 能力强，但引入 HCL/声明式模型，学习成本高 |

**采用 goose。** 支持 `//go:embed` 把迁移打进二进制，部署时可自动执行；对学习项目足够且行为直观。

---

## 5. 横向基础组件调研

| 能力 | 选型 | 说明 |
|---|---|---|
| 配置 | 环境变量 + `caarlos0/env/v11` | 结构体 tag 映射；`.env` 仅本地开发用，生产用真实环境变量 |
| 日志 | `log/slog` | dev 用 `TextHandler`，prod 用 `JSONHandler`；请求日志中间件注入 request_id |
| 校验 | `go-playground/validator/v10` | 只校验传输层 DTO；领域不变量在 domain 层手写 |
| 错误 | 哨兵错误 + `errors.Is/As` | domain 定义 `ErrNotFound`/`ErrConflict`/`ErrValidation`，adapter 映射为 HTTP 状态码 |
| 认证（后续） | `golang-jwt/jwt/v5` + `bcrypt`/`argon2id` | 不进入 MVP，预留中间件位置 |
| API 契约（可选） | OpenAPI 3.1 + `oapi-codegen` | 后续可改为契约优先；MVP 用手写 DTO |
| 测试 | `testing`、`httptest`、`testcontainers-go` | 单元测试用假 Repository；集成测试跑真实 Postgres |
| 断言库（可选） | `testify` | 使用广泛；也可只用标准库 `t.Fatalf` |
| Lint | `golangci-lint` | 启用 `govet`、`errcheck`、`staticcheck`、`depguard` |
| 热重载 | `air` | 本地开发自动重启 |
| 可观测性（后续） | OpenTelemetry + `prometheus/client_golang` | MVP 不做，避免过度设计 |
| 容器化 | Dockerfile（多阶段）+ docker-compose | compose 只跑 PostgreSQL，便于本地开发 |

---

## 6. 项目架构调研

### 6.1 官方布局说了什么

Go 官方《Organizing a Go module》明确推荐：

- **`cmd/`**：存放可执行命令，尤其适合「既有命令又有可导入包」的仓库；
  > "A common convention is placing all commands in a repository into a `cmd` directory"
- **`internal/`**：存放不希望被外部模块导入的包；
  > "It's recommended to keep packages in `internal` as much as possible."
- **`pkg/`**：官方文档**完全没有提及**，既未推荐也未禁止。

社区流传甚广的 `golang-standards/project-layout` 在 README 中明确声明自己**不是官方标准**，其中 `pkg/` 也「未被普遍接受」。

**结论：`cmd/` + `internal/` 是可以放心使用的共识；`pkg/` 不用。**

### 6.2 分层方式：层级优先 vs 领域优先

| 方式 | 结构 | 优点 | 缺点 |
|---|---|---|---|
| 层级优先 | `internal/domain`、`internal/usecase`、`internal/adapter/...` | 层次一眼可见，依赖规则清晰，适合学习与中小项目 | 业务变多后单层目录会变长 |
| 领域优先 | `internal/<domain>/{domain,usecase,adapter}` | 领域内高内聚，适合多限界上下文 | 跨领域依赖与共享基础设施易混乱 |

本项目是单一限界上下文（笔记），采用 **层级优先 + 端口与适配器（Ports & Adapters）**，层次最清晰。

### 6.3 依赖规则（必须机器化约束）

```
cmd  ──────────────► 所有层（仅装配，不含业务逻辑）
adapter/http ──────► usecase, domain
adapter/postgres ──► domain
usecase ───────────► domain
domain ────────────► 仅标准库
```

硬性规则：

1. `domain` 不 import 任何本项目内部包，也不 import pgx/chi 等基础设施库；
2. `usecase` 只 import `domain` 与标准库；
3. `adapter` 之间禁止互相 import；
4. `domain` 实体不得直接作为 HTTP 响应返回，必须有 DTO；
5. usecase 不得感知 `http.Request`、`http.ResponseWriter`、SQL 驱动；
6. 用 `golangci-lint` 的 `depguard` 把上述规则写进配置，而不是靠自觉。

### 6.4 目录结构

```text
learning-go-web-notebook/
├── cmd/
│   └── api/
│       └── main.go                 # 唯一入口：读配置 → 初始化依赖 → 启动/优雅关闭
├── internal/
│   ├── domain/                     # 最内层：纯业务模型，无基础设施依赖
│   │   ├── notebook.go             #   实体、值对象、领域行为
│   │   ├── errors.go               #   ErrNotFound / ErrConflict / ErrValidation
│   │   └── repository.go           #   Repository 接口（端口）
│   ├── usecase/                    # 应用层：编排业务流程
│   │   ├── notebook.go             #   创建/查询/更新/删除用例
│   │   └── notebook_test.go
│   ├── adapter/                    # 最外层：与外界交互
│   │   ├── http/
│   │   │   ├── router.go           #   路由注册、中间件挂载
│   │   │   ├── middleware.go       #   request_id、recover、访问日志、超时
│   │   │   ├── notebook.go         #   handler：解码 DTO → 调 usecase → 编码响应
│   │   │   ├── dto.go              #   请求/响应结构体与校验 tag
│   │   │   └── errors.go           #   领域错误 → HTTP 状态码
│   │   └── postgres/
│   │       ├── db.go               #   pgxpool 初始化、ping
│   │       ├── notebook.go         #   Repository 接口实现
│   │       └── sqlc/               #   sqlc 生成代码（只允许本包引用）
│   ├── config/
│   │   └── config.go               #   环境变量 → Config 结构体
│   └── platform/
│       ├── logger/
│       │   └── logger.go           #   slog 初始化
│       └── server/
│           └── server.go           #   http.Server 配置、超时、优雅关闭
├── db/
│   ├── migrations/                 # goose SQL 迁移
│   └── queries/                    # sqlc 输入 .sql
├── api/
│   └── openapi.yaml                # 可选：后续契约优先时启用
├── configs/
│   └── .env.example
├── scripts/
│   ├── dev.sh
│   └── migrate.sh
├── .golangci.yml
├── Makefile
├── docker-compose.yml              # 本地 PostgreSQL
├── sqlc.yaml
├── go.mod
├── go.sum
└── README.md
```

### 6.5 一次请求的流转

```text
HTTP 请求
  → chi Router / middleware（request_id、recover、日志、超时）
  → adapter/http handler
      · 解码 JSON → DTO
      · validator 校验请求形状
  → usecase（业务规则、事务边界、权限判断）
      · 只依赖 domain.Repository 接口
  → adapter/postgres（pgx / sqlc）
  → 返回领域对象 / 领域错误
  → handler 把领域对象映射为 DTO，把领域错误映射为 HTTP 状态码
  → JSON 响应
```

### 6.6 各层职责边界（防混杂清单）

| 层 | 应该做 | 绝对不要做 |
|---|---|---|
| `domain` | 实体、不变量、领域错误、Repository 接口 | 引用 chi/pgx/slog；写 HTTP 状态码；写 SQL |
| `usecase` | 业务编排、事务边界、调用 Repository | 解析 HTTP 请求；拼 SQL；返回 DTO |
| `adapter/http` | 路由、中间件、DTO、参数校验、错误映射 | 直接访问数据库；写业务规则 |
| `adapter/postgres` | SQL、扫描行、事务实现 | 写业务规则；返回 HTTP 概念 |
| `cmd/api` | 依赖装配、启动、优雅关闭 | 任何业务逻辑 |
| `platform` | 日志、服务器、通用基础设施 | 业务概念 |

---

## 7. 落地实施计划

### Phase 0：工具链与骨架（半天）

1. 安装 Go（推荐 `brew install go`，当前 1.27.x）；
2. `go mod init github.com/<your-name>/learning-go-web-notebook`（替换为真实仓库路径）；
3. 在 `go.mod` 中写 `go 1.27.0` 与 `toolchain go1.27.2`；
4. `git init`，建立 `.gitignore`（`.env`、`bin/`、`tmp/`、`air` 产物）；
5. 创建上述目录骨架与 `Makefile`。

**验收**：`go version`、`go env GOTOOLCHAIN` 正常；`go build ./...` 通过。

### Phase 1：可运行的空服务（半天）

1. `internal/config`：env → Config；
2. `internal/platform/logger`：slog 初始化；
3. `internal/platform/server`：`http.Server`（设置 `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout`、`IdleTimeout`）与 graceful shutdown；
4. chi 路由：`GET /healthz`、`GET /readyz`；
5. 中间件：request_id、recover、访问日志。

**验收**：`make run` 后 curl 健康检查返回 200；Ctrl+C 优雅退出。

### Phase 2：领域与用例（1 天）

1. `domain/notebook.go`：`Notebook` 实体（ID、Title、Content、CreatedAt、UpdatedAt）；
2. `domain/errors.go`：哨兵错误；
3. `domain/repository.go`：`NotebookRepository` 接口；
4. `usecase/notebook.go`：Create / Get / List / Update / Delete；
5. 内存实现 `adapter/memory`（仅测试用或 Phase 3 前联调）；
6. usecase 单元测试，使用假 Repository。

**验收**：用例测试覆盖正常路径与错误路径；domain 包 `go list -deps` 中无第三方包。

### Phase 3：HTTP 接口（1 天）

1. DTO + validator；
2. handler 五件套：`POST/GET(list)/GET(id)/PATCH/DELETE /api/v1/notebooks`；
3. 错误映射：`ErrNotFound → 404`、`ErrConflict → 409`、`ErrValidation → 422`、其他 → 500；
4. `httptest` 覆盖 handler；
5. 统一响应格式（成功数据 + 错误结构）。

**验收**：用 curl 或 `.http` 文件跑通完整 CRUD；错误响应结构一致。

### Phase 4：接入 PostgreSQL（1–2 天）

1. `docker-compose.yml` 启动 PostgreSQL；
2. goose 迁移：`notebooks` 表、索引；
3. sqlc 配置与查询，生成代码；
4. `adapter/postgres` 实现 Repository；
5. `cmd/api/main.go` 完成装配；
6. 集成测试用 `testcontainers-go` 启动真实数据库。

**验收**：重启服务数据仍在；迁移可重复执行；`go test ./...` 全绿。

### Phase 5：工程化（半天）

1. `.golangci.yml` + depguard 分层规则；
2. `air` 热重载；
3. `Makefile`：`run`、`test`、`lint`、`fmt`、`migrate-up`、`migrate-down`、`sqlc`；
4. 多阶段 `Dockerfile`；
5. `README.md`：本地启动步骤、目录说明、依赖规则。

**验收**：`make lint` 无告警；违反分层规则的 import 会被 lint 拦截。

### Phase 6（可选，按需推进）

- JWT 认证与用户表；
- OpenAPI 契约与 `oapi-codegen`；
- OpenTelemetry + Prometheus；
- 分页、搜索、软删除；
- CI（GitHub Actions：test + lint + build）。

---

## 8. 验收标准（Definition of Done）

- [ ] `go build ./...`、`go test ./...`、`go vet ./...`、`golangci-lint run` 全部通过
- [ ] 目录层次与职责边界符合第 6 节，depguard 规则生效
- [ ] `domain` 包不依赖任何第三方库
- [ ] HTTP handler 中不存在 SQL 或业务规则
- [ ] usecase 中不存在 `http.*` 类型
- [ ] 所有数据库迁移可通过 goose 从零重建
- [ ] 健康检查、优雅关闭、请求超时、panic 恢复均可用
- [ ] README 能让他人从零启动项目

---

## 9. 关键决策点与备选方案

| 决策点 | 默认选择 | 何时切换 |
|---|---|---|
| 路由 | chi | 需要最大教程/生态 → Gin |
| 数据访问 | sqlc | 想先快速 CRUD → GORM（Repository 边界不变） |
| 迁移 | goose | 需要最广数据库支持 → golang-migrate |
| 架构 | 层级优先 | 领域数量增长到 3+ → 领域优先 |
| 数据库 | PostgreSQL | 纯本地学习、想零依赖 → SQLite（仍用 sqlc） |
| API 契约 | 手写 DTO | 前后端多人协作 → OpenAPI + oapi-codegen |

---

## 10. 参考资料

- Go 官方《Organizing a Go module》 — https://go.dev/doc/modules/layout
- Go 官方 Toolchain 文档 — https://go.dev/doc/toolchain
- Go 官方下载页（版本信息） — https://go.dev/dl/
- Go 1.22 路由增强 — https://go.dev/blog/routing-enhancements
- JetBrains《The Go Ecosystem in 2025》 — https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/
- 2025 Go Developer Survey — https://go.dev/blog/survey2025
- chi — https://github.com/go-chi/chi
- Gin — https://gin-gonic.com/en/docs/
- Echo — https://github.com/labstack/echo
- Fiber（及 fasthttp 注意事项） — https://docs.gofiber.io/
- pgx — https://github.com/jackc/pgx
- sqlc — https://sqlc.dev/
- GORM — https://gorm.io/docs/
- goose — https://github.com/pressly/goose
- golang-migrate — https://github.com/golang-migrate/migrate
- Atlas 迁移工具对比 — https://atlasgo.io/
- 社区布局参考（非官方） — https://github.com/golang-standards/project-layout
