package guiio_util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sphynx/config"
)

var (
	defaultValue = map[string]any{
		"server_name":       "guiio",
		"port":              8080,
		"db_max_idle_conns": 10,
		"db_max_conns":      100,
		"db_max_timout":     10,
	}
)

// convertValue converts JSON value to appropriate type based on default value
func convertValue(key string, value any) any {
	if defaultValue, exists := defaultValue[key]; exists {
		switch defaultValue.(type) {
		case int:
			if floatVal, ok := value.(float64); ok {
				return int(floatVal)
			}
		case int64:
			if floatVal, ok := value.(float64); ok {
				return int64(floatVal)
			}
		case float64:
			return value
		case string:
			return fmt.Sprintf("%v", value)
		case bool:
			return value
		}
	}
	return value
}

func GetEnv() (map[string]config.ConfigValue[any], error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	env, err := os.ReadFile(pwd + "/env/.env.json")

	if err != nil {
		env, err = os.ReadFile(pwd + "/bin/.env.json")
		if err != nil {
			return nil, err
		}
	}

	envMap := make(map[string]any)
	if err := json.Unmarshal(env, &envMap); err != nil {
		return nil, err
	}

	for key, val := range defaultValue {
		if _, ok := envMap[key]; !ok {
			envMap[key] = val
		}
	}

	envConfig := make(map[string]config.ConfigValue[any])

	for k, v := range envMap {
		convertedValue := convertValue(k, v)

		defaultVal := defaultValue[k]

		envConfig[k] = config.ConfigValue[any]{
			Value:   convertedValue,
			Default: defaultVal,
		}
	}

	return envConfig, nil
}
