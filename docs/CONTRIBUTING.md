# 贡献指南（CONTRIBUTING.md）

> 目的：为人工开发者与自动化 agent（AI coding agents）提供统一、可检索的规范与启动信息，确保任务推进、代码变更与文档同步的一致性与可追溯性。

---

## 1. 快速启动（必做） 🔧
- 环境：需要 Go >= 1.21。查看 `go.mod`。建议本地使用 `go 1.21.x`。
- 常用命令：
  - 运行全部测试：`go test ./...`
  - 构建：`go build ./...`
  - 格式化：`gofmt -w .` 或 `go fmt ./...`
  - 静态检查（若安装 golangci-lint）：`golangci-lint run`。

## 2. 项目关键位置（Agent 应优先读取） 📂
- `docs/TODO.md`：任务清单（优先级、改动位置、测试要点、交付标准）。
- `docs/work/工作记录.md`：按时间序追加的变更日志（每次完成任务后必须追加）。
- `docs/CONTRIBUTING.md`：本文件（包含 agent 行为规范、模板与示例）。
- 代码关键文件：
  - `internal/protocol/resp.go` / `internal/protocol/resp_test.go`：RESP 协议解析器
  - `internal/storage/storage.go` / `internal/storage/storage_test.go`：数据存储与过期
  - `internal/server/server.go` / `internal/server/server_integration_test.go`：TCP 服务与命令路由
  - `cmd/redisx/main.go`：程序入口与配置

## 3. 任务与变更工作流（必遵） ✅
1. 在 `docs/TODO.md` 中添加或确认任务（包含优先级、预期改动文件、测试要求）。
2. 优先采用 TDD：先新增失败的单元/集成测试（放在对应 `_test.go`），然后实现最小改动通过测试。
3. 每次实现完成后：
   - 将对应任务在 `docs/TODO.md` 标注为 **已完成** 并记录改动文件与测试结果。
   - 在 `docs/work/工作记录.md` 中追加一条日志（采用模板，见下）。
4. 提交/PR：遵守分支与提交约定（见下），PR 描述需包含改动概述、测试说明与回滚提示。

## 4. 分支与提交约定 🧭
- 分支命名：`feat/<short>`、`fix/<short>`、`docs/<short>`、`chore/<short>`。
- 提交信息：以动词开头、简洁明了，例如：`resp: support null bulk ($-1)`。
- PR 标题：`<scope>: <简短描述>`（例如 `protocol: add pipelining support`）。

## 5. PR 要求与验证流程 ✅
- 所有 PR 必须包含：改动描述、受影响文件、测试说明、验证步骤与回滚建议。
- CI（`go test ./...`）必须通过；至少一个项目维护者审阅通过才能合并。
- 若必须临时跳过测试或采用例外流程，需在 PR 描述中记录原因并在 `docs/work/工作记录.md` 中写明补救计划。

## 6. Agent（AI）专属规范与模板 🤖
- 启动检查（Agent 启动时必须执行）：
  1. 读取：`docs/CONTRIBUTING.md`, `docs/TODO.md`, `docs/work/工作记录.md`, `dosc/编码规范.md`, `第一阶段实施计划.md`, `技术原理分析.md`。
  2. 执行：`git status --porcelain`（保证工作区状态可控或记录原因）。
  3. 执行：`go test ./...`（确保测试通过）。
- 行为准则：
  - 每次修改前：在 `docs/TODO.md` 添加/更新任务小节（若为新任务）。
  - 每次修改后：将任务标为已完成并在 `docs/work/工作记录.md` 追加日志。
  - 每次工具调用或修改前应输出简短 preamble（说明当前发现或计划的下一步，最多 2 句）。
  - 遵循小步实现原则：一次 PR 只实现一件具体功能并附带测试。
- Agent 文本模板：
  - Preamble 示例："Perfect! I found that X is failing due to Y; next step is to add a test for Z and implement fix."（保持简短明确）
  - TODO 更新模板：`- **已完成**：<任务名称> — 修改 <files>；测试：<tests>；验证：<commands>`
  - 工作记录模板：
    ```markdown
    ## 更新 - <短标题>（日期：YYYY-MM-DD）
    - 变更文件：<file1>, <file2>
    - 测试：<新增/修改的测试描述，`go test ./...` 是否通过>
    - 验证：<如何手动或自动验证>
    - 备注：<建议或后续动作>
    ```

## 7. 代码风格与质量标准 🔍
- 使用 `go fmt` 进行格式化，建议在本地或 pre-commit 中运行。
- 代码需有明确错误处理；避免吞掉错误。遵循项目现有的风格（见 `docs/编码规范.md`）。
- 所有新增逻辑应有单元测试，重要的端到端行为应有集成测试。

## 8. 常见任务示例（Agent 可参考的执行步骤）
- 例：实现 `PEXPIRE` / `PTTL`
  1. 在 `internal/storage/storage_test.go` 添加毫秒级过期测试（失败）。
  2. 在 `internal/protocol` 或 `internal/server` 添加命令处理测试（失败）。
  3. 在 `internal/storage/storage.go` 增加 ms TTL 支持并实现清理逻辑；修改 `server` 中的 `SET PX` 逻辑以支持 ms。
  4. 运行 `go test ./...`，修复直到通过，更新 `docs/TODO.md` 与 `docs/work/工作记录.md`，提交 PR。

## 9. 联系人与审阅者
- 当前维护者：请在 PR / issue 中 @ 项目负责人或相关 reviewer（在 README 或团队沟通工具中指明）。

---

感谢遵守本项目贡献指南；如需扩展或细化 agent 行为规范可在此文件中迭代更新。