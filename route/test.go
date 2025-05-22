package route

import (
	"app/pkg/api"
	"app/store/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type test struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := test{m}

		v1 := r.Group("/test/v1")
		v1.POST("", s.NoAuth(), s.v1_Test())
	})
}

func (s test) v1_Test() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := api.New[db.AuthSessionDto](http.NewRequest(http.MethodGet, "http://localhost:3000/auth/v1/me", c.Request.Body))
		req.SetAuthorization(c.GetHeader("Authorization"))

		res := req.Call()
		if res.Status != http.StatusOK {
			c.Error(res.Error)
			return
		}

		c.JSON(res.Status, res.Data)
	}
}
