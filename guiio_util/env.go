package guiio_util

import (
	"encoding/json"
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

	envConfig := make(map[string]config.ConfigValue[any])

	for k, v := range envMap {
		d := v
		if dv, ok := defaultValue[k]; !ok {
			d = dv
		}

		envConfig[k] = config.ConfigValue[any]{
			Value:   v,
			Default: d,
		}

	}

	return envConfig, nil
}
