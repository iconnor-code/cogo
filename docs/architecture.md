# 架构说明

## 目标

`cogo` 的目标是把微服务中的通用能力沉淀成可复用组件，业务层只需关注 RPC/HTTP 接口与业务逻辑。

## 核心抽象

- `core/IConfig`：配置重载
- `core/ILogger`：统一日志能力
- `core/IServer`：服务生命周期（`Start` / `Stop`）
- `core/IRegistry`：服务注册与反注册
- `core/ISrvCtx`：并发安全的请求级上下文（logger/config/biz/user/扩展字段）

## 实现分层

- `core/impl/config`：Viper 本地文件加载
- `core/impl/logger`：Zap + Lumberjack
- `core/impl/server`：
  - `grpc.go`：gRPC 服务启动与关闭
  - `http.go`：HTTP/gRPC-Gateway 服务启动与关闭（支持 TLS）
  - `metrics.go`：Prometheus 指标暴露
- `core/impl/registry`：Consul / Etcd 注册实现
- `core/impl/rpcclient`：推荐的 gRPC 连接池；支持 DNS/Kubernetes 与 Consul resolver、客户端负载均衡和统一关闭
- `core/impl/srvctx`：请求上下文实现

## 典型请求流程

1. 请求进入 gRPC 服务。
2. `SrvCtxInterceptor` 注入 `ISrvCtx`。
3. `RequestLogInterceptor` 在最外层观察最终状态。
4. `ErrorInterceptor` 统一错误边界，`RecoveryInterceptor` 兜底 panic。
5. 循环检查、业务信息和用户身份拦截器补充上下文。
6. 业务 Handler 执行并返回具有明确 Kind 的应用错误。

## 设计特点

- 接口优先，便于替换实现。
- 组件组合式使用，按服务实际需求选择。
- 通过 `ISrvCtx` 在拦截器与业务层传递统一上下文。
- 业务 gateway 依赖 protobuf 生成的类型化客户端；注册、发现和连接生命周期留在基础设施层。
