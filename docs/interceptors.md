# 拦截器说明

本文档描述 `interceptor/` 下各 gRPC Unary 拦截器的职责与建议顺序。

## 拦截器列表

- `SrvCtxInterceptor(config, logger)`
  - 注入 `ISrvCtx` 到 `context`。
  - 其他拦截器依赖它提供的 logger/config。

- `RecoveryInterceptor()`
  - 捕获 panic，只记录方法、panic 和堆栈，不记录请求或响应载荷。
  - 返回内部错误，由外层 `ErrorInterceptor` 转换为安全的 gRPC `Internal`。

- `ErrorInterceptor()`
  - 将 `cerrs.CError` 的 transport-neutral `Kind` 映射为稳定的 gRPC code。
  - 只向客户端返回 `PublicMessage`；内部 cause 和调用位置不会进入响应。
  - 未分类错误统一返回 `Internal: internal error occurred`。

- `CycleCheckInterceptor()`
  - 读取 `metadata[caller_methods]` 检查循环调用。
  - 当前方法会追加到 outgoing metadata 中。

- `BizInfoInterceptor()`
  - 从配置注入当前 `biz_id` / `biz_name`。
  - 从 incoming metadata 读取调用链 `biz_id` / `biz_name`。

- `UserInfoInterceptor(whiteList...)`
  - 从 `metadata[authorization]` 读取 `Bearer <JWT>`。
  - 校验 JWT 签名、过期时间 `exp` 和 token ID `jti`。
  - 可选接入 `TokenRevocationChecker`，用于检查 token 是否已被撤销。
  - 解析后将 `user_id` / `user_email` / `is_admin` 写入 `ISrvCtx`。
  - `whiteList` 中的方法跳过鉴权。

- `RequestLogInterceptor()`
  - 只记录方法、耗时和最终 gRPC code，不记录 request、response 或 context。
  - 对健康检查方法 `grpc.health.v1.Health/Check` 做了日志过滤。
  - 取消请求记为 `Info`，预期业务失败记为 `Warn`，服务端失败记为 `Error`。

## 建议顺序

推荐链路：

1. `SrvCtxInterceptor`
2. `RequestLogInterceptor`
3. `ErrorInterceptor`
4. `RecoveryInterceptor`
5. `CycleCheckInterceptor`
6. `BizInfoInterceptor`
7. `UserInfoInterceptor`

说明：

- `SrvCtxInterceptor` 应放最前，否则后续拦截器读取 `core.SrvCtx` 会失败。
- `RequestLogInterceptor` 位于错误边界外层，按最终 gRPC code 记录结果。
- `ErrorInterceptor` 位于 `RecoveryInterceptor` 外层，确保 panic 恢复结果也经过安全错误映射。
- 业务 handler 返回有明确 `Kind` 的错误，不在 handler 中解析错误字符串。

## 错误分类

业务和应用代码使用 `cerrs` 的语义构造函数：

- `InvalidArgument` → `InvalidArgument`
- `Unauthenticated` → `Unauthenticated`
- `PermissionDenied` → `PermissionDenied`
- `NotFound` → `NotFound`
- `AlreadyExists` → `AlreadyExists`
- `FailedPrecondition` → `FailedPrecondition`
- `Unavailable` → `Unavailable`

`cerrs.New` 和 `cerrs.Wrap` 表示内部错误。数据库、Redis、服务发现等 cause 可以保留用于服务端日志，但不会作为公开消息返回。

## Metadata 约定

- `authorization`：`Bearer <JWT>` 用户访问令牌（`UserInfoInterceptor` 使用，包含 `user_id` / `user_email` / `is_admin` / `exp` / `jti`）
- `biz_id`：上游业务 ID，可多值
- `biz_name`：上游业务名，可多值
- `caller_methods`：调用方法链（循环调用检查）

## 用户身份拦截器

`UserInfoInterceptor` 负责统一处理 gRPC 用户身份：

- 从 incoming metadata 读取 `authorization: Bearer <JWT>`。
- 使用当前服务配置中的 JWT secret 校验 token。
- 要求 token 包含标准过期时间 `exp` 和 token ID `jti`。
- 可选调用撤销检查器，拒绝已退出登录或已轮换的 token。
- 将用户信息写入 `ISrvCtx`，业务代码通过 `core.SrvCtxFromContext(ctx).GetUserInfo()` 读取。

### JWT Claims

access token 至少需要包含：

```json
{
  "user_id": 123,
  "user_email": "user@example.com",
  "is_admin": false,
  "exp": 1893456000,
  "jti": "token-id"
}
```

`cogo/pkg/token.JwtToken.GenerateToken` 会自动写入 `exp` 和 `jti`。服务侧不要手动拼接 JWT，优先使用该封装生成 token。

### 使用方式

通过 `core/impl/server.NewGrpcServiceServer` 创建 gRPC 服务时，默认会加入 `UserInfoInterceptor`。业务服务只需要配置公开方法：

```go
server.NewGrpcServiceServer(config, logger, server.GrpcServiceOption{
	PublicMethods: []string{
		accountpb.AuthService_Login_FullMethodName,
		accountpb.AuthService_Register_FullMethodName,
	},
	RegisterServices: func(baseServer *grpc.Server) error {
		accountpb.RegisterAuthServiceServer(baseServer, authServer)
		return nil
	},
})
```

`PublicMethods` 中的方法会跳过用户鉴权。健康检查方法会由框架自动加入公开方法列表。

如果服务需要支持退出登录后 access token 立即失效，实现 `TokenRevocationChecker` 并注入：

```go
type TokenRevocationStore interface {
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
}

server.NewGrpcServiceServer(config, logger, server.GrpcServiceOption{
	PublicMethods:          publicMethods,
	TokenRevocationChecker: tokenRevocationStore,
	RegisterServices:       registerServices,
})
```

撤销检查器只接收 `jti`，不需要解析 JWT，也不需要知道业务用户表。典型实现是 Redis 黑名单：退出登录或刷新 token 时写入 `revoked:{jti}`，TTL 设置为 `exp - now`。

### 基本流程

1. 请求进入 gRPC interceptor 链。
2. `SrvCtxInterceptor` 注入 `ISrvCtx`。
3. `UserInfoInterceptor` 判断当前方法是否在公开方法列表中。
4. 公开方法直接放行。
5. 非公开方法读取 `metadata[authorization]` 中的 `Bearer <JWT>`。
6. 校验 JWT 签名、签名算法和 `exp`。
7. 读取 `jti`，如果配置了 `TokenRevocationChecker`，检查 token 是否已撤销。
8. 读取 `user_id` / `user_email` / `is_admin`，写入 `ISrvCtx.UserInfo`。
9. 调用业务 handler。

### 错误语义

- 缺少或格式错误的 `Authorization: Bearer <JWT>`：返回 `Unauthenticated`。
- JWT 非法、过期、缺少 `jti` 或已撤销：返回 `Unauthenticated`。
- 撤销检查器自身失败，例如 Redis 不可用：返回 `Internal`。
