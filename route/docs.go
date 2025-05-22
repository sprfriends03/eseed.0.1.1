package route

import (
	_ "app/docs"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		r.GET("/docs/*any", swagger.WrapHandler(swaggerfiles.Handler, swagger.DefaultModelsExpandDepth(-1)))
		r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	})
}
