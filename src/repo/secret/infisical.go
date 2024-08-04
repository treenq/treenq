package secret

import (
	"sync"
	"time"

	infisical "github.com/infisical/go-sdk"
)

type Secret struct {
	client infisical.InfisicalClientInterface
	mu     sync.Mutex
	ttl    time.Time
	token  string
	secret infisical.Secret
}

func NewSecret(authClientId string, authClientSecret string) (*Secret, error) {
	client := infisical.NewInfisicalClient(infisical.Config{
		SiteUrl: "https://app.infisical.com",
	})

	secret := &Secret{
		client: client,
		mu:     sync.Mutex{},
	}

	// Issue the initial token
	if err := secret.issueToken(authClientId, authClientSecret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (se *Secret) issueToken(authClientId string, authClientSecret string) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	res, err := se.client.Auth().UniversalAuthLogin(authClientId, authClientSecret)
	if err != nil {
		return err
	}

	if res.ExpiresIn > int64(5*time.Second) {
		se.token = res.AccessToken
		se.ttl = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
	}

	return nil
}

func (se *Secret) GetSecret(secretKey, env, projectId, authClientId, authClientSecret string) error {
	se, err := NewSecret(authClientId, authClientSecret)

	if err != nil {
		return err
	}

	envSecret, err := se.client.Secrets().Retrieve(
		infisical.RetrieveSecretOptions{
			SecretKey:   secretKey,
			Environment: env,
			ProjectID:   projectId,
			SecretPath:  "/",
		},
	)
	if err != nil {
		return err
	}
	se.secret = envSecret
	return nil
}
