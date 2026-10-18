---
name: go-web-project-rules
description: "新建 Go Web 项目骨架、在既有 Go Web 项目中新增或修改 HTTP 接口、调整目录与分层、编写 domain 与 usecase 与 adapter 代码、装配配置与启动流程、接入 chi 与 pgx 与 sqlc 与 goose 时触发。"
disable-model-invocation: false
---

# Go Web 项目编码规则

以下规则为强制约束，逐条执行。

## 目录与布局

1. 可执行程序入口只放在 `cmd/<二进制名>/main.go`；本项目固定为 `cmd/api/main.go`。

2. 业务代码全部放在 `internal/` 下；禁止在 `cmd/` 中写业务规则、SQL 或跨层编排。

3. 最内层业务代码放在 `internal/domain/`；应用编排放在 `internal/usecase/`；外部交互放在 `internal/adapter/`。

4. HTTP 交互只放在 `internal/adapter/http/`；PostgreSQL 交互只放在 `internal/adapter/postgres/`；一个外部依赖一个 adapter 包。

5. 基础设施代码放在 `internal/platform/`；配置加载放在 `internal/config/`；两者都不得包含业务规则。

6. 数据库迁移放在 `db/migrations/`，sqlc 输入查询放在 `db/queries/`；禁止把 SQL 散落到 Go 源文件中。

7. API 请求样例放在 `api/`，本地环境变量模板放在 `configs/.env.example`，构建与迁移脚本放在 `scripts/`。

8. 禁止创建 `pkg/`、`src/`、`models/`、`utils/` 目录；确需对外复用代码时先向用户确认目标位置。

## 依赖方向

9. 依赖方向只允许 `cmd → adapter / usecase / platform / config`、`adapter → usecase / domain`、`usecase → domain`、`domain → 标准库`。

10. `internal/domain/` 只导入标准库；禁止导入 `chi`、`pgx`、`slog` 或本项目其他包。

11. `internal/usecase/` 只导入 `internal/domain/` 与标准库；禁止导入 `net/http`、`database/sql`、`pgx` 或 adapter 包。

12. HTTP adapter 与 PostgreSQL adapter 之间禁止互相导入；跨 adapter 复用逻辑下沉到 `usecase` 或 `domain`。

13. 分层规则以 `.golangci.yml` 的 `depguard` 为准；新增跨层导入前必须同步更新该配置。

## domain 层

14. 实体、值对象、领域不变量、领域错误与 Repository 端口只放在 `internal/domain/`。

15. 实体字段只使用标准库类型；禁止在实体上出现 `json`、`db`、`validate` 等传输或持久化 tag。

16. 不变量在构造函数与实体方法中校验；禁止把领域校验只放在 HTTP DTO 或数据库约束中。

17. 领域错误用哨兵错误 `ErrXxx` 定义，调用方用 `errors.Is` 判定；禁止在 domain 中构造 HTTP 状态码。

18. Repository 端口定义在 domain，由 adapter 实现；禁止在 domain 中导入具体驱动或生成代码。

## usecase 层

19. 业务规则、流程编排与状态判定只写在 `internal/usecase/`。

20. usecase 的输入用本层结构体承载；HTTP DTO 在 adapter 中转换为 usecase 输入，禁止把 DTO 传入 usecase。

21. usecase 只依赖 `domain` 端口与标准库；禁止出现 `http.Request`、`http.ResponseWriter`、SQL 或 pgx 类型。

22. 事务边界与跨 Repository 的编排放在 usecase；单个 Repository 方法内能完成的操作不额外开事务。

23. usecase 返回领域对象或领域错误；禁止返回 HTTP 状态码、JSON 结构或 adapter DTO。

## HTTP adapter

24. 路由注册、中间件挂载与 URL 版本前缀只写在 `internal/adapter/http/router.go`。

25. 中间件顺序固定为 `RequestID → AccessLog → Recoverer → Timeout`；调整顺序前必须确认日志与恢复语义。

26. 一个业务域一个 handler 文件；handler 结构体通过构造函数注入 usecase、validator 与 logger。

27. handler 只做解码、传输层校验、调用 usecase、映射 DTO 与错误；禁止在 handler 中写业务条件分支、循环或 SQL，这类逻辑下沉到 usecase。

    > 示例：
    >
    > ```go
    > notebook, err := h.service.Create(r.Context(), usecase.CreateNotebookInput{Title: req.Title})
    > if err != nil { writeError(w, r, h.log, err); return }
    > writeJSON(w, http.StatusCreated, dataResponse{Data: toNotebookResponse(notebook)})
    > ```

28. 请求体解码用 `json.Decoder`，开启 `DisallowUnknownFields()` 并用 `http.MaxBytesReader` 限制大小。

29. 请求 DTO 与响应 DTO 分别定义，字段按用途命名；禁止把领域实体直接作为 JSON 响应返回。

30. 传输层校验只用 `go-playground/validator`，校验 tag 写在 DTO 上；领域不变量仍由 domain 校验。

