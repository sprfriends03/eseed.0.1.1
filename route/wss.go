package route

import (
	"app/pkg/ecode"
	"app/pkg/ws"

	"github.com/gin-gonic/gin"
)

type wss struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := wss{m}

		v1 := r.Group("/websocket/v1")
		v1.GET("", s.BearerAuth(), s.v1_Connect())
	})
}

// @Tags Websocket
// @Summary Websocket
// @Security BearerAuth
// @Success 101
// @Router /websocket/v1 [get]
func (s *wss) v1_Connect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if conn, err := s.ws.Upgrade(c.Writer, c.Request); err != nil {
			c.Error(ecode.UpgradeRequired)
			return
		} else {
			defer conn.Close()
			ws.NewConn(conn, s.store, s.Session(c))
		}
	}
}
