package secret

import (
	"sync"

	infisical "github.com/infisical/go-sdk"
)

type Secret struct {
	client infisical.InfisicalClientInterface
	mu     sync.Mutex
}

func NewSecret(secret *Secret) (*Secret, error) {
	secret.mu.Lock()
	client := infisical.NewInfisicalClient(infisical.Config{
		SiteUrl: "https://app.infisical.com",
	})
	// use INFISICAL_UNIVERSAL_AUTH_CLIENT_ID, INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET env
	_, err := client.Auth().UniversalAuthLogin("", "")
	if err != nil {
		return nil, err
	}
	secret.client = client
	defer secret.mu.Unlock()
	return secret, nil
}

func (se *Secret) GetSecret(secretKey string, env string, projectId string) (infisical.Secret, error) {
	return se.client.Secrets().Retrieve(
		infisical.RetrieveSecretOptions{
			SecretKey:   secretKey,
			Environment: env,
			ProjectID:   projectId,
			SecretPath:  "/",
		},
	)
}
