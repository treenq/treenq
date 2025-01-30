package api

import (
	"encoding/base64"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GithubClientID            string       `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GithubPrivateKey          StringBase64 `envconfig:"GITHUB_PRIVATE_KEY" required:"true"`
	GithubSecret              string       `envconfig:"GITHUB_SECRET" required:"true"`
	GithubRedirectURL         string       `envconfig:"GITHUB_REDIRECT_URL" required:"true"`
	GithubWebhookSecret       string       `envconfig:"GITHUB_WEBHOOK_SECRET" required:"true"`
	GithubWebhookSecretEnable bool         `envconfig:"GITHUB_WEBHOOK_SECRET_ENABLE" default:"true"`
	GithubWebhookURL          string       `envconfig:"GITHUB_WEBHOOK_URL" required:"true"`

	JwtTtl time.Duration `envconfig:"JWT_TTL" default:"5m"`

	DoToken        string `envconfig:"DO_TOKEN" required:"true"`
	DockerRegistry string `envconfig:"DOCKER_REGISTRY" required:"true"`

	DbDsn         string `envconfig:"DB_DSN" required:"true"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR" required:"true"`

	HttpPort string `envconfig:"HTTP_PORT" default:"8000"`

	BuilderPackage string `envconfig:"BUILDER_PACKAGE" required:"false"`

	KubeConfig string `envconfig:"KUBE_CONFIG" required:"true"`

	AuthPrivateKey StringBase64  `envconfig:"AUTH_PRIVATE_KEY" required:"true"`
	AuthPublicKey  StringBase64  `envconfig:"AUTH_PUBLIC_KEY" required:"true"`
	AuthTtl        time.Duration `envconfig:"AUTH_TTL" default:"24h"`
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
