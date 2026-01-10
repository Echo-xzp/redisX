# redisX

redisX 是用于学习与验证的类 Redis 缓存服务，目标实现基础的 RESP 协议、SET/GET/DEL/EXISTS、键过期等功能（MVP）。

快速启动：

1. 安装 Go 1.21+
2. 在项目根目录运行：

```bash
# 在开发环境直接运行
go run ./cmd/redisx
# 或构建二进制
go build -o redisx ./cmd/redisx
./redisx
```

默认监听端口：随机端口（开发测试时使用 `:0` 启动），生产使用 `:6379`。

使用 redis-cli 测试示例：

```bash
# 使用 redis-cli 连接（假设监听 6379）
redis-cli -p 6379 PING
# 返回：PONG

# 基本命令
redis-cli -p 6379 SET mykey "hello"
redis-cli -p 6379 GET mykey
# 返回："hello"

# 带过期选项（EX 秒，PX 毫秒）
redis-cli -p 6379 SET tempkey "1" EX 1    # 1 秒后过期
redis-cli -p 6379 SET tempkey2 "2" PX 500 # 500ms -> 向上取整为 1 秒

# 查询信息
redis-cli -p 6379 INFO
```

开发与测试：

- 运行单元测试：

```bash
go test ./...
```

- 构建验证：

```bash
go build ./...
```

文档与工作记录：

- 开发过程中所有重要改动会记录在 `docs/work/工作记录.md`。

欢迎提交 issue 或 PR 来完善功能与修复 bug。
