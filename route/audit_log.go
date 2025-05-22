package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type audit_log struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := audit_log{m}

		v1cms := r.Group("/cms/v1/auditlogs")
		v1cms.GET("", s.BearerAuth(enum.PermissionSystemAuditLog), s.v1cms_List())
	})
}

// @Tags Cms
// @Summary List Audit Logs
// @Security BearerAuth
// @Param query query db.AuditLogCmsQuery false "query"
// @Success 200 {object} []db.AuditLogCmsDto
// @Router /cms/v1/auditlogs [get]
func (s audit_log) v1cms_List() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		qb := &db.AuditLogCmsQuery{}
		if err := c.ShouldBind(qb); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query := qb.BuildCore()
		query.TenantId = gopkg.Pointer(session.TenantId)

		domains, _ := s.store.Db.AuditLog.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.AuditLogDomain) *db.AuditLogCmsDto { return domain.CmsDto() })

		s.Pagination(c, s.store.Db.AuditLog.Count(c.Request.Context(), query), query.Page, query.Limit)

		c.JSON(http.StatusOK, results)
	}
}
