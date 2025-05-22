package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type webhook struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := webhook{m}

		v1 := r.Group("/webhook/v1")
		v1.POST("", s.BasicAuth(), s.v1_Webhook())
	})
}

// @Tags Webhook
// @Summary Webhook
// @Security BasicAuth
// @Param body body db.M true "body"
// @Success 200 {object} db.M
// @Router /webhook/v1 [post]
func (s webhook) v1_Webhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.Data(http.StatusOK, c.ContentType(), body)
	}
}
