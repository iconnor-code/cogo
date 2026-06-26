# 拦截器说明

本文档描述 `interceptor/` 下各 gRPC Unary 拦截器的职责与建议顺序。

## 拦截器列表

- `SrvCtxInterceptor(config, logger)`
  - 注入 `ISrvCtx` 到 `context`。
  - 其他拦截器依赖它提供的 logger/config。

- `RecoveryInterceptor()`
  - 捕获 panic，记录日志，并返回统一内部错误（`UnknownErrCode`）。

- `CycleCheckInterceptor()`
  - 读取 `metadata[caller_methods]` 检查循环调用。
  - 当前方法会追加到 outgoing metadata 中。

- `BizInfoInterceptor()`
  - 从配置注入当前 `biz_id` / `biz_name`。
  - 从 incoming metadata 读取调用链 `biz_id` / `biz_name`。

- `UserInfoInterceptor(whiteList...)`
  - 从 `metadata[access_token]` 读取 JWT。
  - 解析后将 `user_id` / `user_email` / `is_admin` 写入 `ISrvCtx`。
  - `whiteList` 中的方法跳过鉴权。

- `RequestLogInterceptor()`
  - 记录请求耗时。
  - 对健康检查方法 `grpc.health.v1.Health/Check` 做了日志过滤。
  - 对标准 gRPC status 错误保持透传，对内部错误进行统一包装返回。

## 建议顺序

推荐链路：

1. `SrvCtxInterceptor`
2. `RecoveryInterceptor`
3. `CycleCheckInterceptor`
4. `BizInfoInterceptor`
5. `UserInfoInterceptor`
6. `RequestLogInterceptor`

说明：

- `SrvCtxInterceptor` 应放最前，否则后续拦截器读取 `core.SrvCtx` 会失败。
- `RecoveryInterceptor` 早放，尽快兜底 panic。
- `RequestLogInterceptor` 放后面，便于拿到尽量完整的执行信息。

## Metadata 约定

- `access_token`：用户访问 token（`UserInfoInterceptor` 使用，包含 `user_id` / `user_email` / `is_admin`）
- `biz_id`：上游业务 ID，可多值
- `biz_name`：上游业务名，可多值
- `caller_methods`：调用方法链（循环调用检查）
