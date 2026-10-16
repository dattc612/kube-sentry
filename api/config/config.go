package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dattc/go-though-storm/api/common"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*models.Config, error) {
	v := viper.New()

	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("namespace.allowed", []string{})
	v.SetDefault("auth.local.jwtSecret", common.GenerateJWTSecret())

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg models.Config
	if err := v.Unmarshal(&cfg,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			StringToByteSliceHookFunc()))); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}
	return &cfg, nil
}

// Keycloak AuthConfig represents the configuration for Keycloak authentication.
func IssuerURL(cfg *models.KeycloakAuthConfig) string {
	return fmt.Sprintf("%s/realms/%s", cfg.URL, cfg.Realm)
}

func StringToByteSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t == reflect.TypeOf([]byte{}) {
			return []byte(data.(string)), nil
		}
		return data, nil
	}
}
