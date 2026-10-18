# learning-go-web-notebook

一个用于学习 Go Web 开发的笔记类 REST API 服务，按照 [`go-web-research-plan.md`](./go-web-research-plan.md) 中的技术路线与分层架构落地。

## 技术栈

| 能力 | 选型 |
|---|---|
| 语言 | Go 1.27 |
| 路由 / HTTP | `net/http` + `go-chi/chi v5` |
| 数据库 | PostgreSQL 16 + `jackc/pgx v5`（`pgxpool`） |
| 数据访问 | `sqlc`（SQL 优先，编译期类型安全） |
| 迁移 | `pressly/goose`（SQL 迁移，支持 `go:embed`） |
| 配置 | 环境变量 + `caarlos0/env` + `godotenv`（仅本地） |
| 日志 | 标准库 `log/slog` |
| 校验 | `go-playground/validator v10`（仅传输层） |
| 测试 | `testing` + `httptest` + 真实 PostgreSQL 集成测试 |

## 目录结构

```text
.
├── cmd/api/main.go                 # 组合根：配置 → 依赖装配 → 启动/优雅关闭
├── internal/
│   ├── domain/                     # 纯业务层：实体、不变量、领域错误、Repository 端口
│   ├── usecase/                    # 应用层：编排用例，只依赖 domain
│   ├── adapter/
│   │   ├── http/                   # 路由、中间件、DTO、校验、错误映射
│   │   └── postgres/               # Repository 实现 + sqlc 生成代码
│   ├── config/                     # 环境变量 → Config
│   └── platform/                   # logger、http server 等基础设施
├── db/
│   ├── migrations/                 # goose 迁移（打包进二进制）
│   └── queries/                    # sqlc 输入 SQL
├── api/requests.http               # 可直接执行的 API 请求
├── configs/.env.example            # 本地环境变量模板
├── scripts/                        # dev / migrate 脚本
├── sqlc.yaml
├── Makefile
└── docker-compose.yml              # 本地 PostgreSQL（5433）
```

### 依赖方向（由 depguard 约束）

```text
cmd            → 所有层（仅装配）
adapter/http   → usecase, domain
adapter/postgres → domain
usecase        → domain
domain         → 仅标准库
```

## 快速开始

### 1. 前置条件

- Go 1.27+（`brew install go`）
- PostgreSQL 16（本机 5432，或使用 `make docker-up` 启动 5433）

### 2. 创建数据库

推荐使用项目自带的 compose（账号/密码均为 `postgres`，自动创建 `learning_notebook`）：

```bash
make docker-up
```

如果改用本机已有的 PostgreSQL，请自行创建数据库并调整 `DATABASE_URL` 端口：

```bash
psql -U postgres -c "CREATE DATABASE learning_notebook;"
```

### 3. 配置环境变量

```bash
cp configs/.env.example configs/.env
```

默认连接串：

```text
postgres://postgres:postgres@localhost:5433/learning_notebook?sslmode=disable
```

项目自带的 compose 服务映射到 **5433**，避免与本机已有的 5432 实例冲突。如果你使用已有的 5432 实例，请先在那边创建 `learning_notebook` 数据库，再把端口改回 5432。

### 4. 启动

```bash
make run
```

服务默认监听 `:8080`。首次启动会自动执行 goose 迁移（`DATABASE_AUTO_MIGRATE=true`）。

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}

curl http://localhost:8080/readyz
# {"status":"ok"}（会 ping PostgreSQL）
```

## API

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/notebooks` | 创建笔记 |
| GET | `/api/v1/notebooks?limit=20&offset=0` | 分页列表 |
| GET | `/api/v1/notebooks/{id}` | 查询单条 |
| PATCH | `/api/v1/notebooks/{id}` | 部分更新（title / content 至少一个） |
| DELETE | `/api/v1/notebooks/{id}` | 删除 |
| GET | `/healthz` | 存活检查 |
| GET | `/readyz` | 就绪检查（含数据库） |

### 创建

```bash
curl -i -X POST http://localhost:8080/api/v1/notebooks \
  -H 'Content-Type: application/json' \
  -d '{"title":"第一篇笔记","content":"Hello, Go Web!"}'
```

```json
{
  "data": {
    "id": "3f0f2b7e-...",
    "title": "第一篇笔记",
    "content": "Hello, Go Web!",
    "created_at": "2026-10-18T06:30:00Z",
    "updated_at": "2026-10-18T06:30:00Z"
  }
}
```

### 列表

```bash
curl 'http://localhost:8080/api/v1/notebooks?limit=20&offset=0'
```

```json
{
  "data": [ ... ],
  "meta": { "limit": 20, "offset": 0, "total": 1 }
}
```

### 部分更新

```bash
curl -X PATCH http://localhost:8080/api/v1/notebooks/<id> \
  -H 'Content-Type: application/json' \
  -d '{"title":"更新后的标题"}'
```

### 删除

```bash
curl -i -X DELETE http://localhost:8080/api/v1/notebooks/<id>
# HTTP/1.1 204 No Content
```

### 错误格式

```json
{
  "error": {
    "code": "validation_failed",
    "message": "request validation failed",
    "details": { "title": "is required" },
    "request_id": "9a8b7c6d5e4f3a2b"
  }
}
```

| 领域错误 | HTTP 状态码 |
|---|---|
| `ErrValidation` | 422 |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| 其他 | 500 |

每个响应都会带 `X-Request-ID`，并出现在访问日志与错误体中。

## 常用命令

```bash
make help              # 列出所有命令
make run               # 启动服务
make build             # 编译到 bin/api
make test              # 单元测试
make test-integration  # PostgreSQL 集成测试（需 TEST_DATABASE_URL）
make lint              # golangci-lint（含 depguard 分层规则）
make fmt               # gofmt
make sqlc              # 修改 db/queries 后重新生成代码
make migrate-up        # goose 手动执行迁移
make migrate-status    # 查看迁移状态
make docker-up         # 启动本地 PostgreSQL（5433）
```

### 集成测试

```bash
docker exec postgresql psql -U postgres -c "CREATE DATABASE learning_notebook_test;"
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/learning_notebook_test?sslmode=disable' \
  make test-integration
```

未设置 `TEST_DATABASE_URL` 时，该测试会自动跳过。

## 开发说明

### 修改表结构

1. 新增迁移文件：`make migrate-create NAME=add_tags`（需要 goose CLI）；
2. 编辑 `db/migrations/*.sql`，必须包含 `-- +goose Up` 与 `-- +goose Down`；
3. 更新 `db/queries/*.sql`；
4. `make sqlc` 重新生成；
5. `make migrate-up`（或启动服务时自动迁移）。

### 分层约束

- `domain` 只能 import 标准库；
- `usecase` 只能 import `domain` 与标准库；
- `adapter` 之间禁止互相依赖；
- HTTP handler 不写 SQL / 业务规则，usecase 不感知 `http.Request` 与 SQL；
- 以上规则写在 `.golangci.yml` 的 `depguard` 中，由 `make lint` 机器化约束。

### 热重载（可选）

```bash
brew install air
air
```

## 生产提示

- 使用真实环境变量，不依赖 `configs/.env`；
- 生产建议 `DATABASE_AUTO_MIGRATE=false`，由发布流水线执行迁移；
- `APP_ENV=production` 时日志切换为 JSON；
- 使用 `Dockerfile` 多阶段构建镜像：

```bash
docker build -t learning-go-web-notebook .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL='postgres://postgres:postgres@host.docker.internal:5432/learning_notebook?sslmode=disable' \
  learning-go-web-notebook
```
