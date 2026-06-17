package core

import (
	"fmt"
	"strings"
)

func GetString(config IConfig, key string) (string, error) {
	value, ok := lookupConfig(config, key)
	if !ok {
		return "", fmt.Errorf("config %q is required", key)
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("config %q must be string, got %T", key, value)
	}
	return str, nil
}

func GetInt(config IConfig, key string) (int, error) {
	value, ok := lookupConfig(config, key)
	if !ok {
		return 0, fmt.Errorf("config %q is required", key)
	}
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float64:
		if float64(int(v)) != v {
			return 0, fmt.Errorf("config %q must be integer, got %v", key, value)
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("config %q must be int, got %T", key, value)
	}
}

func GetStringSlice(config IConfig, key string) ([]string, error) {
	value, ok := lookupConfig(config, key)
	if !ok {
		return nil, fmt.Errorf("config %q is required", key)
	}
	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		res := make([]string, 0, len(v))
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("config %q[%d] must be string, got %T", key, i, item)
			}
			res = append(res, str)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("config %q must be []string, got %T", key, value)
	}
}

func lookupConfig(config IConfig, key string) (any, bool) {
	if config == nil {
		return nil, false
	}
	if value := config.Get(key); value != nil {
		return value, true
	}
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return nil, false
	}
	var current any = config.Get(parts[0])
	if current == nil {
		return nil, false
	}
	for _, part := range parts[1:] {
		switch node := current.(type) {
		case map[string]any:
			current = node[part]
		case map[string]string:
			current = node[part]
		case map[string]int:
			current = node[part]
		default:
			return nil, false
		}
		if current == nil {
			return nil, false
		}
	}
	return current, true
}
