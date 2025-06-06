package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/util"
	"app/store/db"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// membershipAdapter is a package-level adapter to satisfy db.MembershipDomainForVerification.
// It wraps the concrete *db.MembershipDomain type.
type membershipAdapter struct {
	adaptee *db.MembershipDomain // Assumed type from s.store.Db.Membership.FindByID
}

// GetStatus implements a method for db.MembershipDomainForVerification.
func (a *membershipAdapter) GetStatus() *string {
	if a.adaptee == nil || a.adaptee.Status == nil {
		return nil
	}
	return a.adaptee.Status // Assumes *db.MembershipDomain has a public 'Status' field
}

// GetExpirationDate implements a method for db.MembershipDomainForVerification.
func (a *membershipAdapter) GetExpirationDate() *time.Time {
	if a.adaptee == nil || a.adaptee.ExpirationDate == nil {
		return nil
	}
	return a.adaptee.ExpirationDate // Assumes *db.MembershipDomain has a public 'ExpirationDate' field
}

type auth struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := auth{m}

		v1 := r.Group("/auth/v1")
		v1.POST("/login", s.NoAuth(), s.v1_Login())
		v1.POST("/register", s.NoAuth(), s.v1_Register())
		v1.POST("/refresh-token", s.NoAuth(), s.v1_RefreshToken())
		v1.POST("/logout", s.BearerAuth(), s.v1_Logout())
		v1.POST("/change-password", s.BearerAuth(), s.v1_ChangePassword())
		v1.GET("/me", s.BearerAuth(), s.v1_GetMe())
		v1.GET("/flush-cache", s.BearerAuth(), s.v1_FlushCache())

		// Member-specific routes
		v1.POST("/member/register", s.NoAuth(), s.v1_MemberRegister())
		v1.POST("/member/login", s.NoAuth(), s.v1_MemberLogin())
		v1.GET("/member/verify-email", s.NoAuth(), s.v1_VerifyMemberEmail())
	})
}

// @Tags Auth
// @Summary Login
// @Description Authenticates a standard user and returns JWT tokens.
// @Param body body db.AuthLoginData true "User login credentials"
// @Success 200 {object} db.AuthTokenDto "Successfully logged in"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid credentials"
// @Failure 404 {object} ecode.Error "Not Found - Tenant or User not found"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/login [post]
func (s auth) v1_Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := &db.AuthLoginData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		tenant, err := s.store.Db.Tenant.FindOneByKeycode(c.Request.Context(), data.Keycode)
		if err != nil {
			c.Error(ecode.TenantNotFound)
			return
		}

		user, err := s.store.Db.User.FindOneByTenant_Username(c.Request.Context(), enum.Tenant(db.SID(tenant.ID)), data.Username)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		if !util.VerifyPassword(data.Password, gopkg.Value(user.Password)) {
			c.Error(ecode.UserOrPasswordIncorrect)
			return
		}

		token, err := s.oauth.GenerateToken(c.Request.Context(), db.SID(user.ID))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

// @Tags Auth
// @Summary Register a new standard User
// @Description Creates a new standard user profile.
// @Param body body db.AuthRegisterData true "User registration details"
// @Success 200 {object} object "{message: string}" "Successfully registered"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 404 {object} ecode.Error "Tenant Not Found"
// @Failure 409 {object} ecode.Error "Conflict - Username already exists"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/register [post]
func (s auth) v1_Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := &db.AuthRegisterData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		tenant, err := s.store.Db.Tenant.FindOneByKeycode(c.Request.Context(), data.Keycode)
		if err != nil {
			c.Error(ecode.TenantNotFound)
			return
		}

		domain := &db.UserDomain{}
		domain.Username = gopkg.Pointer(data.Username)
		domain.DataStatus = gopkg.Pointer(enum.DataStatusEnable)
		domain.Password = gopkg.Pointer(util.HashPassword(data.Password))
		domain.RoleIds = gopkg.Pointer([]string{})
		domain.IsRoot = gopkg.Pointer(false)
		domain.TenantId = gopkg.Pointer(enum.Tenant(db.SID(tenant.ID)))

		if _, err = s.store.Db.User.Save(c.Request.Context(), domain); err != nil {
			c.Error(ecode.UserConflict.Stack(err)) // Assumes UserConflict is the correct ecode
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}

