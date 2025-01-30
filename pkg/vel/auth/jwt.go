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

var _privateKey = []byte(`-----BEGIN PRIVATE KEY-----
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
wXzDvRRc72l1qEfDdJg1nd3dSz8WqWg=
-----END PRIVATE KEY-----
`)

var _publicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA0zYsZSxQ0Uiekj/Gam3z
dtTjuskarpl7W0VU04efRtPRC2HAILDzyeRX4X2roUvmURHR3vBcHVEDy3UDgJOs
ojL3GOuS6AOIOcWo1YqdEE4ehL5KPNoiIbqG04S6LBF8k9+Cz4RCu+GdpyREna8X
sIPHyhlBwLRGyCadFjyUSDoaTzsuxFpH9HnwPnJvf2LoJSDWa5pUdbPez4med3Qo
XGL0aPfHQO+vpZCL8VvxDq1AnvkyeTfph3m14E3/IGLadcIpxJorek92Jf9Wxj7G
SyKfK/HLtIu3vk6JAYxldbSf7/1TIQa8YeY+PsgQIyOJETVd+xLIxIHj+1AbzFM9
gCJ5CiqDAcN3JJbzOtWjYTL63/ZZOBUrUhaq5RSGxPKcDaCI9hjKF3UkCnzW/S1a
9hruyC+aZDDIVJA4mX4wTiy3A1d4/4+7MFGiZaqpcvNJOV9ymzj/j7xG5pRWwgrt
YeaJB4cyZc3l5+9AS4MLjIHF3Zufg/QlAPaGmFKyUp8QPUpqTz/DVoX28Bw1+1j8
TNlKqB2BykuwKC0luyb3vzMO1l2YlpcQYRmWUmFpNVfB6XQI46kzGVAeeoFwe/m1
mAhEYVShPItGekkMbj9uo0Jl2njfqP6aCfNvka0kjRFR0KEEX3UeXkVt2RIEQvkn
6qqcc5Eezqr1e38X6gBmknECAwEAAQ==
-----END PUBLIC KEY-----
`)

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
				http.Error(w, "authorization header is empty", http.StatusUnauthorized)
				return
			}

			token, isBearer := strings.CutPrefix(authHeader, "Bearer ")
			if !isBearer {
				http.Error(w, "authorization token is not bearer kind", http.StatusUnauthorized)
				return
			}

			claims, err := jwtIssuer.VerifyToken(token)
			if err != nil {
				http.Error(w, "token is invalid", http.StatusForbidden)
			}
			*r = *r.WithContext(ClaimsToCtx(r.Context(), claims))
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
