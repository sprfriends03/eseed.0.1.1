package route

import (
	"app/pkg/enum"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type meta struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := meta{m}

		v1 := r.Group("/rest/v1/metas")
		v1.GET("", s.BearerAuth(), s.v1_Meta())
	})
}

// @Tags Rest
// @Summary Get Metas
// @Security BearerAuth
// @Success 200 {object} map[string][]string
// @Router /rest/v1/metas [get]
func (s meta) v1_Meta() gin.HandlerFunc {
	return func(c *gin.Context) {
		tags := enum.Tags()
		tags[string(enum.KindPermission)] = gopkg.MapFunc(s.Permissions(s.Session(c)), func(e enum.Permission) string { return string(e) })
		c.JSON(http.StatusOK, tags)
	}
}
