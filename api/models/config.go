package models

import "github.com/golang-jwt/jwt/v5"

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Kube          KubeConfig          `mapstructure:"kube"`
	Namespace     NamespaceConfig     `mapstructure:"namespace"`
	EditableFiles EditableFilesConfig `mapstructure:"editable_files"`
	Auth          AuthConfig          `mapstructure:"auth"`
}

type AuthConfig struct {
	Users     []UserCredentials `mapstructure:"users"`
	JwtSecret []byte            `mapstructure:"jwt"`
}
type UserCredentials struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Group    string `mapstructure:"group"`
}
type Claims struct {
	Username string `json:"username"`
	Group    string `json:"group"`
	jwt.RegisteredClaims
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type KubeConfig struct {
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
}

type NamespaceConfig struct {
	Allowed []string `mapstructure:"allowed"`
}

type EditableFilesConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Patterns []string `mapstructure:"patterns"`
}
