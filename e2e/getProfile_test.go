package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/client"
)

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	profile, err := apiClient.GetProfile(ctx)
	assert.NoError(t, err)
	assert.Equal(t, client.GetProfileResponse{
		UserInfo: client.UserInfo{
			Email:       "some@email.jo",
			DisplayName: "RegularUser",
		},
	}, profile)
}

func TestGetProfileInvalidToken(t *testing.T) {
	ctx := context.Background()
	_, err := client.NewClient("http://localhost:8000", http.DefaultClient, nil).GetProfile(ctx)
	var e *client.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, e.Code, "UNAUTHORIZED")
}

func TestGetProfileNoToken(t *testing.T) {
	ctx := context.Background()
	_, err := client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{
		"Authorization": "Bearer invalid",
	}).GetProfile(ctx)
	var e *client.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, e.Code, "UNAUTHORIZED")
}
