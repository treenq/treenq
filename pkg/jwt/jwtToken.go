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

func NewJwtIssuer(issuerId string,secretKey string,ttl time.Duration) (*JwtIssuer) {
	return &JwtIssuer{
		issuer: issuerId,
		secret:secretKey,
		ttl: ttl,
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
	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ttl)),
		Issuer:    j.issuer,
	})
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return "", fmt.Errorf("failed to parse pem: %w", err)
	}
	myJwt, err := token.SignedString(privateKey)
	return myJwt, err
}
