package route

import (
	"github.com/gin-gonic/gin"
)

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		r.Static("/assets", "./views/admin/assets")
		r.StaticFile("/favicon.ico", "./views/admin/favicon.ico")

		r.GET("", func(c *gin.Context) { c.File("./views/index.html") })

		v1a := r.Group("/admin")
		v1a.Use(func(c *gin.Context) {
			c.Header("Cross-Origin-Opener-Policy", "same-origin")
			c.Header("Cross-Origin-Resource-Policy", "same-origin")
			c.Header("Cross-Origin-Embedder-Policy", "credentialless")
			c.Next()
		})
		v1a.GET("/*any", func(c *gin.Context) { c.File("./views/admin/index.html") })
	})
}
