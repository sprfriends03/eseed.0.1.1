package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/util"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

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
	})
}

// @Tags Auth
// @Summary Login
// @Param body body db.AuthLoginData true "body"
// @Success 200 {object} db.AuthTokenDto
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
// @Summary Register
// @Param body body db.AuthRegisterData true "body"
// @Success 200
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
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		c.Status(http.StatusOK)
	}
}

// @Tags Auth
// @Summary Refresh Token
// @Param body body db.AuthRefreshTokenData true "body"
// @Success 200 {object} db.AuthTokenDto
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
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

// @Tags Auth
// @Summary Change Password
// @Security BearerAuth
// @Param body body db.AuthChangePasswordData true "body"
// @Success 200
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

		c.Status(http.StatusOK)
	}
}

// @Tags Auth
// @Summary Get Me
// @Security BearerAuth
// @Success 200 {object} db.AuthSessionDto
// @Router /auth/v1/me [get]
func (s auth) v1_GetMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		c.JSON(http.StatusOK, session)
	}
}

// @Tags Auth
// @Summary Logout
// @Security BearerAuth
// @Success 200
// @Router /auth/v1/logout [post]
func (s auth) v1_Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		s.oauth.RevokeToken(c.Request.Context(), session.AccessToken)
		c.Status(http.StatusOK)
	}
}

// @Tags Auth
// @Summary Flush Cache
// @Security BearerAuth
// @Success 200
// @Router /auth/v1/flush-cache [get]
func (s auth) v1_FlushCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)
		if !session.IsRoot || session.IsTenant {
			c.Error(ecode.Forbidden)
			return
		}
		s.store.Rdb.FlushAll(c.Request.Context())
		c.Status(http.StatusOK)
	}
}