// @Tags Auth
// @Summary Refresh Token (Member-Aware)
// @Description Refreshes JWT tokens. If the user is a member and active, member-specific claims are included in the new access token. Otherwise, a standard user token is issued.
// @Param body body db.AuthRefreshTokenData true "Refresh token data (including keycode, username, refresh_token)"
// @Success 200 {object} db.AuthTokenDto "Successfully refreshed tokens"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid refresh token, user token version mismatch, JTI revoked, or (if member refresh attempted) member inactive/invalid"
// @Failure 404 {object} ecode.Error "Not Found - Tenant or User not found"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/refresh-token [post]
func (s auth) v1_RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := &db.AuthRefreshTokenData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		tenant, err := s.store.Db.Tenant.FindOneByKeycode(c.Request.Context(), data.Keycode)
		if err != nil {
			c.Error(ecode.TenantNotFound)
			return
		}

		if _, err := s.store.Db.User.FindOneByTenant_Username(c.Request.Context(), enum.Tenant(db.SID(tenant.ID)), data.Username); err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		token, err := s.oauth.RefreshToken(c.Request.Context(), data.RefreshToken)
		if err != nil {
			c.Error(err) // RefreshToken method itself returns appropriate ecode.Error
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

// @Tags Auth
// @Summary Change Password
// @Description Allows an authenticated user to change their password.
// @Security BearerAuth
// @Param body body db.AuthChangePasswordData true "Old and new password details"
// @Success 200 {object} object "{message: string}" "Password changed successfully"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 401 {object} ecode.Error "Unauthorized - Old password incorrect"
// @Failure 404 {object} ecode.Error "Not Found - User not found"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/change-password [post]
func (s auth) v1_ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.AuthChangePasswordData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		domain, err := s.store.Db.User.FindOneById(c.Request.Context(), session.UserId)
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		if !util.VerifyPassword(data.OldPassword, gopkg.Value(domain.Password)) {
			c.Error(ecode.OldPasswordIncorrect)
			return
		}

		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: domain.ID}}
		update.Password = gopkg.Pointer(util.HashPassword(data.NewPassword))

		s.store.Db.User.Save(c.Request.Context(), update)
		s.oauth.RevokeTokenByUser(c.Request.Context(), db.SID(domain.ID))

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

// @Tags Auth
// @Summary Get Current User/Member Information
// @Description Retrieves session information for the authenticated user. Includes member-specific details if the user is a member.
// @Security BearerAuth
// @Success 200 {object} db.AuthSessionDto "User and Member session information"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Router /auth/v1/me [get]
func (s auth) v1_GetMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		c.JSON(http.StatusOK, session)
	}
}

// @Tags Auth
// @Summary Logout User/Member
// @Description Invalidates the current user/member's access token.
// @Security BearerAuth
// @Success 200 {object} object "{message: string}" "Successfully logged out"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 500 {object} ecode.Error "Internal Server Error during token revocation"
// @Router /auth/v1/logout [post]
func (s auth) v1_Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		err := s.oauth.RevokeToken(c.Request.Context(), session.AccessToken)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

// @Tags Auth
// @Summary Flush Cache (Admin)
// @Description Flushes all keys in the Redis cache. Requires root admin privileges.
// @Security BearerAuth
// @Success 200 {object} object "{message: string}" "Cache flushed successfully"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid or expired token"
// @Failure 403 {object} ecode.Error "Forbidden - Insufficient privileges"
// @Router /auth/v1/flush-cache [get]
func (s auth) v1_FlushCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		if !session.IsRoot || session.IsTenant {
			c.Error(ecode.Forbidden)
			return
		}
		s.store.Rdb.FlushAll(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"message": "Cache flushed successfully"})
	}
}

