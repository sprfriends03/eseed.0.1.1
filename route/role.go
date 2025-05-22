package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
	"golang.org/x/exp/slices"
)

type role struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := role{m}

		v1cms := r.Group("/cms/v1/roles")
		v1cms.GET("", s.BearerAuth(enum.PermissionRoleView), s.v1cms_List())
		v1cms.POST("", s.BearerAuth(enum.PermissionRoleCreate), s.v1cms_Create())
		v1cms.PUT("/:role_id", s.BearerAuth(enum.PermissionRoleUpdate), s.v1cms_Update())
	})
}

// @Tags Cms
// @Summary List Roles
// @Security BearerAuth
// @Param query query db.RoleCmsQuery false "query"
// @Success 200 {object} []db.RoleCmsDto
// @Router /cms/v1/roles [get]
func (s role) v1cms_List() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		qb := &db.RoleCmsQuery{}
		if err := c.ShouldBind(qb); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query := qb.BuildCore()
		query.TenantId = gopkg.Pointer(session.TenantId)

		domains, _ := s.store.Db.Role.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.RoleDomain) *db.RoleCmsDto { return domain.CmsDto() })

		s.Pagination(c, s.store.Db.Role.Count(c.Request.Context(), query), query.Page, query.Limit)

		c.JSON(http.StatusOK, results)
	}
}

// @Tags Cms
// @Summary Create Role
// @Security BearerAuth
// @Param body body db.RoleCmsData true "body"
// @Success 200 {object} db.RoleCmsDto
// @Router /cms/v1/roles [post]
func (s role) v1cms_Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.RoleCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		if slices.ContainsFunc(data.Permissions, func(e enum.Permission) bool { return !slices.Contains(s.Permissions(session), e) }) {
			c.Error(ecode.InvalidPermission)
			return
		}

		domain := data.Domain(&db.RoleDomain{})
		domain.TenantId = gopkg.Pointer(session.TenantId)
		domain.CreatedBy = gopkg.Pointer(session.Username)
		domain.UpdatedBy = gopkg.Pointer(session.Username)

		domain, err := s.store.Db.Role.Save(c.Request.Context(), domain)
		if err != nil {
			c.Error(ecode.RoleConflict.Stack(err))
			return
		}

		s.AuditLog(c, s.store.Db.Role.CollectionName(), enum.DataActionCreate, data, domain, db.SID(domain.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Update Role
// @Security BearerAuth
// @Param role_id path string true "role_id"
// @Param body body db.RoleCmsData true "body"
// @Success 200 {object} db.RoleCmsDto
// @Router /cms/v1/roles/{role_id} [put]
func (s role) v1cms_Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.RoleCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		if slices.ContainsFunc(data.Permissions, func(e enum.Permission) bool { return !slices.Contains(s.Permissions(session), e) }) {
			c.Error(ecode.InvalidPermission)
			return
		}

		domain, err := s.store.Db.Role.FindOneById(c.Request.Context(), c.Param("role_id"))
		if err != nil || gopkg.Value(domain.TenantId) != session.TenantId {
			c.Error(ecode.RoleNotFound)
			return
		}

		update := &db.RoleDomain{BaseDomain: db.BaseDomain{ID: domain.ID, UpdatedBy: gopkg.Pointer(session.Username)}}
		update = data.Domain(update)

		domain, err = s.store.Db.Role.Save(c.Request.Context(), update)
		if err != nil {
			c.Error(ecode.RoleConflict.Stack(err))
			return
		}

		s.AuditLog(c, s.store.Db.Role.CollectionName(), enum.DataActionUpdate, data, domain, db.SID(domain.ID))

		users, _ := s.store.Db.User.FindAllByRole(c.Request.Context(), db.SID(domain.ID))
		gopkg.LoopFunc(users, func(user *db.UserDomain) { s.store.DelUser(c.Request.Context(), db.SID(user.ID)) })

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}
