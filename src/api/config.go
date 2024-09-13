package api

import (
	"encoding/base64"
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

	AuthID       string       `envconfig:"AUTH_ID" required:"true"`
	AuthSecret   StringBase64 `envconfig:"AUTH_SECRET" required:"true"`
	AuthKeyID    string       `envconfig:"AUTH_KEY_ID" required:"true"`
	AuthEndpoint string       `envconfig:"AUTH_ENDPOINT" required:"true"`
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