// @Tags Member Auth
// @Summary Register a new Member
// @Description Creates a new User and an associated Member profile. Sends a verification email. DateOfBirth format: YYYY-MM-DD.
// @Param body body db.MemberRegisterData true "Member registration details"
// @Success 201 {object} object{message=string} "Successfully registered. Please check your email to verify your account."
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data (e.g., missing fields, invalid email, invalid DateOfBirth format 'YYYY-MM-DD', or other validation issues)"
// @Failure 404 {object} ecode.Error "Not Found - Tenant not found (e.g., invalid keycode)"
// @Failure 409 {object} ecode.Error "Conflict - Username or Email already exists OR Member data conflict (e.g., unique constraint on member's email if different from user's)"
// @Failure 500 {object} ecode.Error "Internal Server Error - Could be DB operation failure or email sending failure"
// @Router /auth/v1/member/register [post]
func (s auth) v1_MemberRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := &db.MemberRegisterData{}
		if err := c.ShouldBindJSON(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		ctx := c.Request.Context()

		// Parse date of birth from string
		dob, err := time.Parse("2006-01-02", data.DateOfBirth)
		if err != nil {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_date_format").Desc(fmt.Errorf("DateOfBirth must be in YYYY-MM-DD format")))
			return
		}

		// Age verification - check if user is 18 years or older
		today := time.Now()
		age := today.Year() - dob.Year()

		// If birthday hasn't occurred yet this year, subtract 1
		if today.Month() < dob.Month() || (today.Month() == dob.Month() && today.Day() < dob.Day()) {
			age--
		}

		if age < 18 {
			c.Error(ecode.New(http.StatusForbidden, "age_verification_failed").Desc(fmt.Errorf("Members must be 18 years or older to register")))
			return
		}

		tenant, err := s.store.Db.Tenant.FindOneByKeycode(ctx, data.Keycode)
		if err != nil {
			c.Error(ecode.TenantNotFound.Desc(err))
			return
		}
		tenantID := enum.Tenant(db.SID(tenant.ID))

		userDomain := &db.UserDomain{
			Username:      gopkg.Pointer(data.Username),
			Password:      gopkg.Pointer(util.HashPassword(data.Password)),
			Email:         gopkg.Pointer(data.Email),
			TenantId:      gopkg.Pointer(tenantID),
			DataStatus:    gopkg.Pointer(enum.DataStatusEnable),
			IsRoot:        gopkg.Pointer(false),
			EmailVerified: gopkg.Pointer(false),
			DateOfBirth:   &dob, // Set the date of birth
		}

		savedUser, err := s.store.Db.User.Save(ctx, userDomain)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// More specific message could be "Username or email already exists."
				c.Error(ecode.UserConflict.Desc(err))
			} else {
				c.Error(ecode.InternalServerError.Desc(err))
			}
			return
		}

		verificationToken := uuid.NewString()
		tokenExpiresAt := time.Now().Add(24 * time.Hour)

		savedUser.EmailVerificationToken = gopkg.Pointer(verificationToken)
		savedUser.EmailVerificationTokenExpiresAt = gopkg.Pointer(tokenExpiresAt)
		savedUser.EmailVerified = gopkg.Pointer(false)

		if _, err = s.store.Db.User.Save(ctx, savedUser); err != nil { // Save again to update token fields
			logrus.Errorf("Failed to save user with email verification token: %v", err)
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		memberDomain := &db.MemberDomain{
			UserID:       gopkg.Pointer(db.SID(savedUser.ID)),
			Email:        gopkg.Pointer(data.Email), // Consider if this should be distinct from User.Email or always synced
			FirstName:    gopkg.Pointer(data.FirstName),
			LastName:     gopkg.Pointer(data.LastName),
			DateOfBirth:  gopkg.Pointer(dob),
			Phone:        gopkg.Pointer(data.Phone),
			TenantId:     gopkg.Pointer(tenantID),
			JoinDate:     gopkg.Pointer(time.Now()),
			MemberStatus: gopkg.Pointer("pending_verification"), // Default status
			KYCStatus:    gopkg.Pointer("pending_kyc"),          // Default status
		}

		if _, err := s.store.Db.Member.Save(ctx, memberDomain); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// This might indicate a unique index on Member.Email or Member.UserID if User save succeeded
				c.Error(ecode.New(http.StatusConflict, "member_data_conflict").Desc(err))
			} else {
				c.Error(ecode.InternalServerError.Desc(err))
			}
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" { // Fallback if Origin header is not present
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
			origin = scheme + "://" + c.Request.Host
		}
		verificationLinkBase := origin + "/auth/v1/member/verify-email?token="

		if err := s.mail.SendMemberVerificationEmail(data.Email, data.Username, verificationToken, verificationLinkBase); err != nil {
			logrus.Errorf("Failed to send verification email to %s for user %s: %v", data.Email, data.Username, err)
			// Do not return error to client for email failure, but log it. Registration is technically successful.
		}

		logrus.Infof("Member registration: User %s created. Member profile created. Email verification token for %s: %s. Verification email initiated.", data.Username, data.Email, verificationToken)

		c.JSON(http.StatusCreated, gin.H{"message": "Member registered successfully. Please check your email to verify your account."})
	}
}

