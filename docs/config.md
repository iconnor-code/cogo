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
  file_path: "./logs"
  max_size: 100
  max_age: 7
  max_backups: 10

registry:
  name: account.grpc
  address: 127.0.0.1
  port: 9000
  health_check:
    interval: "10s"
    timeout: "3s"

consul:
  address: "127.0.0.1:8500"

etcd:
  endpoints:
    - "127.0.0.1:2379"

nacos:
  servers:
    - "127.0.0.1:8848"
  # 或使用单节点写法：
  # address: "127.0.0.1:8848"
  namespace_id: ""
  group_name: "DEFAULT_GROUP"
  cluster_name: ""
  username: ""
  password: ""
  timeout_ms: 5000
  not_load_cache_at_start: true
  log_dir: "/tmp/nacos/log"
  cache_dir: "/tmp/nacos/cache"
  log_level: "info"

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

- `mode`：日志模式；`debug` 时会额外输出到 stdout。
- `grpc.listen`：gRPC 监听地址。
- `http.listen`：HTTP/gateway 监听地址。
- `http.ssl`：可选；启用 https 时需要 `cert_file` 与 `key_file`。
- `metrics.listen`：Prometheus 指标监听地址。
- `registry.*`：服务注册信息。
- `consul.address`：Consul 地址。
- `etcd.endpoints`：Etcd endpoint 列表。
- `nacos.servers` / `nacos.address`：Nacos server 地址，`servers` 支持多节点，地址格式为 `host:port`。
- `nacos.namespace_id`：Nacos namespace ID。
- `nacos.group_name`：Nacos 注册发现和配置中心默认 group，未配置时为 `DEFAULT_GROUP`。
- `nacos.cluster_name`：Nacos 注册实例 cluster，未配置时为空。
- `nacos.username` / `nacos.password`：Nacos 鉴权信息，可选。
- `nacos.timeout_ms`：Nacos SDK 请求超时时间，默认 `5000`。
- `nacos.log_dir` / `nacos.cache_dir` / `nacos.log_level`：Nacos SDK 日志与缓存配置。
- `mysql.*`：MySQL 连接与连接池。
- `redis.*`：Redis 连接参数。
- `jwt.*`：JWT 签名密钥与过期策略。
- `smtp.*`：SMTP 发信参数。

## Nacos 配置中心

Nacos 配置中心通过 `core/impl/config` 的 option 接入：

```go
nacosClient, err := client.NewNacos(bootstrapConfig)
if err != nil {
	return err
}

conf, err := config.NewConfig(
	config.WithNacosClient(nacosClient),
	config.WithNacosConfig("account.yaml", "DEFAULT_GROUP", "yaml"),
)
```

`WithNacosConfig` 的第三个参数为空时，会根据 data ID 后缀推断配置类型；无后缀时默认 `yaml`。

## 注意事项

- 多处实现使用类型断言读取配置，类型不匹配会导致 panic；建议严格遵循示例类型。
- `http.ssl` 目前要求 `map[string]string` 形态。
- `biz_id` 与 `biz_name` 被 `BizInfoInterceptor` 用于补全当前服务信息。
