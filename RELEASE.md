# redisX Release v0.1.0

发布日期：2026-01-10

简介

这是 redisX 的首个发布（v0.1.0），包含基础的 RESP 协议支持、SET/GET/DEL/EXISTS、键过期（EX/PX）、INFO 命令与测试覆盖。

重要改动/高亮

- 基本 RESP 协议解析（行协议与 RESP 数组/批量字符串）。
- 命令实现：SET（支持 EX/PX）、GET、DEL、EXPIRE、TTL、EXISTS、INFO、PING、QUIT。
- 存储层：基于内存的线程安全 map，支持懒删除与后台 janitor 清理。
- 单元测试与集成测试覆盖关键行为（过期、协议解析、服务器行为）。

发布产物（位于 `release/` 目录）

- redisx-darwin-amd64
  - SHA256: C2D80685970FC4E86227D9318C91CAAA0D796E354A760284CB8722A8C2A11B9D
- redisx-linux-amd64
  - SHA256: 857E04670DFF5B18A61906510EE01CE42CD51B9EEA3CCAB55DC030EB64AFEAA4
- redisx-linux-arm64
  - SHA256: BC2AD028317242FE6699E5199B923B64649A5C879F15A2BA3A98D12F7B97F044
- redisx-windows-amd64.exe
  - SHA256: 743D2A0DA88A4034833B94D8D9BC61F8181B494C7FB3CA7BE6836BA4131F400B

如何验证

- 下载相应二进制并验证校验和：

```powershell
Get-FileHash .\redisx-linux-amd64 -Algorithm SHA256
# 比对输出的 Hash 与上表
```

运行（快速示例）

```bash
# 在对应平台上：
./redisx-darwin-amd64      # 或 redisx-linux-amd64
# Windows:
./redisx-windows-amd64.exe
```

使用 `redis-cli` 测试示例：

```bash
redis-cli -p 6379 PING
redis-cli -p 6379 SET mykey "hello"
redis-cli -p 6379 GET mykey
redis-cli -p 6379 SET temp PX 500
redis-cli -p 6379 INFO
```

后续计划

- 添加并行/压力测试以验证高并发场景下的稳定性与性能。
- 增加更多协议兼容性测试与更完整的命令支持。

感谢使用与反馈！欢迎通过 issue/PR 贡献改进或报告 bug。

---

# redisX Release v0.2.0

发布日期：2026-01-11

简介

这是 redisX 的第二版发布（v0.2.0），包含对过期精度、命令路由、资源限制等重要功能的增强与 bug 修复。

重要改动/高亮

- 毫秒级过期支持：新增 PEXPIRE / PTTL，并支持 `SET ... PX <ms>`（毫秒级 TTL）。
- 命令路由与扩展：新增命令路由模块以及 INCR、MGET、PERSIST 的实现与测试。
- 资源限制：在服务器层新增 MaxConns、ConnTimeout 与 MaxMemoryBytes，支持在内存限制下拒绝写入（`TrySet` 系列）。
- RESP 与协议修复：增强 RESP 解析以支持 pipelining、null bulk 等边界情况，改进解析健壮性与测试覆盖。
- 存储增强：统一使用 Unix 毫秒时间戳保存过期时间，增加内存使用计数、TrySet/TrySetWithMs、IncrBy 与后台 janitor 的一致性修复。
- 测试：新增/修复大量单元测试与集成测试，当前本地测试通过。

发布产物（本地构建并校验）

- redisx-darwin-amd64
  - SHA256: A449A5ED57949F24D2B6EFA0D1836564B4C3211478824F23AA1B1707E771D2F4
- redisx-linux-amd64
  - SHA256: 7EC992D08C94D82688BEB12E73EE493CEF04E5C3D0F35B4681F1989FAC6A42AD
- redisx-linux-arm64
  - SHA256: E05B66B796382C51FE63C9005B8A948BEE20605F62FCC5E28F3F9178C90CE154
- redisx-windows-amd64.exe
  - SHA256: F8A162C59DB07E709181DFEADD5FDBAEA277E26CDA98CB533DC8F33EB124A1C3

如何验证

- 在本地运行 `go test ./...` 确认所有单元和集成测试通过。
- 运行二进制并使用 `redis-cli` 验证 PEXPIRE/PTTL、SET PX、INCR、MGET、以及 MaxConns/MaxMemory 的行为。

感谢贡献者与社区的反馈与审查。