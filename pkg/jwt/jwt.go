package jwt

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtIssuer struct {
	issuer string
	secret string
	ttl    time.Duration
}

func NewJwtIssuer(issuerId string, secretKey string, ttl time.Duration) *JwtIssuer {
	return &JwtIssuer{
		issuer: issuerId,
		secret: secretKey,
		ttl:    ttl,
	}
}
func (j *JwtIssuer) GeneratedJwtToken() (string, error) {
	pemData, err := base64.RawStdEncoding.DecodeString(j.secret)
	if len(pemData) == 0 {
		return "", fmt.Errorf("failed to decode pem: %w", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return "", fmt.Errorf("failed to decode pem")
	}

	now := time.Now()
	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		Issuer:    j.issuer,
	})
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return "", fmt.Errorf("failed to parse pem: %w", err)
	}
	token, err := claims.SignedString(privateKey)
	return token, err
}
