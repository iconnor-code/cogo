# 变更记录

## 2026-05-24

### 修复

- 修复 `IDiscovery` 接口与 `KitConsulDiscovery` 实现签名不一致问题。
- 修复 `SrvCtx` 扩展字段 map 未初始化导致的 panic。
- 修复 logger 全局字段未实际生效问题，并增强 `AddGlobalFields` 类型安全。
- 修复 Etcd 注册续租逻辑中的 panic 风险与实例 ID 初始化问题。
- 修复 `RecoveryInterceptor` 捕获 panic 后未返回错误的问题。
- 修复 SMTP 发送流程并发请求下共享状态串写问题。
- 修复 `UserInfoInterceptor` 对 JWT claims 的不安全类型断言，增加 `user_id` 安全转换和边界校验。

### 测试

- 新增 `cerrs` 单元测试：错误创建、包装、解包与匹配。
- 新增 `srvctx` 单元测试：上下文字段、业务信息、用户信息存取。
- 新增 `RecoveryInterceptor` 单元测试：panic 恢复与错误码断言。
- 新增 `UserInfoInterceptor` 单元测试：鉴权成功、非法 token、数值转换边界。
