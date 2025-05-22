package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/mail"
	"app/pkg/util"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type user struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := user{m}

		v1cms := r.Group("/cms/v1/users")
		v1cms.GET("", s.BearerAuth(enum.PermissionUserView), s.v1cms_List())
		v1cms.POST("", s.BearerAuth(enum.PermissionUserCreate), s.v1cms_Create())
		v1cms.PUT("/:user_id", s.BearerAuth(enum.PermissionUserUpdate), s.v1cms_Update())
		v1cms.POST("/:user_id/reset-password", s.BearerAuth(enum.PermissionUserUpdate), s.v1cms_ResetPassword())
		v1cms.GET("/roles", s.BearerAuth(enum.PermissionUserView), s.v1cms_ListRoles())
	})
}

// @Tags Cms
// @Summary List Users
// @Security BearerAuth
// @Param query query db.UserCmsQuery false "query"
// @Success 200 {object} []db.UserCmsDto
// @Router /cms/v1/users [get]
func (s user) v1cms_List() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		qb := &db.UserCmsQuery{}
		if err := c.ShouldBind(qb); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query := qb.BuildCore()
		query.TenantId = gopkg.Pointer(session.TenantId)

		domains, _ := s.store.Db.User.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.UserDomain) *db.UserCmsDto { return domain.CmsDto() })

		s.Pagination(c, s.store.Db.User.Count(c.Request.Context(), query), query.Page, query.Limit)

		c.JSON(http.StatusOK, results)
	}
}

// @Tags Cms
// @Summary Create User
// @Security BearerAuth
// @Param body body db.UserCmsData true "body"
// @Success 200 {object} db.UserCmsDto
// @Router /cms/v1/users [post]
func (s user) v1cms_Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.UserCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		for i := range data.RoleIds {
			role, err := s.store.Db.Role.FindOneById(c.Request.Context(), data.RoleIds[i])
			if err != nil || gopkg.Value(role.TenantId) != session.TenantId {
				c.Error(ecode.RoleNotFound)
				return
			}
		}

		domain := data.Domain(&db.UserDomain{})
		domain.Password = gopkg.Pointer(util.HashPassword(util.RandomPassword()))
		domain.IsRoot = gopkg.Pointer(false)
		domain.TenantId = gopkg.Pointer(session.TenantId)
		domain.CreatedBy = gopkg.Pointer(session.Username)
		domain.UpdatedBy = gopkg.Pointer(session.Username)

		domain, err := s.store.Db.User.Save(c.Request.Context(), domain)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionCreate, data, domain, db.SID(domain.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Update User
// @Security BearerAuth
// @Param user_id path string true "user_id"
// @Param body body db.UserCmsData true "body"
// @Success 200 {object} db.UserCmsDto
// @Router /cms/v1/users/{user_id} [put]
func (s user) v1cms_Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.UserCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		for i := range data.RoleIds {
			role, err := s.store.Db.Role.FindOneById(c.Request.Context(), data.RoleIds[i])
			if err != nil || gopkg.Value(role.TenantId) != session.TenantId {
				c.Error(ecode.RoleNotFound)
				return
			}
		}

		domain, err := s.store.Db.User.FindOneById(c.Request.Context(), c.Param("user_id"))
		if err != nil || gopkg.Value(domain.IsRoot) || gopkg.Value(domain.TenantId) != session.TenantId {
			c.Error(ecode.UserNotFound)
			return
		}

		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: domain.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update = data.Domain(update)

		domain, err = s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		s.store.DelUser(c.Request.Context(), db.SID(domain.ID))
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionUpdate, data, domain, db.SID(domain.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Reset Password For User
// @Security BearerAuth
// @Param user_id path string true "user_id"
// @Success 200 {object} db.UserCmsDto
// @Router /cms/v1/users/{user_id}/reset-password [post]
func (s user) v1cms_ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		domain, err := s.store.Db.User.FindOneById(c.Request.Context(), c.Param("user_id"))
		if err != nil || gopkg.Value(domain.IsRoot) || gopkg.Value(domain.TenantId) != session.TenantId {
			c.Error(ecode.UserNotFound)
			return
		}

		password := util.RandomPassword()

		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: domain.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update.Password = gopkg.Pointer(util.HashPassword(password))

		domain, err = s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		s.oauth.RevokeTokenByUser(c.Request.Context(), db.SID(domain.ID))
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionResetPassword, nil, domain, db.SID(domain.ID))

		if err := s.v1cms_SendPassword(c, domain, password); err != nil {
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

func (s user) v1cms_SendPassword(c *gin.Context, domain *db.UserDomain, password string) error {
	session := s.Session(c)
	tenant, _ := s.store.GetTenant(c.Request.Context(), string(session.TenantId))

	return s.mail.SendPassword(&mail.Password{
		Subject:  c.GetHeader("Origin"),
		Domain:   c.GetHeader("Origin"),
		Username: gopkg.Value(domain.Username),
		Email:    gopkg.Value(domain.Email),
		Password: password,
		Keycode:  tenant.Keycode,
	})
}

// @Tags Cms
// @Summary List Roles
// @Security BearerAuth
// @Success 200 {object} []db.RoleBaseDto
// @Router /cms/v1/users/roles [get]
func (s user) v1cms_ListRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		query := &db.RoleQuery{Query: db.Query{Sorts: "name.asc"}}
		query.TenantId = gopkg.Pointer(session.TenantId)

		domains, _ := s.store.Db.Role.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.RoleDomain) *db.RoleBaseDto { return domain.BaseDto() })

		c.JSON(http.StatusOK, results)
	}
}
