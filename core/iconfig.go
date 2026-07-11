// Package core core interface
package core

type IConfig interface {
	GetMode() string
	GetBizID() int
	GetBizName() string
	GetGRPC() GRPCConfig
	GetHTTP() HTTPConfig
	GetLogger() LoggerConfig
	GetMetrics() MetricsConfig
	GetMySQL() MySQLConfig
	GetRedis() RedisConfig
	GetEtcd() EtcdConfig
	GetConsul() ConsulConfig
	GetDiscovery() DiscoveryConfig
	GetRegistry() RegistryConfig
	GetSMTP() SMTPConfig
	GetJWT() JWTConfig
	GetAdmin() AdminConfig
	GetOSS() OSSConfig
	Reload() error
}

type Config struct {
	Mode    string `mapstructure:"mode" yaml:"mode"`
	BizID   int    `mapstructure:"biz_id" yaml:"biz_id"`
	BizName string `mapstructure:"biz_name" yaml:"biz_name"`

	GRPC      GRPCConfig      `mapstructure:"grpc" yaml:"grpc"`
	HTTP      HTTPConfig      `mapstructure:"http" yaml:"http"`
	Logger    LoggerConfig    `mapstructure:"logger" yaml:"logger"`
	Metrics   MetricsConfig   `mapstructure:"metrics" yaml:"metrics"`
	MySQL     MySQLConfig     `mapstructure:"mysql" yaml:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis" yaml:"redis"`
	Etcd      EtcdConfig      `mapstructure:"etcd" yaml:"etcd"`
	Consul    ConsulConfig    `mapstructure:"consul" yaml:"consul"`
	Discovery DiscoveryConfig `mapstructure:"discovery" yaml:"discovery"`
	Registry  RegistryConfig  `mapstructure:"registry" yaml:"registry"`
	SMTP      SMTPConfig      `mapstructure:"smtp" yaml:"smtp"`
	JWT       JWTConfig       `mapstructure:"jwt" yaml:"jwt"`
	Admin     AdminConfig     `mapstructure:"admin" yaml:"admin"`
	OSS       OSSConfig       `mapstructure:"oss" yaml:"oss"`
}

type GRPCConfig struct {
	Listen          string `mapstructure:"listen" yaml:"listen"`
	GatewayEndpoint string `mapstructure:"gateway_endpoint" yaml:"gateway_endpoint"`
}

type HTTPConfig struct {
	Listen string        `mapstructure:"listen" yaml:"listen"`
	SSL    HTTPSSLConfig `mapstructure:"ssl" yaml:"ssl"`
}

type HTTPSSLConfig struct {
	CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
	KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
}

type LoggerConfig struct {
	Level      int    `mapstructure:"level" yaml:"level"`
	FilePath   string `mapstructure:"file_path" yaml:"file_path"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age"`
}

type MetricsConfig struct {
	Enable bool   `mapstructure:"enable" yaml:"enable"`
	Listen string `mapstructure:"listen" yaml:"listen"`
	Prefix string `mapstructure:"prefix" yaml:"prefix"`
}

type MySQLConfig struct {
	DSN  string          `mapstructure:"dsn" yaml:"dsn"`
	Pool MySQLPoolConfig `mapstructure:"pool" yaml:"pool"`
}

type MySQLPoolConfig struct {
	MaxOpenConns int `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	MaxLifetime  int `mapstructure:"max_lifetime" yaml:"max_lifetime"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr" yaml:"addr"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"db" yaml:"db"`
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints" yaml:"endpoints"`
}

type ConsulConfig struct {
	Address string `mapstructure:"address" yaml:"address"`
}

// DiscoveryConfig selects how logical service names are resolved for gRPC
// clients. Provider is either "dns" or "consul"; an empty provider disables
// discovery until a caller requests a downstream connection.
type DiscoveryConfig struct {
	Provider        string            `mapstructure:"provider" yaml:"provider"`
	RefreshInterval string            `mapstructure:"refresh_interval" yaml:"refresh_interval"`
	Timeout         string            `mapstructure:"timeout" yaml:"timeout"`
	Services        map[string]string `mapstructure:"services" yaml:"services"`
}

type RegistryConfig struct {
	Provider    string                    `mapstructure:"provider" yaml:"provider"`
	Name        string                    `mapstructure:"name" yaml:"name"`
	Address     string                    `mapstructure:"address" yaml:"address"`
	Port        int                       `mapstructure:"port" yaml:"port"`
	HealthCheck RegistryHealthCheckConfig `mapstructure:"health_check" yaml:"health_check"`
}

type RegistryHealthCheckConfig struct {
	Interval string `mapstructure:"interval" yaml:"interval"`
	Timeout  string `mapstructure:"timeout" yaml:"timeout"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret" yaml:"access_secret"`
	AccessExpire  int    `mapstructure:"access_expire" yaml:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire" yaml:"refresh_expire"`
}

type AdminConfig struct {
	UserIDs []int `mapstructure:"user_ids" yaml:"user_ids"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name" yaml:"bucket_name"`
	BaseURL         string `mapstructure:"base_url" yaml:"base_url"`
	UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
	PresignExpire   int    `mapstructure:"presign_expire" yaml:"presign_expire"`
}
