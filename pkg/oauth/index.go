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
	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Oauth struct {
	store *store.Store
}

// oauthRefreshMembershipAdapter adapts *db.MembershipDomain to db.MembershipDomainForVerification
// This is defined at the package level to avoid potential linter issues with locally defined methods.
type oauthRefreshMembershipAdapter struct {
	adaptee *db.MembershipDomain
}

func (a *oauthRefreshMembershipAdapter) GetStatus() *string {
	if a.adaptee == nil || a.adaptee.Status == nil {
		return nil
	}
	return a.adaptee.Status // Assumes MembershipDomain.Status is *string
}

func (a *oauthRefreshMembershipAdapter) GetExpirationDate() *time.Time {
	if a.adaptee == nil || a.adaptee.ExpirationDate == nil {
		return nil
	}
	return a.adaptee.ExpirationDate // Assumes MembershipDomain.ExpirationDate is *time.Time
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

// GenerateMemberToken creates access and refresh tokens for a verified member,
// including member-specific claims in the access token.
func (s *Oauth) GenerateMemberToken(ctx context.Context, user *db.UserDomain, member *db.MemberDomain, currentMembershipStatus *string) (*db.AuthTokenDto, error) {
	client, err := s.store.GetClient(ctx, env.ClientId)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	// User status check
	if user.DataStatus == nil || *user.DataStatus == enum.DataStatusDisable {
		return nil, ecode.New(http.StatusForbidden, "user_account_disabled")
	}

	// Tenant ID must exist if user is valid
	if user.TenantId == nil {
		return nil, ecode.New(http.StatusInternalServerError, "user_missing_tenant_id")
	}
	tenant, err := s.store.GetTenant(ctx, string(*user.TenantId))
	if err != nil {
		// If GetTenant itself returns an ecode.Error, pass it directly, otherwise wrap
		if _, ok := err.(*ecode.Error); ok {
			return nil, err
		}
		return nil, ecode.InternalServerError.Desc(err)
	}

	if tenant.DataStatus == enum.DataStatusDisable {
		return nil, ecode.New(http.StatusForbidden, "tenant_disabled")
	}

	now := time.Now()
	key := []byte(client.ClientId + client.ClientSecret + client.SecureKey)
	// Using hardcoded expiry for now, consistent with GenerateToken. TODO: Make configurable via env if needed.
	accessTokenExpiry := time.Hour
	refreshTokenExpiry := time.Hour * 24
	dto := &db.AuthTokenDto{ExpiresIn: int64(accessTokenExpiry.Seconds()), TokenType: "Bearer"}

	// Member Access Token Claims
	memberAccessClaims := &MemberAccessClaims{
		AccessClaims: AccessClaims{
			BaseClaims: BaseClaims{
				Jti:       uuid.NewString(),
				Subject:   db.SID(user.ID),
				IssuedAt:  now.Unix(),
				NotBefore: now.Unix(),
				ExpiredAt: now.Add(accessTokenExpiry).Unix(),
			},
			JwtType: JwtTypeAccess,
			Version: *user.VersionToken, // Assuming VersionToken is not nil and populated
		},
		MembershipID:     member.CurrentMembershipID, // Assumes this field is populated
		MembershipStatus: currentMembershipStatus,    // Passed in, reflects current membership state
		KYCStatus:        member.KYCStatus,           // Assumes this field is populated
	}

	dto.AccessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, memberAccessClaims).SignedString(key)
	if err != nil {
		return nil, ecode.InternalServerError.Desc(err)
	}

	// Refresh Token Claims
	refreshClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			Jti:       memberAccessClaims.AccessClaims.BaseClaims.Jti,
			Subject:   db.SID(user.ID),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiredAt: now.Add(refreshTokenExpiry).Unix(),
		},
		JwtType: JwtTypeRefresh,
		Version: *user.VersionToken,
	}

	dto.RefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(key)
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

	userDomain, err := s.store.Db.User.FindOneById(ctx, claims.Subject)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ecode.UserNotFound.Desc(err)
		}
		return nil, ecode.InternalServerError.Desc(err)
	}

	if userDomain.VersionToken == nil {
		return nil, ecode.InvalidToken.Desc(fmt.Errorf("user token version (pointer) is nil during refresh"))
	}
	if *userDomain.VersionToken != claims.Version {
		return nil, ecode.InvalidToken.Desc(fmt.Errorf("user token version mismatch during refresh"))
	}

	if s.IsRevoked(ctx, claims.Jti) {
		return nil, ecode.InvalidToken.Desc(fmt.Errorf("refresh token JTI already revoked"))
	}

	memberDomain, memberErr := s.store.Db.Member.FindByUserID(ctx, claims.Subject)

	if memberErr == nil && memberDomain != nil {
		var currentMembershipStatusForToken *string

		getMembershipFunc := func(ctx context.Context, membershipID string) (db.MembershipDomainForVerification, error) {
			concreteMembership, findErr := s.store.Db.Membership.FindByID(ctx, membershipID)
			if findErr != nil {
				if findErr == mongo.ErrNoDocuments {
					return nil, ecode.New(http.StatusNotFound, "membership_not_found_for_refresh").Desc(findErr)
				}
				return nil, ecode.InternalServerError.Desc(findErr)
			}
			if concreteMembership == nil {
				return nil, ecode.New(http.StatusNotFound, "membership_record_nil_for_refresh")
			}
			if concreteMembership.Status != nil {
				statusVal := *concreteMembership.Status
				currentMembershipStatusForToken = &statusVal
			}
			return &oauthRefreshMembershipAdapter{adaptee: concreteMembership}, nil // Use package-level adapter
		}

		if errVerify := db.VerifyMemberActive(ctx, memberDomain, getMembershipFunc); errVerify == nil {
			logrus.Infof("Refreshing token for active member: %s", claims.Subject)
			token, errGen := s.GenerateMemberToken(ctx, userDomain, memberDomain, currentMembershipStatusForToken)
			if errGen != nil {
				logrus.Errorf("Error generating member token during refresh for user %s: %v", claims.Subject, errGen)
				return nil, errGen
			}
			s.store.Rdb.Set(ctx, fmt.Sprintf("jti:%v", claims.Jti), true, time.Hour*24)
			return token, nil
		} else {
			logrus.Warnf("Member %s verification failed during token refresh: %v. Issuing standard user token.", claims.Subject, errVerify)
		}
	} else if memberErr != nil && memberErr != mongo.ErrNoDocuments {
		logrus.Errorf("Error fetching member for user %s during refresh token: %v. Issuing standard user token.", claims.Subject, memberErr)
	}

	logrus.Infof("Refreshing standard token for user: %s", claims.Subject)
	tk, err := s.GenerateToken(ctx, db.SID(userDomain.ID))
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
	claims := &MemberAccessClaims{}

	parsedToken, err := jwt.ParseWithClaims(access, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ecode.New(http.StatusUnauthorized, "unexpected_signing_method").Desc(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}
		return key, nil
	})

	if err != nil {
		// Simplified error handling due to persistent linter issues with jwt.ValidationError constants.
		// This will report a more generic error for JWT parsing/validation failures.
		logrus.Warnf("JWT parsing/validation failed: %v. Returning generic invalid token error.", err)
		return nil, ecode.New(http.StatusUnauthorized, "invalid_token") // Generic error
	}

	if !parsedToken.Valid || claims.AccessClaims.JwtType != JwtTypeAccess {
		return nil, ecode.New(http.StatusUnauthorized, "token_not_valid_or_wrong_type")
	}

	if claims.AccessClaims.BaseClaims.ExpiredAt < time.Now().Unix() || time.Now().Unix() < claims.AccessClaims.BaseClaims.NotBefore {
		return nil, ecode.New(http.StatusUnauthorized, "token_expired_or_not_valid_at_current_time")
	}

	if s.IsRevoked(ctx, claims.AccessClaims.BaseClaims.Jti) {
		return nil, ecode.New(http.StatusUnauthorized, "token_revoked_jti")
	}

	user, err := s.store.GetUser(ctx, claims.AccessClaims.BaseClaims.Subject)
	if err != nil {
		if _, ok := err.(*ecode.Error); ok {
			return nil, err
		}
		return nil, ecode.InternalServerError.Desc(err)
	}

	if user.VersionToken != claims.AccessClaims.Version {
		return nil, ecode.New(http.StatusUnauthorized, "token_version_mismatch")
	}

	session := &db.AuthSessionDto{
		UserId:           claims.AccessClaims.BaseClaims.Subject,
		Username:         user.Username,
		TenantId:         user.TenantId,
		Name:             user.Name,
		Phone:            user.Phone,
		Email:            user.Email,
		IsRoot:           user.IsRoot,
		Permissions:      user.Permissions,
		AccessToken:      access,
		MembershipID:     claims.MembershipID,
		MembershipStatus: claims.MembershipStatus,
		KYCStatus:        claims.KYCStatus,
	}

	if claims.MembershipID != nil && *claims.MembershipID != "" {
		session.IsMember = gopkg.Pointer(true)
	} else {
		session.IsMember = gopkg.Pointer(false)
	}

	session.IsTenant = user.IsTenant

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
