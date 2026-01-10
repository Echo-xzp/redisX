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

## ⚙️ 进行中
- Concurrency & pressure tests（并发压力与基准测试）

## 📝 待实现（优先级排序建议）
1. CI（GitHub Actions）与自动构建/测试（高）
2. Protocol edge tests（RESP 边界、错误场景、流水线/多命令）
3. Connection / resource limits（连接超时、最大连接数、内存上限等）
4. Persistence backend（RDB/AOF 或可插拔后端，后续阶段）
5. Memory eviction strategy（LRU/LFU/TTL 优化）
6. Performance benchmark & tuning（基准测试与热点优化）
7. Security（TLS、认证、ACL，后续）
8. Remote release & GitHub Release（推送二进制并在 GitHub 上创建 Release）
9. Contributor guide（贡献流程、PR 模板、CODEOWNERS）

## 📌 下步建议
- 立刻优先：编写并运行并发压力测试（验证线程安全与性能瓶颈）。
- 并行进行：配置 CI（自动运行 tests 与 cross-build），以保证 PR 的基础质量。 

---

更新记录：
- 2026-01-10: 初始 TODO 列表（由开发迭代与发布过程生成）。

欢迎就优先级或细化任务列表给出意见。