package route

import (
	"app/pkg/ecode"
	"app/pkg/encryption"
	"app/pkg/enum"
	"app/pkg/util"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/mongo"
)

type client struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := client{m}

		v1cms := r.Group("/cms/v1/clients")
		v1cms.GET("", s.BearerAuth(enum.PermissionClientView), s.v1cms_List())
		v1cms.POST("", s.BearerAuth(enum.PermissionClientCreate), s.v1cms_Create())
		v1cms.DELETE("/:client_id", s.BearerAuth(enum.PermissionClientDelete), s.v1cms_Delete())
	})
}

// @Tags Cms
// @Summary List Clients
// @Security BearerAuth
// @Param query query db.ClientCmsQuery false "query"
// @Success 200 {object} []db.ClientCmsDto
// @Router /cms/v1/clients [get]
func (s client) v1cms_List() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		qb := &db.ClientCmsQuery{}
		if err := c.ShouldBind(qb); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query := qb.BuildCore()
		query.TenantId = gopkg.Pointer(session.TenantId)

		domains, _ := s.store.Db.Client.FindAll(c.Request.Context(), query)
		results := gopkg.MapFunc(domains, func(domain *db.ClientDomain) *db.ClientCmsDto { return domain.CmsDto() })

		s.Pagination(c, s.store.Db.Client.Count(c.Request.Context(), query), query.Page, query.Limit)

		c.JSON(http.StatusOK, results)
	}
}

// @Tags Cms
// @Summary Create Client
// @Security BearerAuth
// @Param body body db.ClientCmsData true "body"
// @Success 200 {object} db.ClientCmsDto
// @Router /cms/v1/clients [post]
func (s client) v1cms_Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		data := &db.ClientCmsData{}
		if err := c.ShouldBind(data); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		retry := 0

	RETRY_CREATE_CLIENT:
		domain := data.Domain(&db.ClientDomain{})
		domain.ClientId = gopkg.Pointer(util.RandomClientId())
		domain.ClientSecret = gopkg.Pointer(encryption.Encrypt(util.RandomClientSecret(), gopkg.Value(domain.ClientId)))
		domain.SecureKey = gopkg.Pointer(encryption.Encrypt(util.RandomSecureKey(), gopkg.Value(domain.ClientId)))
		domain.IsRoot = gopkg.Pointer(false)
		domain.TenantId = gopkg.Pointer(session.TenantId)
		domain.CreatedBy = gopkg.Pointer(session.Username)
		domain.UpdatedBy = gopkg.Pointer(session.Username)

		domain, err := s.store.Db.Client.Save(c.Request.Context(), domain)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) && retry < 5 {
				retry++
				goto RETRY_CREATE_CLIENT
			}
			c.Error(ecode.ClientConflict.Stack(err))
			return
		}

		s.AuditLog(c, s.store.Db.Client.CollectionName(), enum.DataActionCreate, data, domain, db.SID(domain.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}

// @Tags Cms
// @Summary Delete Client
// @Security BearerAuth
// @Param client_id path string true "client_id"
// @Success 200 {object} db.ClientCmsDto
// @Router /cms/v1/clients/{client_id} [delete]
func (s client) v1cms_Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		domain, err := s.store.Db.Client.FindOneByClientId(c.Request.Context(), c.Param("client_id"))
		if err != nil || gopkg.Value(domain.IsRoot) || gopkg.Value(domain.TenantId) != session.TenantId {
			c.Error(ecode.ClientNotFound)
			return
		}

		s.store.Db.Client.DeleteOne(c.Request.Context(), domain)
		s.store.DelClient(c.Request.Context(), gopkg.Value(domain.ClientId))

		s.AuditLog(c, s.store.Db.Client.CollectionName(), enum.DataActionDelete, nil, domain, db.SID(domain.ID))

		c.JSON(http.StatusOK, domain.CmsDto())
	}
}
