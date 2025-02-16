package auth

import (
	"context"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/treenq/treenq/pkg/vel"
)

type JwtIssuer struct {
	issuer string
	secret []byte
	public []byte
	ttl    time.Duration
}

func NewJwtIssuer(issuerId string, secretKey []byte, publicKey []byte, ttl time.Duration) *JwtIssuer {
	return &JwtIssuer{
		issuer: issuerId,
		secret: secretKey,
		public: publicKey,
		// secret: _privateKey,
		// public: _publicKey,
		ttl: ttl,
	}
}

func (j *JwtIssuer) GenerateJwtToken(claims map[string]interface{}) (string, error) {
	block, _ := pem.Decode(j.secret)
	if block == nil {
		return "", fmt.Errorf("failed to decode pem")
	}

	now := time.Now()
	jwtClaims := jwt.MapClaims{
		"iat": jwt.NewNumericDate(now),
		"exp": jwt.NewNumericDate(now.Add(j.ttl)),
		"iss": j.issuer,
	}
	for k, v := range claims {
		jwtClaims[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to parse pem: %w", err)
	}
	tokenStr, err := token.SignedString(privateKey)
	return tokenStr, err
}

func (j *JwtIssuer) VerifyToken(tokenString string) (map[string]interface{}, error) {
	block, _ := pem.Decode(j.secret)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(j.public)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pem: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify that the signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Verify issuer
	if iss, ok := claims["iss"].(string); !ok || iss != j.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

func NewJwtMiddleware(jwtIssuer *JwtIssuer, l *slog.Logger) vel.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, (&vel.Error{Code: "UNAUTHORIZED", Message: "authorization header is empty"}).JsonString(), http.StatusUnauthorized)
				return
			}

			token, isBearer := strings.CutPrefix(authHeader, "Bearer ")
			if !isBearer {
				http.Error(w, (&vel.Error{Code: "UNAUTHORIZED", Message: "no bearer token found"}).JsonString(), http.StatusUnauthorized)
				return
			}

			claims, err := jwtIssuer.VerifyToken(token)
			if err != nil {
				http.Error(w, (&vel.Error{Code: "UNAUTHORIZED", Message: "token is invalid", Err: err}).JsonString(), http.StatusForbidden)
				return
			}
			*r = *r.WithContext(ClaimsToCtx(r.Context(), claims))
			h.ServeHTTP(w, r)
		})
	}
}

type claimsCtxType int

var claimsCtxKey claimsCtxType = 1

func ClaimsToCtx(ctx context.Context, claims map[string]interface{}) context.Context {
	return context.WithValue(ctx, claimsCtxKey, claims)
}

func ClaimsFromCtx(ctx context.Context) map[string]interface{} {
	v := ctx.Value(claimsCtxKey)
	if v == nil {
		return nil
	}
	return v.(map[string]interface{})
}
