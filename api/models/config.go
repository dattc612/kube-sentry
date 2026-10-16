package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Server                 ServerConfig               `mapstructure:"server"`
	Kube                   KubeConfig                 `mapstructure:"kube"`
	Namespace              NamespaceConfig            `mapstructure:"namespace"`
	EditableFiles          EditableFilesConfig        `mapstructure:"editableFiles"`
	Auth                   AuthConfig                 `mapstructure:"auth"`
	Jobs                   []JobConfig                `mapstructure:"jobs"`
	AutomationExtraConfigs []AutomationJobExtraConfig `mapstructure:"automationExtraConfigs"`
}

type AuthConfig struct {
	Local    LocalAuthConfig    `mapstructure:"local"`
	Keycloak KeycloakAuthConfig `mapstructure:"keycloak"`
	Mode     string             `mapstructure:"mode"`
}

type LocalAuthConfig struct {
	Enabled   bool              `mapstructure:"enabled"`
	Users     []UserCredentials `mapstructure:"users"`
	JwtSecret []byte            `mapstructure:"jwtSecret"`
}

type KeycloakAuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	URL          string `mapstructure:"url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
}

type UserCredentials struct {
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Group    string   `mapstructure:"group"`
	Roles    []string `mapstructure:"roles"`
}

type Claims struct {
	Username string   `json:"username"`
	Group    string   `json:"group"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type KubeConfig struct {
	KubeconfigPath string `mapstructure:"kubeconfigPath"`
	Context        string `mapstructure:"context"`
}

type NamespaceConfig struct {
	Allowed []string `mapstructure:"allowed"`
}

type EditableFilesConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Patterns []string `mapstructure:"patterns"`
}

type JobConfig struct {
	ID          int          `mapstructure:"jobID"`
	JobName     string       `mapstructure:"jobName"`
	Description string       `mapstructure:"description"`
	Steps       []StepConfig `mapstructure:"steps"`
	AllowedPod  []string     `mapstructure:"allowedPod"`
}

type StepConfig struct {
	Name        string   `mapstructure:"name"`
	Description string   `mapstructure:"description"`
	Command     []string `mapstructure:"command"`
}

type AutomationJobExtraConfig struct {
	Code       string   `json:"code"`
	Enabled    bool     `json:"enabled"`
	AllowedPod []string `json:"allowed_pod,omitempty"`
}
