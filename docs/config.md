# 配置说明

当前实现使用 Viper 从 YAML 文件读取配置。下面是按代码实际读取路径整理的配置项。

## 示例配置

```yaml
mode: debug
biz_id: 1001
biz_name: account-service

grpc:
  listen: ":9000"

http:
  listen: ":8080"
  # 可选，若启用 https 需为 map[string]string 结构
  # ssl:
  #   cert_file: "/path/to/cert.pem"
  #   key_file: "/path/to/key.pem"

metrics:
  listen: ":9090"

logger:
  # 可选；未配置或设为空时仅输出 stdout
  file_path: "./logs"
  max_size: 100
  max_age: 7
  max_backups: 10

registry:
  provider: consul # 留空即关闭注册
  name: account.grpc
  address: 127.0.0.1
  port: 9000
  health_check:
    interval: "10s"
    timeout: "3s"

consul:
  address: "127.0.0.1:8500"

discovery:
  provider: dns # dns | consul；留空即关闭
  refresh_interval: "10s" # 仅 consul 使用
  timeout: "3s" # 仅 consul 使用，单次健康实例查询超时
  services:
    account: "dns:///account:9000"

etcd:
  endpoints:
    - "127.0.0.1:2379"

mysql:
  dsn: "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
  pool:
    max_open_conns: 100
    max_idle_conns: 20
    max_lifetime: 300

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0

jwt:
  access_secret: "replace-me"
  access_expire: 2      # hour
  refresh_expire: 7     # day

smtp:
  host: "smtp.example.com"
  port: 465
  username: "noreply@example.com"
  password: "replace-me"
```

## 配置项说明

- `mode`：运行模式。
- `logger.file_path`：可选。日志始终输出到 stdout；配置此目录时，额外将低于 `error` 级别的日志轮转写入 `info.log`，将 `error` 及以上日志轮转写入 `error.log`。未配置或为空时禁用文件轮转。
- `grpc.listen`：gRPC 监听地址。
- `http.listen`：HTTP/gateway 监听地址。
- `http.ssl`：可选；启用 https 时需要 `cert_file` 与 `key_file`。
- `metrics.listen`：Prometheus 指标监听地址。
- `registry.provider`：注册实现；当前默认工厂支持 `consul`，留空或 `none` 时不注册。
- `registry.*`：启用注册时使用的服务实例信息。
- `discovery.provider`：下游服务寻址策略；`dns` 适用于固定地址和 Kubernetes Service DNS，`consul` 使用健康实例 resolver。
- `discovery.services`：`dns` 策略下逻辑服务名到 gRPC target 的映射。
- `discovery.timeout`：Consul 单次健康实例查询超时，默认 `3s`。
- `consul.address`：Consul 地址。
- `etcd.endpoints`：Etcd endpoint 列表。
- `mysql.*`：MySQL 连接与连接池。
- `redis.*`：Redis 连接参数。
- `jwt.*`：JWT 签名密钥与过期策略。
- `smtp.*`：SMTP 发信参数。

敏感配置可通过 `MYSITE_MYSQL_DSN`、`MYSITE_REDIS_ADDR`、
`MYSITE_REDIS_PASSWORD`、`MYSITE_JWT_ACCESS_SECRET`、`MYSITE_SMTP_*` 和
`MYSITE_OSS_*` 环境变量覆盖，便于 Kubernetes Secret 注入。

## 注意事项

- 多处实现使用类型断言读取配置，类型不匹配会导致 panic；建议严格遵循示例类型。
- `http.ssl` 目前要求 `map[string]string` 形态。
- `biz_id` 与 `biz_name` 被 `BizInfoInterceptor` 用于补全当前服务信息。
