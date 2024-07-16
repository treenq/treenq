package secret

import (
	infisical "github.com/infisical/go-sdk"
)

type Secret struct {
	client infisical.InfisicalClientInterface
}

func NewSecretsManager(clientId string, secret string) (*Secret, error) {
	client := infisical.NewInfisicalClient(infisical.Config{
		SiteUrl: "https://app.infisical.com",
	})
	// use INFISICAL_UNIVERSAL_AUTH_CLIENT_ID, INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET env
	_, err := client.Auth().UniversalAuthLogin(clientId, secret)
	if err != nil {
		return nil, err
	}

	return &Secret{
		client: client,
	}, nil
}

func (sm *Secret) GetSecret(secretKey string, env string, projectId string) (infisical.Secret, error) {
	return sm.client.Secrets().Retrieve(
		infisical.RetrieveSecretOptions{
			SecretKey:   secretKey,
			Environment: env,
			ProjectID:   projectId,
			SecretPath:  "/",
		},
	)
}
