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

// MemberAccessClaims defines the claims specific to a member's access token.
// It includes standard access claims along with member-specific details.
type MemberAccessClaims struct {
	AccessClaims             // Embeds BaseClaims, JwtType, Version
	MembershipID     *string `json:"mid,omitempty"` // Current active Membership ID
	MembershipStatus *string `json:"mst,omitempty"` // Status of the current membership (e.g., active, expired)
	KYCStatus        *string `json:"kys,omitempty"` // Member's KYC status (e.g., verified, pending)
	// TODO: Consider adding MemberRole or specific permissions if distinct from general user roles
}
