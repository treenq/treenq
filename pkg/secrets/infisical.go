package secrets

import (
	infisical "github.com/infisical/go-sdk"
)

type SecretsManager struct {
	client infisical.InfisicalClientInterface
}

func NewSecretsManager() (*SecretsManager, error) {
	client := infisical.NewInfisicalClient(infisical.Config{
		SiteUrl: "https://app.infisical.com",
	})
	// use INFISICAL_UNIVERSAL_AUTH_CLIENT_ID, INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET env
	_, err := client.Auth().UniversalAuthLogin("", "")
	if err != nil {
		return nil, err
	}

	return &SecretsManager{
		client: client,
	}, nil
}

func (sm *SecretsManager) GetSecret(secretKey string, env string, projectId string) (infisical.Secret, error) {
	return sm.client.Secrets().Retrieve(
		infisical.RetrieveSecretOptions{
			SecretKey:   secretKey,
			Environment: env,
			ProjectID:   projectId,
			SecretPath:  "/",
		},
	)
}
