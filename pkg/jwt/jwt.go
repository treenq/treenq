package jwt

import (
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	issuer string
	secret []byte
	ttl    time.Duration
}

func NewIssuer(issuerId string, secretKey []byte, ttl time.Duration) *Issuer {
	return &Issuer{
		issuer: issuerId,
		secret: secretKey,
		ttl:    ttl,
	}
}
func (j *Issuer) GeneratedJwtToken() (string, error) {
	block, _ := pem.Decode(j.secret)
	if block == nil {
		return "", fmt.Errorf("failed to decode pem")
	}

	now := time.Now()
	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		Issuer:    j.issuer,
	})
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to parse pem: %w", err)
	}
	token, err := claims.SignedString(privateKey)
	return token, err
}