31. 校验错误按字段返回 `details`，字段名用 JSON 名称；禁止把 `validator.ValidationErrors` 原样返回给客户端。

32. 路径参数用 `chi.URLParam` 读取，查询参数显式解析并返回错误；禁止把解析失败当成零值继续执行。

33. 领域错误到 HTTP 状态的映射集中在 `internal/adapter/http/errors.go`；handler 禁止逐个 `switch` 领域错误。

34. 响应统一用 `data` 包装，列表响应附加 `meta`；错误统一用 `error` 对象，包含 `code`、`message`、`details` 与 `request_id`。

    > 示例：
    >
    > ```json
    > {"data": {...}}
    > {"data": [...], "meta": {"limit": 20, "offset": 0, "total": 1}}
    > {"error": {"code": "not_found", "message": "notebook not found", "request_id": "..."}}
    > ```

35. 成功状态码按语义选择：创建用 `201` 并设置 `Location`，删除用 `204`，其余用 `200`；禁止在 handler 中手写非标准状态码。

## PostgreSQL adapter 与 sqlc

36. 连接池只在 `internal/adapter/postgres/db.go` 中创建与关闭；业务代码禁止自行 `pgx.Connect`。

37. 连接池创建后必须 `Ping` 验证可达性；失败时返回错误并终止启动。

38. Repository 实现放在 `internal/adapter/postgres/`，每个实现用 `var _ domain.XxxRepository = (*XxxRepository)(nil)` 做编译期断言。

39. sqlc 生成代码只放在 `internal/adapter/postgres/sqlc/`，只允许 `internal/adapter/postgres/` 导入；禁止在其他包直接使用。

40. 生成代码禁止手工修改；改表或改查询后运行 `make sqlc` 重新生成。

41. SQL 只写在 `db/queries/*.sql`，每条查询用 `-- name: Xxx :one`、`:many` 或 `:execrows` 声明名称与返回形态。

42. 查询名称用 `MixedCaps` 动词短语；禁止在 Go 中拼接 SQL 片段或在查询文件名中表达业务版本。

43. PostgreSQL 驱动错误必须在 adapter 边界转换为领域哨兵错误：`pgx.ErrNoRows` 转 `domain.ErrNotFound`，唯一冲突转 `domain.ErrConflict`，非法输入转 `domain.ErrValidation`。

    > 示例：
    >
    > ```go
    > if errors.Is(err, pgx.ErrNoRows) {
    >     return nil, fmt.Errorf("get notebook: %w", domain.ErrNotFound)
    > }
    > ```

44. adapter 返回领域对象或标量；禁止返回 HTTP 概念、DTO 或把业务规则写进 Repository。

## 配置与启动装配

45. 配置只从 `internal/config/config.go` 的 `config.Load()` 读取；业务代码禁止调用 `os.Getenv` 或读取 `configs/.env`。

46. 配置字段用 `caarlos0/env` 的 `env` 与 `envDefault` tag 声明类型和默认值；新增配置项同步更新 `configs/.env.example`。

47. `cmd/api/main.go` 的 `run()` 是组合根：加载配置、构造 logger、执行迁移、创建连接池、装配 Repository、usecase、handler、router 与 server；禁止把业务逻辑写进组合根。

48. 依赖通过构造函数显式注入；禁止使用全局单例、`init()` 装配或服务定位器。

49. HTTP server 必须配置 `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout` 与 `IdleTimeout`；禁止使用无超时的裸 `http.ListenAndServe`，改用 `internal/platform/server` 的 `Server`。

50. 进程收到 `SIGINT` 或 `SIGTERM` 时必须调用 `http.Server.Shutdown` 并设置关闭超时；启动失败必须返回错误并由 `main()` 记录后退出。

## 迁移与查询生成

51. 迁移文件只放在 `db/migrations/`，命名遵循 goose 的版本前缀；必须包含 `-- +goose Up` 与 `-- +goose Down`。

52. 已应用或已合并的迁移文件禁止修改；结构变更必须新增迁移文件。

53. `Down` 必须能撤销对应的 `Up`；禁止写空的 `Down`，用户明确要求不可逆迁移时除外。

54. 迁移通过 `db/migrations/migrate.go` 的 `Run` 执行并嵌入二进制；本地也可用 `make migrate-up`。

55. 修改表结构后必须同步更新 `db/queries/*.sql` 并运行 `make sqlc`；禁止只改迁移不改查询。

## 日志、健康检查与请求上下文

56. 日志统一用 `log/slog`，在组合根创建后逐层注入；禁止在包级变量中保存 logger。

57. 请求 ID 由中间件生成或沿用 `X-Request-ID`，通过 `context` 传递；日志与错误响应必须带同一个请求 ID。

58. 访问日志由中间件统一记录 `method`、`path`、`status`、`bytes`、`duration_ms` 与 `request_id`；禁止在 handler 中逐条记录请求信息。

59. 服务必须提供 `GET /healthz` 与 `GET /readyz`；`readyz` 必须检查数据库可达性，检查失败返回 `503`。
