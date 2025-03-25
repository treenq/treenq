package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GithubClientID string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	// GithubPrivateKey is used to issue JWT in order to access the users' repositories
	GithubPrivateKey StringBase64 `envconfig:"GITHUB_PRIVATE_KEY" required:"true"`
	// GithubSecret is used to verify oauth code
	GithubSecret      string `envconfig:"GITHUB_SECRET" required:"true"`
	GithubRedirectURL string `envconfig:"GITHUB_REDIRECT_URL" required:"true"`
	// GithubWebhookSecret is used to verify the webhooks source
	GithubWebhookSecret string `envconfig:"GITHUB_WEBHOOK_SECRET" required:"true"`
	// TODO: Enable in e2e tests
	GithubWebhookSecretEnable bool   `envconfig:"GITHUB_WEBHOOK_SECRET_ENABLE" default:"true"`
	GithubWebhookURL          string `envconfig:"GITHUB_WEBHOOK_URL" required:"true"`

	JwtTtl time.Duration `envconfig:"JWT_TTL" default:"5m"`

	DockerRegistry string `envconfig:"DOCKER_REGISTRY" required:"true"`

	DbDsn         string `envconfig:"DB_DSN" required:"true"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR" required:"true"`

	HttpPort string `envconfig:"HTTP_PORT" default:"8000"`

	BuilderPackage string `envconfig:"BUILDER_PACKAGE" required:"false"`

	KubeConfig FileSource `envconfig:"KUBE_CONFIG" required:"true"`

	AuthPrivateKey StringBase64  `envconfig:"AUTH_PRIVATE_KEY" required:"true"`
	AuthPublicKey  StringBase64  `envconfig:"AUTH_PUBLIC_KEY" required:"true"`
	AuthTtl        time.Duration `envconfig:"AUTH_TTL" default:"24h"`

	// Host is a main app host to provide a quick preview for the deployed apps
	Host string `envconfig:"HOST" required:"true"`

	RegistryTLSVerify bool   `env:"REGISTRY_TLS_VERIFY" default:"false"`
	RegistryCertDir   string `env:"REGISTRY_CERT_DIR" default:"./certs"`
	RegistryAuthType  string `env:"REGISTRY_AUTH_TYPE" default:"token"`
	RegistryUsername  string `env:"REGISTRY_AUTH_USERNAME" default:"testuser"`
	RegistryPassword  string `env:"REGISTRY_AUTH_PASSWORD" default:"testpassword"`
	RegistryToken     string `env:"REGISTRY_AUTH_TOKEN" default:"dGVzdHVzZXI6dGVzdHBhc3N3b3Jk" `
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

type FileSource string

func (s *FileSource) Decode(value string) error {
	f, err := os.Open(value)
	if err != nil {
		return fmt.Errorf("failed to read file source config %s: %w", value, err)
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file source config %s: %w", value, err)
	}

	*s = FileSource(b)
	return nil
}

func NewConfig() (Config, error) {
	conf := Config{}

	if err := envconfig.Process("", &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
