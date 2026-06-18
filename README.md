# cogo

`cogo` 是一个 Go 微服务基础组件库，提供配置、日志、服务启动、注册发现、拦截器和常见客户端封装。

## 功能概览

- 核心接口抽象：`IConfig`、`ILogger`、`IServer`、`IRegistry`、`IDiscovery`、`ISrvCtx`
- 服务启动能力：gRPC、HTTP(gRPC-Gateway)、Prometheus Metrics
- 服务治理：Consul / Nacos / Etcd 注册，Consul / Nacos 发现 + 负载均衡 + 重试
- 中间件：请求日志、恢复、上下文注入、用户信息解析、业务信息透传、循环调用检测
- 客户端封装：MySQL(Gorm)、Redis、Consul、Nacos、Etcd
- 工具包：JWT、Captcha、SMTP

## 目录结构

```text
cerrs/         # 错误模型
client/        # 外部依赖客户端封装
core/          # 接口定义
core/impl/     # 接口实现
interceptor/   # gRPC unary 拦截器
pkg/           # 业务通用工具
utils/         # 杂项工具
docs/          # 项目文档
```

## 快速接入示意

1. 初始化配置与日志。
2. 初始化基础客户端（如 MySQL、Redis、Consul/Etcd）。
3. 按需组装拦截器链。
4. 创建并启动 gRPC/HTTP/Metrics 服务。

参考文档：

- [配置说明](docs/config.md)
- [架构说明](docs/architecture.md)
- [拦截器说明](docs/interceptors.md)
- [变更记录](docs/changelog.md)

## 最低要求

- Go `1.23+`（当前 `go.mod` 为 `go 1.23.0`）

## 测试

```bash
go test ./...
```
