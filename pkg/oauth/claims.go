package oauth

import (
	"github.com/golang-jwt/jwt/v5"
)

type JwtType string

const (
	JwtTypeAccess  JwtType = "access"
	JwtTypeRefresh JwtType = "refresh"
)

type BaseClaims struct {
	Jti       string `json:"jti"`
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiredAt int64  `json:"exp"`
}

func (s BaseClaims) GetExpirationTime() (*jwt.NumericDate, error) { return nil, nil }

func (s BaseClaims) GetNotBefore() (*jwt.NumericDate, error) { return nil, nil }

func (s BaseClaims) GetIssuedAt() (*jwt.NumericDate, error) { return nil, nil }

func (s BaseClaims) GetAudience() (jwt.ClaimStrings, error) { return nil, nil }

func (s BaseClaims) GetIssuer() (string, error) { return "", nil }

func (s BaseClaims) GetSubject() (string, error) { return "", nil }

type AccessClaims struct {
	BaseClaims
	JwtType JwtType `json:"typ"`
	Version int64   `json:"ver"`
}

type RefreshClaims struct {
	BaseClaims
	JwtType JwtType `json:"typ"`
	Version int64   `json:"ver"`
}
