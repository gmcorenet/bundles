package config

import (
	"fmt"
	"os"

	"github.com/gmcorenet/framework/internal/config"
)

func Load(path string) *config.Config {
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
		if path == "" {
			path = "resources/config"
		}
	}

	cfg := config.GetInstance()
	cfg.Load(path)

	return cfg
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDSN() string {
	cfg := Load("")

	driver := cfg.Get("database.driver").(string)
	host := cfg.Get("database.host").(string)
	port := cfg.Get("database.port")
	name := cfg.Get("database.name").(string)
	user := cfg.Get("database.user").(string)
	password := cfg.Get("database.password").(string)

	return fmt.Sprintf("%s://%s:%s@%s:%v/%s", driver, user, password, host, port, name)
}