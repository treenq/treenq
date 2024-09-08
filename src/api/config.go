package api

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	JwtKey    string        `envconfig:"JWT_KEY" required:"true"`
	JwtSecret string        `envconfig:"JWT_SECRET" required:"true"`
	JwtTtl    time.Duration `envconfig:"JWT_TTL" default:"5m"`

	DoToken string `envconfig:"DO_TOKEN" required:"true"`

	DbDsn         string `envconfig:"DB_DSN" required:"true"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR" required:"true"`

	HttpPort string `envconfig:"HTTP_PORT" default:"8000"`

	AuthID             string `envconfig:"AUTH_ID" required:"true"`
	AuthSecret         string `envconfig:"AUTH_SECRET" required:"true"`
	AuthEndpoint       string `envconfig:"AUTH_ENDPOINT" required:"true"`
	AuthResource       string `envconfig:"AUTH_RESOURCE"`
	SignInRedirectUri  string `envconfig:"SIGN_IN_REDIRECT_URI" required:"true"`
	SignOutRedirectUri string `envconfig:"SIGN_OUT_REDIRECT_URI" required:"true"`
	SessionSecret      string `envconfig:"SESSION_SECRET" required:"true"`
}

func NewConfig() (Config, error) {
	conf := Config{}

	if err := envconfig.Process("", &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