// @Tags Member Auth
// @Summary Login for a Member
// @Description Authenticates a user who is a member, checks their member-specific statuses (KYC, age, membership), and returns JWT tokens if all checks pass.
// @Description //TODO: Consider if email must be verified for member login (currently not checked).
// @Param body body db.AuthLoginData true "Member login credentials (standard username/password and keycode)"
// @Success 200 {object} db.AuthTokenDto "Successfully logged in, returns JWT tokens with member claims"
// @Failure 400 {object} ecode.Error "Bad Request - Invalid input data"
// @Failure 401 {object} ecode.Error "Unauthorized - Invalid credentials OR User account disabled OR Member status check failed (e.g., KYC_VERIFICATION_REQUIRED, MEMBER_TOO_YOUNG, MEMBERSHIP_EXPIRED, MEMBERSHIP_STATUS_INVALID)"
// @Failure 404 {object} ecode.Error "Not Found - Tenant not found OR User not found OR Member profile not associated with the user"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/member/login [post]
func (s auth) v1_MemberLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		data := &db.AuthLoginData{}
		if err := c.ShouldBindJSON(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		ctx := c.Request.Context()

		tenant, err := s.store.Db.Tenant.FindOneByKeycode(ctx, data.Keycode)
		if err != nil {
			c.Error(ecode.TenantNotFound.Desc(err))
			return
		}

		user, err := s.store.Db.User.FindOneByTenant_Username(ctx, enum.Tenant(db.SID(tenant.ID)), data.Username)
		if err != nil {
			c.Error(ecode.UserNotFound.Desc(err))
			return
		}

		// TODO: Consider if user.EmailVerified should be checked here for member login.
		// if user.EmailVerified == nil || !*user.EmailVerified {
		// 	c.Error(ecode.New(http.StatusForbidden, "email_not_verified").DescMessage("Please verify your email address before logging in."))
		// 	return
		// }

		if !util.VerifyPassword(data.Password, gopkg.Value(user.Password)) {
			c.Error(ecode.UserOrPasswordIncorrect)
			return
		}

		member, err := s.store.Db.Member.FindByUserID(ctx, db.SID(user.ID))
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.Error(ecode.New(http.StatusNotFound, "member_profile_not_found_for_user").Desc(fmt.Errorf("No member profile found for this user account.")))
			} else {
				c.Error(ecode.InternalServerError.Desc(err))
			}
			return
		}

		var currentMembershipStatusForToken *string
		getMembershipFunc := func(cbCtx context.Context, membershipID string) (db.MembershipDomainForVerification, error) {
			if s.store.Db.Membership == nil {
				logrus.Error("Membership store (s.store.Db.Membership) is not initialized in callback.")
				return nil, ecode.New(http.StatusInternalServerError, "membership_system_unavailable_in_callback")
			}

			actualConcreteMembership, findErr := s.store.Db.Membership.FindByID(cbCtx, membershipID)
			if findErr != nil {
				if _, ok := findErr.(*ecode.Error); ok { // If it's already an ecode.Error
					return nil, findErr
				} else if findErr == mongo.ErrNoDocuments {
					return nil, ecode.New(http.StatusNotFound, "membership_not_found_in_callback").Desc(findErr)
				}
				return nil, ecode.InternalServerError.Desc(findErr)
			}
			if actualConcreteMembership == nil { // Should be caught by ErrNoDocuments, but as a safeguard
				return nil, ecode.New(http.StatusNotFound, "membership_record_nil_from_store_in_callback")
			}

			// Populate currentMembershipStatusForToken for JWT
			if actualConcreteMembership.Status != nil {
				statusValue := *actualConcreteMembership.Status
				currentMembershipStatusForToken = &statusValue
			}

			adapter := &membershipAdapter{adaptee: actualConcreteMembership}
			return adapter, nil
		}

		// This call will return specific ecodes like KYCVerificationRequired, MemberTooYoung, etc.
		if err := db.VerifyMemberActive(ctx, member, getMembershipFunc); err != nil {
			c.Error(err)
			return
		}

		if user.ID.IsZero() || user.VersionToken == nil { // Should not happen if user is fetched
			logrus.Errorf("User %s is missing ID or VersionToken before member token generation.", db.SID(user.ID))
			c.Error(ecode.New(http.StatusInternalServerError, "user_data_incomplete_for_token"))
			return
		}

		// GenerateMemberToken also checks user.DataStatus and tenant.DataStatus
		token, err := s.oauth.GenerateMemberToken(ctx, user, member, currentMembershipStatusForToken)
		if err != nil {
			if ecodeErr, ok := err.(*ecode.Error); ok { // If GenerateMemberToken returns an ecode.Error
				c.Error(ecodeErr)
			} else {
				c.Error(ecode.InternalServerError.Desc(err))
			}
			return
		}
		logrus.Infof("Member login successful for user %s, member %s", db.SID(user.ID), db.SID(member.ID))
		c.JSON(http.StatusOK, token)
	}
}

