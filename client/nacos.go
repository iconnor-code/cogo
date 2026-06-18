// Package client provides external dependency clients.
package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const defaultNacosGroupName = "DEFAULT_GROUP"

type Nacos struct {
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
	groupName    string
	clusterName  string
}

func NewNacos(config core.IConfig) (*Nacos, error) {
	serverConfigs, err := nacosServerConfigs(config)
	if err != nil {
		return nil, err
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         optionalString(config, "nacos.namespace_id"),
		Username:            optionalString(config, "nacos.username"),
		Password:            optionalString(config, "nacos.password"),
		TimeoutMs:           uint64(optionalInt(config, "nacos.timeout_ms", 5000)),
		NotLoadCacheAtStart: optionalBool(config, "nacos.not_load_cache_at_start", true),
		LogDir:              optionalString(config, "nacos.log_dir"),
		CacheDir:            optionalString(config, "nacos.cache_dir"),
		LogLevel:            optionalString(config, "nacos.log_level"),
	}
	if clientConfig.LogDir == "" {
		clientConfig.LogDir = "/tmp/nacos/log"
	}
	if clientConfig.CacheDir == "" {
		clientConfig.CacheDir = "/tmp/nacos/cache"
	}
	if clientConfig.LogLevel == "" {
		clientConfig.LogLevel = "info"
	}

	param := vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	}
	namingClient, err := clients.NewNamingClient(param)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	configClient, err := clients.NewConfigClient(param)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}

	groupName := optionalString(config, "nacos.group_name")
	if groupName == "" {
		groupName = defaultNacosGroupName
	}

	return &Nacos{
		namingClient: namingClient,
		configClient: configClient,
		groupName:    groupName,
		clusterName:  optionalString(config, "nacos.cluster_name"),
	}, nil
}

func (n *Nacos) NamingClient() naming_client.INamingClient {
	return n.namingClient
}

func (n *Nacos) ConfigClient() config_client.IConfigClient {
	return n.configClient
}

func (n *Nacos) GroupName() string {
	return n.groupName
}

func (n *Nacos) ClusterName() string {
	return n.clusterName
}

func nacosServerConfigs(config core.IConfig) ([]constant.ServerConfig, error) {
	serverValues, _ := optionalStringSlice(config, "nacos.servers")
	if len(serverValues) == 0 {
		address := optionalString(config, "nacos.address")
		if address == "" {
			return nil, cerrs.New("config \"nacos.servers\" or \"nacos.address\" is required")
		}
		serverValues = []string{address}
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(serverValues))
	for _, server := range serverValues {
		host, portText, err := net.SplitHostPort(server)
		if err != nil {
			return nil, cerrs.Wrap(err, fmt.Sprintf("invalid nacos server address %q", server))
		}
		port, err := strconv.ParseUint(portText, 10, 64)
		if err != nil {
			return nil, cerrs.Wrap(err, fmt.Sprintf("invalid nacos server port %q", server))
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: strings.Trim(host, "[]"),
			Port:   port,
			Scheme: optionalString(config, "nacos.scheme"),
		})
	}
	return serverConfigs, nil
}

func optionalString(config core.IConfig, key string) string {
	value := config.Get(key)
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func optionalInt(config core.IConfig, key string, fallback int) int {
	value := config.Get(key)
	if value == nil {
		return fallback
	}
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func optionalBool(config core.IConfig, key string, fallback bool) bool {
	value := config.Get(key)
	if value == nil {
		return fallback
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func optionalStringSlice(config core.IConfig, key string) ([]string, bool) {
	value := config.Get(key)
	if value == nil {
		return nil, false
	}
	switch v := value.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprint(item))
		}
		return result, true
	case string:
		if v == "" {
			return nil, true
		}
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
		return result, true
	}
	return nil, false
}
