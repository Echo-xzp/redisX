# 项目 TODO 列表（Project TODOs）

此文件汇总项目当前的已完成、进行中与待办任务，便于跟踪 Phase-1 的推进与后续工作计划。

---

## ✅ 已完成
- Project scaffold（项目结构初始化）
- RESP parser（支持 inline / RESP 数组 / bulk strings，含单元测试）
- Storage layer（内存 map + 锁，支持 Expire 与 janitor）
- Core commands（PING, SET(EX/PX), GET, DEL, EXISTS, EXPIRE, TTL, INFO, QUIT）
- Unit & integration tests（storage、protocol、server 关键场景测试）
- README & docs（包含使用示例与测试说明）
- Release artifacts & tag（生成 release 二进制并创建本地 tag v0.1.0）

## 🛠 开发任务（按优先级）
1. **已完成**：增强 RESP 解析器 — 支持 pipelining、`null` bulk string、严格的长度与错误处理（已通过单元测试并在本地 `go test ./...` 通过）。
2. 新增命令（高）：`INCR` / `DECR` / `INCRBY` / `DECRBY` / `MGET` / `MSET` / `PERSIST`。
3. 过期精度（高）：实现 `PTTL` / `PEXPIRE` 支持毫秒级过期并改进过期调度策略。
4. 连接与资源限制（高）：连接超时、最大连接数、内存上限与优雅拒绝策略。
5. 服务配置（中）：支持通过环境变量或命令行参数配置端口、`max-memory`、`log-level` 等。
6. 内存驱逐（中）：实现 LRU 或 LFU 策略以在内存受限时驱逐 key。
7. 基准与压力测试（中）：实现 `bench` 和压测脚本以定位性能瓶颈。
8. 日志与指标（中）：扩展 `INFO`，并考虑导出 Prometheus 风格指标以便监控。
9. 持久化（后续）：实现 RDB/AOF 或可插拔持久化后端以保证数据在重启后恢复。
10. 文档（持续）：补充命令参考、贡献指南、PR 模板（参见 `docs/CONTRIBUTING.md`）。

> 注：已将“增强 RESP 解析器”标记为 **进行中**，我可以立即开始并提交小的增量改动以便频繁验证。

---

## 📝 细化实现任务（可逐步执行）

下面将高优先级任务拆成可执行的子任务，包含目标、改动位置与测试建议，便于逐条实现并合并 PR。

- **增强 RESP 解析器（已完成）** ✅
  - 目标：支持 pipelining（流式多命令解析）、`$-1`（null bulk string）和严格长度/错误检测。
  - 改动位置：`internal/protocol/resp.go`（修改以支持 `$-1` 与严格长度校验）；测试：`internal/protocol/resp_test.go`（新增 `TestParseNullBulk`、`TestPipelining`、`TestStrictLengthMismatch`、`TestInvalidBulkHeader`）。
  - 交付标志：已新增单元测试并修复解析器实现，`go test ./...` 本地通过；建议后续新增 server 层的流水线集成测试以覆盖端到端行为。

- **命令路由与命令实现（高）**
  - 目标：引入 `internal/command`，实现路由与 handler 解耦，先实现 `INCR/DECR/INCRBY/DECRBY/MGET/MSET/PERSIST`。
  - 改动位置：新增 `internal/command/router.go`、`internal/command/handlers_*.go`，并在 `internal/server` 中调用路由。
  - 测试：为每个命令编写单元测试，并在 `internal/server/server_integration_test.go` 中增加端到端验证。

- **毫秒级过期（高）**
  - 目标：将过期精度扩展到毫秒，新增 `PTTL` / `PEXPIRE` 命令，并保证 `SET PX` 支持毫秒。
  - 改动位置：`internal/storage/storage.go`（Entry、Set、IsExpired、janitor 调度）、`internal/server` 的命令处理。
  - 测试：`internal/storage/storage_test.go` 与 server 集成测试覆盖 ms 精度场景与并发删除。

- **连接与资源限制（高）**
  - 目标：支持 `max-conns`、连接超时、`max-memory` 配置与优雅拒绝策略。
  - 改动位置：`cmd/redisx/main.go`（flag/env）、`internal/server/server.go`（连接管理）。
  - 测试：模拟超出上限的连接行为、超时连接回收测试。

- **内存驱逐（中）**
  - 目标：实现 LRU（或 LFU）基础策略，在 `max-memory` 场景下驱逐 key 并更新 `INFO` 数据。
  - 改动位置：`internal/storage`（数据结构扩展）、`internal/server/info` 输出扩展。
  - 测试：驱逐触发与命中率基础断言，基准测试评估开销。

- **基准与压测（中）**
  - 目标：添加 `bench` 用例和压测脚本（`test/bench`），定位热点与锁竞争。
  - 改动位置：新增 `test/bench` 和 CI 压测脚本（可选），记录基线。

- **持久化与长期扩展（低）**
  - 目标：设计可插拔持久化接口（RDB/AOF），作为后续迭代。


## 🎯 短期可交付（1-2 周）
1. 完整实现 RESP 增强（pipelining、null bulk、严格错误处理）并补充测试（高优先）。
2. 引入 `internal/command` 路由并实现至少三个命令（推荐：INCR、MGET、PERSIST），通过单元与集成测试（高优先）。
3. 支持毫秒过期（PEXPIRE/PTTL）并更新 `SET PX` 行为（高优先）。

---

## ✅ PR 与测试策略
- 小步提交：每个功能点单独 PR，PR 描述包含改动文件、测试项与回滚计划。
- 测试优先：先编写失败的测试（TDD），然后实现最小改动以通过测试。
- CI 要求：所有更改必须带相应单元测试与集成测试，`go test ./...` 通过为合并前提。

---

更新记录：
- 2026-01-10: 初始 TODO 列表（由开发迭代与发布过程生成）。
- 2026-01-10: 增强 RESP 解析器：新增 pipelining/null bulk/严格长度测试并修复解析器实现（修改 `internal/protocol/resp.go`，新增 `internal/protocol/resp_test.go` 测试用例），本地测试 `go test ./...` 全部通过。
- 2026-01-10: 支持毫秒级过期（PEXPIRE/PTTL）与 `SET PX` 毫秒精度：修改 `internal/storage` 与 `internal/server`，新增 `SetWithMs` / `PExpire` / `PTTL` 实现与单元/集成测试，`go test ./...` 全部通过。

欢迎就优先级或细化任务列表给出意见。