// @Tags Member Auth
// @Summary Verify Member's Email Address
// @Description Verifies a member's email address using a token sent via email upon registration.
// @Param token query string true "Email verification token"
// @Success 200 {object} object{message=string} "Email verified successfully OR Email already verified."
// @Failure 400 {object} ecode.Error "Bad Request - Token missing (e.g., 'missing_verification_token')"
// @Failure 404 {object} ecode.Error "Not Found - Token not found or already used (e.g., 'invalid_or_used_token')"
// @Failure 410 {object} ecode.Error "Gone - Token expired (e.g., 'expired_verification_token')"
// @Failure 500 {object} ecode.Error "Internal Server Error"
// @Router /auth/v1/member/verify-email [get]
func (s auth) v1_VerifyMemberEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.Error(ecode.New(http.StatusBadRequest, "missing_verification_token"))
			return
		}

		ctx := c.Request.Context()
		user, err := s.store.Db.User.FindOneByEmailVerificationToken(ctx, token)

		if err != nil {
			// FindOneByEmailVerificationToken might return a specific ecode.Error or mongo.ErrNoDocuments
			if ecodeErr, ok := err.(*ecode.Error); ok && ecodeErr.ErrCode == "empty_verification_token" { // Check against string code
				c.Error(err) // Propagate specific error
			} else if err == mongo.ErrNoDocuments {
				c.Error(ecode.New(http.StatusNotFound, "invalid_or_used_token"))
			} else {
				c.Error(ecode.InternalServerError.Desc(err))
			}
			return
		}

		if user.EmailVerified != nil && *user.EmailVerified {
			logrus.Infof("Email verification attempted for already verified user %s (token %s)", db.SID(user.ID), token)
			c.JSON(http.StatusOK, gin.H{"message": "Email already verified."})
			return
		}

		if user.EmailVerificationTokenExpiresAt == nil || time.Now().After(*user.EmailVerificationTokenExpiresAt) {
			logrus.Warnf("Expired email verification token used for user %s (token %s)", db.SID(user.ID), token)
			c.Error(ecode.New(http.StatusGone, "expired_verification_token"))
			return
		}

		now := time.Now()
		user.EmailVerified = gopkg.Pointer(true)
		user.EmailVerificationToken = nil          // Clear token after use
		user.EmailVerificationTokenExpiresAt = nil // Clear expiry after use
		user.EmailVerifiedAt = &now                // Set verification timestamp

		updatedUser, err := s.store.Db.User.Save(ctx, user) // Save changes
		if err != nil {
			logrus.Errorf("Failed to save user after email verification for UserID %s: %v", db.SID(user.ID), err)
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		logrus.Infof("Email successfully verified for user %s (token %s)", db.SID(updatedUser.ID), token)
		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully."})
	}
}
