package oauth

import (
	"app/env"
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store"
	"app/store/db"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Oauth struct {
	store *store.Store
}

func New(store *store.Store) *Oauth {
	return &Oauth{store}
}

func (s *Oauth) BearerAuth(r *http.Request) (*db.AuthSessionDto, error) {
	var (
		auth   = r.Header.Get("Authorization")
		prefix = "Bearer "
		access = ""
	)

	if auth != "" && strings.HasPrefix(auth, prefix) {
		access = auth[len(prefix):]
	} else {
		access = r.FormValue("access_token")
	}

	if access == "" {
		return nil, ecode.Unauthorized
	}

	return s.ValidateToken(r.Context(), access)
}

func (s *Oauth) BasicAuth(r *http.Request) (*db.AuthSessionDto, error) {
	clientId, clientSecret, ok := r.BasicAuth()
	if !ok {
		return nil, ecode.Unauthorized
	}

	client, err := s.store.GetClient(r.Context(), clientId)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if client.ClientSecret != clientSecret {
		return nil, ecode.Unauthorized
	}

	tenant, err := s.store.GetTenant(r.Context(), string(client.TenantId))
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if tenant.DataStatus == enum.DataStatusDisable {
		return nil, ecode.Forbidden
	}

	session := &db.AuthSessionDto{
		Username: clientId,
		TenantId: client.TenantId,
	}

	return session, nil
}

func (s *Oauth) GenerateToken(ctx context.Context, uid string) (*db.AuthTokenDto, error) {
	client, err := s.store.GetClient(ctx, env.ClientId)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	user, err := s.store.GetUser(ctx, uid)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if user.DataStatus == enum.DataStatusDisable {
		return nil, ecode.Forbidden
	}

	tenant, err := s.store.GetTenant(ctx, string(user.TenantId))
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if tenant.DataStatus == enum.DataStatusDisable {
		return nil, ecode.Forbidden
	}

	now := time.Now()
	key := []byte(client.ClientId + client.ClientSecret + client.SecureKey)
	dto := &db.AuthTokenDto{ExpiresIn: int64(time.Hour.Seconds()), TokenType: "Bearer"}

	access := &AccessClaims{
		BaseClaims: BaseClaims{
			Jti:       uuid.NewString(),
			Subject:   user.ID,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiredAt: now.Add(time.Hour).Unix(),
		},
		JwtType: JwtTypeAccess,
		Version: user.VersionToken,
	}

	dto.AccessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, access).SignedString(key)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	refresh := &RefreshClaims{
		BaseClaims: BaseClaims{
			Jti:       access.Jti,
			Subject:   user.ID,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiredAt: now.Add(time.Hour * 24).Unix(),
		},
		JwtType: JwtTypeRefresh,
		Version: user.VersionToken,
	}

	dto.RefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refresh).SignedString(key)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	return dto, nil
}

func (s *Oauth) RefreshToken(ctx context.Context, refresh string) (*db.AuthTokenDto, error) {
	client, err := s.store.GetClient(ctx, env.ClientId)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	key := []byte(client.ClientId + client.ClientSecret + client.SecureKey)

	claims := &RefreshClaims{}

	if _, err := jwt.ParseWithClaims(refresh, claims, func(token *jwt.Token) (any, error) { return key, nil }); err != nil || claims.JwtType != JwtTypeRefresh {
		return nil, ecode.InvalidToken
	}

	if claims.ExpiredAt < time.Now().Unix() || time.Now().Unix() < claims.NotBefore {
		return nil, ecode.InvalidToken
	}

	user, err := s.store.GetUser(ctx, claims.Subject)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if user.VersionToken != claims.Version {
		return nil, ecode.InvalidToken
	}

	if s.IsRevoked(ctx, claims.Jti) {
		return nil, ecode.InvalidToken
	}

	tk, err := s.GenerateToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	s.store.Rdb.Set(ctx, fmt.Sprintf("jti:%v", claims.Jti), true, time.Hour*24)

	return tk, nil
}

func (s *Oauth) ValidateToken(ctx context.Context, access string) (*db.AuthSessionDto, error) {
	client, err := s.store.GetClient(ctx, env.ClientId)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	key := []byte(client.ClientId + client.ClientSecret + client.SecureKey)

	claims := &AccessClaims{}

	if _, err := jwt.ParseWithClaims(access, claims, func(token *jwt.Token) (any, error) { return key, nil }); err != nil || claims.JwtType != JwtTypeAccess {
		return nil, ecode.Unauthorized
	}

	if claims.ExpiredAt < time.Now().Unix() || time.Now().Unix() < claims.NotBefore {
		return nil, ecode.Unauthorized
	}

	user, err := s.store.GetUser(ctx, claims.Subject)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	if user.VersionToken != claims.Version {
		return nil, ecode.Unauthorized
	}

	if s.IsRevoked(ctx, claims.Jti) {
		return nil, ecode.Unauthorized
	}

	session := &db.AuthSessionDto{
		Name:        user.Name,
		Phone:       user.Phone,
		Email:       user.Email,
		Username:    user.Username,
		UserId:      user.ID,
		TenantId:    user.TenantId,
		Permissions: user.Permissions,
		IsRoot:      user.IsRoot,
		IsTenant:    user.IsTenant,
		AccessToken: access,
	}

	if session.IsRoot {
		session.Permissions = enum.PermissionRootValues()
		if session.IsTenant {
			session.Permissions = enum.PermissionTenantValues()
		}
	}

	return session, nil
}

func (s *Oauth) RevokeTokenByUser(ctx context.Context, uid string) error {
	if err := s.store.Db.User.IncrementVersionToken(ctx, uid); err != nil {
		return ecode.InternalServerError.Desc(err)
	}
	return s.store.DelUser(ctx, uid)
}

func (s *Oauth) RevokeToken(ctx context.Context, access string) error {
	client, err := s.store.GetClient(ctx, env.ClientId)
	if err != nil {
		return ecode.InternalServerError.Desc(err)
	}

	key := []byte(client.ClientId + client.ClientSecret + client.SecureKey)

	claims := &AccessClaims{}

	if _, err := jwt.ParseWithClaims(access, claims, func(token *jwt.Token) (any, error) { return key, nil }); err != nil || claims.JwtType != JwtTypeAccess {
		return ecode.InvalidToken
	}

	return s.store.Rdb.Set(ctx, fmt.Sprintf("jti:%v", claims.Jti), true, time.Hour*24)
}

func (s *Oauth) IsRevoked(ctx context.Context, jti string) bool {
	bytes, _ := s.store.Rdb.GetBytes(ctx, fmt.Sprintf("jti:%v", jti))
	return len(bytes) > 0
}
