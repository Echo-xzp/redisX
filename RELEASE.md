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