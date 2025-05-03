package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	ErrRegistryUnknownAuthType = errors.New("OCI registry auth type is unknown")
	ErrRegistryBasicAuthEmpty  = errors.New("oci registry basic auth is empty")
	ErrRegistryTokenEmpty      = errors.New("oci registry token is empty")
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
	GithubWebhookSecretEnable bool `envconfig:"GITHUB_WEBHOOK_SECRET_ENABLE" default:"true"`

	JwtTtl time.Duration `envconfig:"JWT_TTL" default:"5m"`

	DockerRegistry string `envconfig:"DOCKER_REGISTRY" required:"true"`

	DbDsn         string `envconfig:"DB_DSN" required:"true"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR" required:"true"`

	HttpPort string `envconfig:"HTTP_PORT" default:"8000"`

	BuilderPackage string `envconfig:"BUILDER_PACKAGE" required:"false"`

	KubeConfig FileSource `envconfig:"KUBE_CONFIG" required:"true"`

	AuthPrivateKey  StringBase64  `envconfig:"AUTH_PRIVATE_KEY" required:"true"`
	AuthPublicKey   StringBase64  `envconfig:"AUTH_PUBLIC_KEY" required:"true"`
	AuthTtl         time.Duration `envconfig:"AUTH_TTL" default:"24h"`
	AuthRedirectUrl string        `envconfig:"AUTH_REDIRECT_URL" required:"true"`

	// Host is a main app host to provide a quick preview for the deployed apps
	Host string `envconfig:"HOST" required:"true"`

	RegistryTLSVerify bool   `envconfig:"REGISTRY_TLS_VERIFY" default:"true"`
	RegistryCertDir   string `envconfig:"REGISTRY_CERT_DIR" required:"false"`
	RegistryAuthType  string `envconfig:"REGISTRY_AUTH_TYPE" required:"true"`
	RegistryUsername  string `envconfig:"REGISTRY_AUTH_USERNAME" required:"false"`
	RegistryPassword  string `envconfig:"REGISTRY_AUTH_PASSWORD" required:"false"`
	RegistryToken     string `envconfig:"REGISTRY_AUTH_TOKEN" required:"false"`

	// HTTP settings
	CorsAllowOrigin string `envconfig:"CORS_ALLOW_ORIGIN" required:"true"`
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

const (
	OciAuthTypeNoauth = "noauth"
	OciAuthTypeBasic  = "basic"
	OciAuthTypeToken  = "token"
)

func NewConfig() (Config, error) {
	conf := Config{}

	if err := envconfig.Process("", &conf); err != nil {
		return conf, err
	}

	if conf.RegistryAuthType != OciAuthTypeNoauth && conf.RegistryAuthType != OciAuthTypeBasic && conf.RegistryAuthType != OciAuthTypeToken {
		return conf, fmt.Errorf("given '%s': %w", conf.RegistryAuthType, ErrRegistryUnknownAuthType)
	}
	if conf.RegistryAuthType == OciAuthTypeBasic && (conf.RegistryUsername == "" || conf.RegistryPassword == "") {
		return conf, ErrRegistryBasicAuthEmpty
	}
	if conf.RegistryAuthType == OciAuthTypeToken && conf.RegistryToken == "" {
		return conf, ErrRegistryTokenEmpty
	}

	return conf, nil
}
