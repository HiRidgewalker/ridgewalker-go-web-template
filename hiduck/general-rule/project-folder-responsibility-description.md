# 项目目录职责规则

以下规则为强制约束，逐条执行。本文件只规定目录分类、目录职责与文件归属；Go 分层编码细节遵循 [go-web-project-rules](../../.agents/skills/go-web-project-rules/SKILL.md)。

## 根目录与顶层分类

FOLDER-1. 顶层目录按三类归位：Go 工具链识别目录为 `cmd/`、`internal/`、`pkg/`、`vendor/`；工程配置与资源目录为 `api/`、`web/`、`configs/`、`init/`、`scripts/`、`build/`、`deployments/`、`test/`、`docs/`、`tools/`、`examples/`、`third_party/`、`githooks/`、`assets/`、`website/`；项目扩展目录为 `db/`、`bin/`、`hiduck/`、`.agents/`、`.claude/`、`.vscode/`。

FOLDER-2. 只创建当前功能实际使用的目录；禁止预建空目录、空分层文件和为将来预留的目录。

FOLDER-3. 根目录只放模块文件、工程配置与项目级说明文档，包括 `go.mod`、`go.sum`、`Makefile`、`README.md`、`.gitignore`、`.golangci.yml`、`sqlc.yaml`、`.air.toml`、`Dockerfile`、`docker-compose.yml`；禁止在根目录堆放业务源码、SQL 或散落脚本。

FOLDER-4. `go.mod` 只放在根目录且只声明一个模块；禁止在子目录创建第二个 `go.mod`，多模块拆分必须先经用户确认。

FOLDER-5. `Makefile` 只保留简短入口目标，多步流程放 `scripts/`；禁止在配方中内联大段脚本。

FOLDER-6. 文件归属按实际承担的能力判定，不因目录名含有 `util`、`common`、`helper` 等词就放宽边界。

## Go 工具链目录

FOLDER-7. `cmd/<二进制名>/` 只放 `main.go` 与启动装配代码，子目录名与产出的二进制名一致；本项目固定为 `cmd/api/main.go`。

FOLDER-8. `cmd/` 中禁止放业务规则、SQL、handler 实现与跨层编排；这些内容下沉到 `internal/` 的对应包。

FOLDER-9. `internal/` 存放不允许外部模块导入的应用与库代码；本项目全部业务代码只放在 `internal/` 下。

FOLDER-10. `internal/` 按 `domain/`、`usecase/`、`adapter/<外部依赖>/`、`platform/<基础设施>/`、`config/` 划分；新增分层前先确认现有层无法承载该职责。

FOLDER-11. `internal/adapter/` 一个外部依赖一个包；禁止创建 `adapter/common/`、`adapter/base/` 这类跨依赖共享包，可复用逻辑下沉到 `usecase/` 或 `domain/`。

FOLDER-12. 本项目禁止创建 `pkg/`；确需向外部模块提供可导入代码时，先向用户说明目标位置与 API 范围再创建。

FOLDER-13. `vendor/` 只由 `go mod vendor` 生成，禁止手工编辑；未启用 vendoring 时禁止创建该目录。

## 工程配置与资源目录

FOLDER-14. `api/` 存放对外协议定义与可直接执行的请求样例，如 OpenAPI 规范、JSON Schema、proto 文件与 `requests.http`；禁止放 handler 代码或业务逻辑。

FOLDER-15. `configs/` 存放配置模板与默认值，如 `.env.example`；真实密钥与本地环境文件 `configs/.env` 禁止提交。

FOLDER-16. 配置的读取与校验代码只放在 `internal/config/`；`configs/` 只存数据文件，禁止在目录内放可执行脚本。

FOLDER-17. `scripts/` 存放构建、安装、迁移、分析等可独立执行的脚本；一个脚本只做一件事，并按用途命名。

FOLDER-18. 可编译的 Go 支持工具放 `tools/`，Shell 与命令行流程放 `scripts/`；禁止把 Go 源码放进 `scripts/`。

FOLDER-19. `test/` 存放跨包共用的测试数据与独立外部测试程序，测试数据放 `test/testdata/`；`*_test.go` 必须与被测代码同目录，禁止集中到 `test/`。

FOLDER-20. `build/package/` 存放多形态打包配置，`build/ci/` 存放持续集成配置；只有出现对应产物或流水线时才创建。

FOLDER-21. `deployments/` 存放容器编排、Kubernetes、Helm 与 Terraform 等部署配置；本项目当前使用根目录 `docker-compose.yml`，出现第二种部署形态时再创建。

FOLDER-22. `web/` 只存放静态资源、服务端模板与单页应用，纯 REST API 项目禁止创建；`init/` 只存放 systemd、supervisord 等进程管理配置，容器化部署禁止创建。

FOLDER-23. `docs/` 存放对外设计与使用文档；本项目文档统一放 `hiduck/doc/`，禁止同时维护 `docs/`。

FOLDER-24. `assets/` 存放仓库级图片与徽标；本项目资源统一放 `hiduck/assets/`，禁止同时维护顶层 `assets/`。

FOLDER-25. `examples/`、`third_party/`、`githooks/`、`website/` 只在出现可运行示例、第三方 fork、Git 钩子或项目站点时创建；禁止为对齐目录清单而预建。

## 项目扩展目录

FOLDER-26. `db/migrations/` 只放 goose 迁移文件与迁移执行代码，`db/queries/` 只放 sqlc 输入 SQL；禁止把 SQL 写进 `internal/` 的 Go 文件。

FOLDER-27. `bin/` 只存放构建产物并由 `.gitignore` 排除；禁止在 `bin/` 中放源码、配置或脚本。

FOLDER-28. `hiduck/` 存放协作文档：`general-rule/` 通用规则、`prd/` 需求、`plan/` 计划、`doc/` 使用文档、`assets/` 图片资源；禁止放入运行时代码。

FOLDER-29. `.agents/skills/` 存放技能源文件与 `agents/openai.yaml`；`.claude/skills/` 只存放指向 `../../.agents/skills/<name>` 的相对软链，禁止复制技能正文。

FOLDER-30. `.vscode/` 只存放编辑器与预览配置；禁止放构建配置、运行参数或密钥。

## 文件命名与目录创建

FOLDER-31. Go 源文件用全小写、无下划线的单词命名，按业务域一文件一职责，如 `notebook.go`、`router.go`；测试文件用 `xxx_test.go` 与被测文件同目录。

FOLDER-32. SQL 迁移文件用 goose 版本前缀加下划线描述，如 `00001_create_notebooks.sql`；查询文件按表或业务域命名为复数，如 `notebooks.sql`。

FOLDER-33. Shell 脚本用 `kebab-case.sh` 命名并声明 shebang；文档用 `YYYY-MM-DD-<短名>.md` 或语义化名称，禁止使用空格。

FOLDER-34. 只有单一职责文件过多或需要独立子域时才建层内子目录；子目录必须继承父目录职责，禁止混放其他层文件。

FOLDER-35. 新增目录前写明其职责与归属判据；无法判定时向用户说明用途、依赖与候选位置后再确认。

## 禁止目录

FOLDER-36. 禁止创建 `src/`；源码直接从 `cmd/` 与 `internal/` 开始，不使用 Java 式源码根目录。

FOLDER-37. 禁止创建 `models/`、`utils/`、`common/`、`helper/`、`base/` 等按抽象概念命名的一级目录；对应能力归入 `internal/domain/`、`internal/platform/` 或具体 adapter 包。

FOLDER-38. 禁止同一职责并行维护两处目录；迁移到新位置后必须删除旧目录并同步全部引用。
