package api

import (
	"encoding/base64"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GithubAppClientID   string       `envconfig:"GITHUB_APP_CLIENT_ID" required:"true"`
	GithubAppPrivateKey StringBase64 `envconfig:"GITHUB_APP_PRIVATE_KEY" required:"true"`
	GithubAppSecret     string       `envconfig:"GITHUB_APP_SECRET" required:"true"`

	JwtTtl time.Duration `envconfig:"JWT_TTL" default:"5m"`

	DoToken        string `envconfig:"DO_TOKEN" required:"true"`
	DockerRegistry string `envconfig:"DOCKER_REGISTRY" required:"true"`

	DbDsn         string `envconfig:"DB_DSN" required:"true"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR" required:"true"`

	HttpPort string `envconfig:"HTTP_PORT" default:"8000"`

	AuthID       string       `envconfig:"AUTH_ID" required:"true"`
	AuthSecret   StringBase64 `envconfig:"AUTH_SECRET" required:"true"`
	AuthKeyID    string       `envconfig:"AUTH_KEY_ID" required:"true"`
	AuthEndpoint string       `envconfig:"AUTH_ENDPOINT" required:"true"`

	GithubWebhookSecret       string `envconfig:"GITHUB_WEBHOOK_SECRET" required:"true"`
	GithubWebhookSecretEnable bool   `envconfig:"GITHUB_WEBHOOK_SECRET_ENABLE" default:"true"`

	BuilderPackage string `envconfig:"BUILDER_PACKAGE" required:"false"`

	KubeConfig string `envconfig:"KUBE_CONFIG" required:"true"`
}

type StringBase64 string

func (s *StringBase64) Decode(value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}
	*s = StringBase64(decoded)
	return nil
}

func NewConfig() (Config, error) {
	conf := Config{}

	if err := envconfig.Process("", &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
