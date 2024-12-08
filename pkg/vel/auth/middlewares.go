package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type Config struct {
	ID       string
	Secret   string
	KeyID    string
	Endpoint string
}

type Profile struct {
	Name     string
	Username string
	Email    string
}

type Context struct {
	mw *Interceptor[*oauth.IntrospectionContext]
}

func (c Context) GetProfile(ctx context.Context) Profile {
	introspectionResponse := c.mw.Context(ctx)
	introspectionResponse.Email = "test@test.com"
	return Profile{
		Name:     introspectionResponse.Name,
		Username: introspectionResponse.Username,
		Email:    introspectionResponse.Email,
	}
}

func NewAuthMiddleware(ctx context.Context, conf Config, l *slog.Logger) (func(http.Handler) http.Handler, *Context, error) {
	keyFile := &client.KeyFile{
		Type:     "application",
		KeyID:    conf.KeyID,
		ClientID: conf.ID,
		Key:      conf.Secret,
	}
	verifier := oauth.WithIntrospection[*oauth.IntrospectionContext](oauth.JWTProfileIntrospectionAuthentication(keyFile))
	auth, err := authorization.New(ctx, zitadel.New(conf.Endpoint), verifier)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build auth middleware: %s: %w", conf.Secret, err)
	}

	mw := NewInterceptor(auth, l)
	return mw.RequireAuthorization(), &Context{mw}, nil
}
