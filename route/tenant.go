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

type tenant struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := tenant{m}

		v1cms := r.Group("/cms/v1/tenants")
		v1cms.GET("", s.BearerAuth(enum.PermissionTenantView), s.v1cms_List())
		v1cms.POST("", s.BearerAuth(enum.PermissionTenantCreate), s.v1cms_Create())
		v1cms.PUT("/:tenant_id", s.BearerAuth(enum.PermissionTenantUpdate), s.v1cms_Update())
		v1cms.POST("/:tenant_id/reset-password", s.BearerAuth(enum.PermissionTenantUpdate), s.v1cms_ResetPassword())
	})
}

// @Tags Cms
// @Summary List Tenants
// @Security BearerAuth
// @Param query query db.TenantCmsQuery false "query"
// @Success 200 {object} []db.TenantCmsDto
// @Router /cms/v1/tenants [get]
func (s tenant) v1cms_List() gin.HandlerFunc {
	return func(c *gin.Context) {
		qb := &db.TenantCmsQuery{}
		if err := c.ShouldBind(qb); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query := qb.BuildCore()

		domains, _ := s.store.Db.Tenant.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.TenantDomain) *db.TenantCmsDto { return domain.CmsDto() })

		s.Pagination(c, s.store.Db.Tenant.Count(c.Request.Context(), query), query.Page, query.Limit)

		c.JSON(http.StatusOK, results)
	}
}

// @Tags Cms
// @Summary Create Tenant
// @Security BearerAuth
// @Param body body db.TenantCmsData true "body"
// @Success 200 {object} db.TenantCmsDto
// @Router /cms/v1/tenants [post]
func (s tenant) v1cms_Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.TenantCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		domain := data.Domain(&db.TenantDomain{})
		domain.IsRoot = gopkg.Pointer(false)
		domain.CreatedBy = gopkg.Pointer(session.Username)
		domain.UpdatedBy = gopkg.Pointer(session.Username)

		domain, err := s.store.Db.Tenant.Save(c.Request.Context(), domain)
		if err != nil {
			c.Error(ecode.TenantConflict.Stack(err))
			return
		}

		user := &db.UserDomain{}
		user.Name = domain.Name
		user.Username = domain.Username
		user.Password = gopkg.Pointer(util.HashPassword(util.RandomPassword()))
		user.DataStatus = gopkg.Pointer(enum.DataStatusEnable)
		user.RoleIds = gopkg.Pointer([]string{})
		user.IsRoot = gopkg.Pointer(true)
		user.TenantId = gopkg.Pointer(enum.Tenant(db.SID(domain.ID)))
		user.CreatedBy = gopkg.Pointer(session.Username)
		user.UpdatedBy = gopkg.Pointer(session.Username)

		user, err = s.store.Db.User.Save(c.Request.Context(), user)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		s.AuditLog(c, s.store.Db.Tenant.CollectionName(), enum.DataActionCreate, data, domain, db.SID(domain.ID))
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionCreate, data, user, db.SID(user.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Update Tenant
// @Security BearerAuth
// @Param tenant_id path string true "tenant_id"
// @Param body body db.TenantCmsData true "body"
// @Success 200 {object} db.TenantCmsDto
// @Router /cms/v1/tenants/{tenant_id} [put]
func (s tenant) v1cms_Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.TenantCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		domain, err := s.store.Db.Tenant.FindOneById(c.Request.Context(), c.Param("tenant_id"))
		if err != nil || gopkg.Value(domain.IsRoot) {
			c.Error(ecode.TenantNotFound)
			return
		}

		update := &db.TenantDomain{BaseDomain: db.BaseDomain{ID: domain.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update = data.Domain(update)

		if gopkg.Value(update.Username) != gopkg.Value(domain.Username) {
			user, _ := s.store.Db.User.FindOneByTenant_Username(c.Request.Context(), enum.Tenant(db.SID(domain.ID)), gopkg.Value(domain.Username))
			user.Username = update.Username
			user.UpdatedBy = update.UpdatedBy
			user, err = s.store.Db.User.Save(c.Request.Context(), user)
			if err != nil {
				c.Error(ecode.UserConflict.Stack(err))
				return
			}
			s.oauth.RevokeTokenByUser(c.Request.Context(), db.SID(user.ID))
		}

		domain, err = s.store.Db.Tenant.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.TenantConflict.Stack(err))
			return
		}

		s.store.DelTenant(c.Request.Context(), db.SID(domain.ID))
		s.AuditLog(c, s.store.Db.Tenant.CollectionName(), enum.DataActionUpdate, data, domain, db.SID(domain.ID))

		users, _ := s.store.Db.User.FindAllByTenant(c.Request.Context(), enum.Tenant(db.SID(domain.ID)))
		gopkg.LoopFunc(users, func(user *db.UserDomain) { s.store.DelUser(c.Request.Context(), db.SID(user.ID)) })

		clients, _ := s.store.Db.Client.FindAllByTenant(c.Request.Context(), enum.Tenant(db.SID(domain.ID)))
		gopkg.LoopFunc(clients, func(client *db.ClientDomain) { s.store.DelClient(c.Request.Context(), gopkg.Value(client.ClientId)) })

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Reset Password For Tenant
// @Security BearerAuth
// @Param tenant_id path string true "tenant_id"
// @Success 200 {object} db.TenantCmsDto
// @Router /cms/v1/tenants/{tenant_id}/reset-password [post]
func (s tenant) v1cms_ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		domain, err := s.store.Db.Tenant.FindOneById(c.Request.Context(), c.Param("tenant_id"))
		if err != nil || gopkg.Value(domain.IsRoot) {
			c.Error(ecode.TenantNotFound)
			return
		}

		user, err := s.store.Db.User.FindOneByTenant_Username(c.Request.Context(), enum.Tenant(db.SID(domain.ID)), gopkg.Value(domain.Username))
		if err != nil {
			c.Error(ecode.UserNotFound)
			return
		}

		password := util.RandomPassword()

		update := &db.UserDomain{BaseDomain: db.BaseDomain{ID: user.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update.Password = gopkg.Pointer(util.HashPassword(password))

		user, err = s.store.Db.User.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.UserConflict.Stack(err))
			return
		}

		s.oauth.RevokeTokenByUser(c.Request.Context(), db.SID(user.ID))
		s.AuditLog(c, s.store.Db.User.CollectionName(), enum.DataActionResetPassword, nil, user, db.SID(user.ID))

		if err := s.v1cms_SendPassword(c, domain, password); err != nil {
			c.Error(ecode.InternalServerError.Stack(err))
			return
		}

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

func (s tenant) v1cms_SendPassword(c *gin.Context, domain *db.TenantDomain, password string) error {
	return s.mail.SendPassword(&mail.Password{
		Subject:  c.GetHeader("Origin"),
		Domain:   c.GetHeader("Origin"),
		Username: gopkg.Value(domain.Username),
		Email:    gopkg.Value(domain.Email),
		Password: password,
		Keycode:  gopkg.Value(domain.Keycode),
	})
}
