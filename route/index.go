package route

import (
	"app/env"
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/mail"
	"app/pkg/oauth"
	"app/pkg/trace"
	"app/pkg/validate"
	"app/pkg/ws"
	"app/store"
	"app/store/db"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis_rate/v10"
	"github.com/nhnghia272/gopkg"
)

var (
	handlers     = make([]handler, 0)
	defaultLimit = redis_rate.PerMinute(100)                                 // Default: 100 requests per minute
	authLimit    = redis_rate.Limit{Rate: 10, Burst: 3, Period: time.Minute} // Auth routes: 10 req/min, burst 3
	// Define a new stricter limit for member login
	memberLoginRateLimit = redis_rate.Limit{Rate: 5, Burst: 2, Period: time.Minute} // Member login: 5 req/min, burst 2
)

type handler = func(*middleware, *gin.Engine)

func Bootstrap(store *store.Store) error {
	mdw := newMdw(store)
	gin.SetMode(gin.ReleaseMode)
	binding.Validator = validate.New()

	app := gin.New()
	app.NoRoute(mdw.NoRoute())
	app.Use(mdw.Cors(), mdw.Compress(), mdw.SecureHeaders(), mdw.Trace(), mdw.Logger(), mdw.Recover(), mdw.Error())

	for i := range handlers {
		handlers[i](mdw, app)
	}

	fmt.Println("Version: v1.0.0")
	return app.Run(":" + env.Port)
}

type middleware struct {
	store   *store.Store
	oauth   *oauth.Oauth
	mail    *mail.Mail
	ws      *ws.Ws
	limiter *redis_rate.Limiter
}

func newMdw(store *store.Store) *middleware {
	return &middleware{store, oauth.New(store), mail.New(store), ws.New(store), redis_rate.NewLimiter(store.Rdb.Instance())}
}

func (s middleware) NoRoute() gin.HandlerFunc {
	return func(c *gin.Context) { c.Error(ecode.ApiNotFound) }
}

func (s middleware) Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowWebSockets = true
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization")
	config.AddExposeHeaders("X-Pagination-Total", "X-Pagination-Page", "X-Pagination-Limit")
	return cors.New(config)
}

func (s middleware) Compress() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

// SecureHeaders adds common security-related HTTP headers.
func (s middleware) SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		// Basic CSP, can be expanded. Allows resources only from the same origin.
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		// X-XSS-Protection is deprecated by modern browsers in favor of CSP.
		// HSTS (Strict-Transport-Security) should be added carefully after confirming site-wide HTTPS.
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	data *bytes.Buffer
}

func (s *responseWriter) Write(b []byte) (int, error) {
	s.data.Write(b)
	return s.ResponseWriter.Write(b)
}

func (s middleware) Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(trace.New(c.Request.Context()))

		var (
			body, data any
			start      = time.Now()
			writer     = &responseWriter{c.Writer, &bytes.Buffer{}}
		)

		c.Writer = writer

		if slices.Contains([]string{http.MethodPost, http.MethodPut}, c.Request.Method) {
			if raw, _ := c.GetRawData(); len(raw) > 0 {
				json.Unmarshal(raw, &body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
			}
		}

		c.Next()

		if values := trace.Value(c.Request.Context()); len(values) > 0 {
			json.Unmarshal(writer.data.Bytes(), &data)

			var (
				session  = s.Session(c)
				request  = trace.E{K: "request", V: db.M{"method": c.Request.Method, "path": c.Request.URL.String(), "body": body, "tenant": session.TenantId, "username": session.Username}}
				response = trace.E{K: "response", V: db.M{"status": writer.Status(), "latency": time.Since(start).String(), "data": data}}
				traces   = []trace.E{request}
			)

			if slices.Contains([]reflect.Kind{reflect.Slice, reflect.Array}, reflect.ValueOf(data).Kind()) {
				response = trace.E{K: "response", V: db.M{"status": writer.Status(), "latency": time.Since(start).String()}}
			}

			traces = append(traces, values...)
			traces = append(traces, response)

			traceStr := strings.Join(gopkg.MapFunc(traces, func(e trace.E) string { return e.String() }), " -> ")
			fmt.Printf("[TRA] %s | %s\n", time.Now().Format("2006/01/02 - 15:04:05"), traceStr)

			trace.Clear(c.Request.Context())
		}
	}
}

func (s middleware) Logger() gin.HandlerFunc {
	return gin.Logger()
}

func (s middleware) Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		s.ErrorFunc(c, err)
	})
}

func (s middleware) Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			s.ErrorFunc(c, c.Errors.Last().Err)
		}
	}
}

func (s middleware) ErrorFunc(c *gin.Context, err any) {
	switch e := err.(type) {
	case *ecode.Error:
		c.JSON(e.Status, e)
	default:
		err := ecode.InternalServerError.Stack(fmt.Errorf("%v", e))
		c.JSON(err.Status, err)
	}
}

func (s middleware) NoAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.Limiter(c, nil) {
			c.Error(ecode.TooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s middleware) BearerAuth(permissions ...enum.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := s.oauth.BearerAuth(c.Request)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if len(permissions) > 0 && !slices.ContainsFunc(permissions, func(e enum.Permission) bool { return slices.Contains(session.Permissions, e) }) {
			c.Error(ecode.Forbidden)
			c.Abort()
			return
		}
		if s.Limiter(c, session) {
			c.Error(ecode.TooManyRequests)
			c.Abort()
			return
		}
		s.Session(c, session)
		c.Next()
	}
}

func (s middleware) BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := s.oauth.BasicAuth(c.Request)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		if s.Limiter(c, session) {
			c.Error(ecode.TooManyRequests)
			c.Abort()
			return
		}
		s.Session(c, session)
		c.Next()
	}
}

func (s middleware) RequireKYCStatus(requiredStatus string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		if session.IsMember == nil || !*session.IsMember {
			c.Error(ecode.New(http.StatusForbidden, "member_required_for_kyc_check").Desc(fmt.Errorf("This action requires a member account.")))
			c.Abort()
			return
		}

		if session.KYCStatus == nil {
			c.Error(ecode.New(http.StatusForbidden, "kyc_status_missing").Desc(fmt.Errorf("KYC status is not available for this member.")))
			c.Abort()
			return
		}

		if *session.KYCStatus != requiredStatus {
			formattedMsg := fmt.Sprintf("Action requires KYC status '%s', but current status is '%s'.", requiredStatus, *session.KYCStatus)
			c.Error(ecode.New(http.StatusForbidden, "kyc_status_insufficient").Desc(fmt.Errorf(formattedMsg)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMembershipStatus is a middleware that checks if the authenticated user is a member
// and has the specified Membership status. It should be used AFTER BearerAuth.
func (s middleware) RequireMembershipStatus(requiredStatus string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.Session(c)

		// Check if IsMember is populated and true
		if session.IsMember == nil || !*session.IsMember {
			c.Error(ecode.New(http.StatusForbidden, "member_required_for_membership_check").Desc(fmt.Errorf("This action requires a member account for membership status validation.")))
			c.Abort()
			return
		}

		// Check if MembershipStatus is populated
		if session.MembershipStatus == nil {
			c.Error(ecode.New(http.StatusForbidden, "membership_status_missing").Desc(fmt.Errorf("Membership status is not available for this member.")))
			c.Abort()
			return
		}

		// Compare MembershipStatus with the required status
		if *session.MembershipStatus != requiredStatus {
			formattedMsg := fmt.Sprintf("Action requires membership status '%s', but current status is '%s'.", requiredStatus, *session.MembershipStatus)
			c.Error(ecode.New(http.StatusForbidden, "membership_status_insufficient").Desc(fmt.Errorf(formattedMsg)))
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s middleware) Limiter(c *gin.Context, session *db.AuthSessionDto) bool {
	var (
		key string
		err error
		res *redis_rate.Result
	)

	if session == nil || session.Username == "" {
		key = c.ClientIP()
	} else {
		key = session.Username
	}
	// TODO: revisit member auth endpoint limits later. (Original TODO)

	switch c.FullPath() {
	case "/auth/v1/member/login": // Apply specific stricter limit for member login
		res, err = s.limiter.Allow(c.Request.Context(), key, memberLoginRateLimit)
	case "/auth/v1/login", "/auth/v1/register", "/auth/v1/refresh-token":
		res, err = s.limiter.Allow(c.Request.Context(), key, authLimit)
	default:
		res, err = s.limiter.Allow(c.Request.Context(), key, defaultLimit)
	}

	if err != nil {
		return true // Fail open on limiter error (block request)
	}
	return res.Allowed == 0
}

func (s middleware) Session(c *gin.Context, session ...*db.AuthSessionDto) *db.AuthSessionDto {
	if len(session) == 0 {
		session, ok := c.Get(reflect.TypeOf(db.AuthSessionDto{}).Name())
		if !ok {
			return &db.AuthSessionDto{}
		}
		return session.(*db.AuthSessionDto)
	}
	c.Set(reflect.TypeOf(db.AuthSessionDto{}).Name(), session[0])
	return session[0]
}

func (s middleware) Pagination(c *gin.Context, total, page, limit int64) {
	c.Header("X-Pagination-Total", strconv.Itoa(int(total)))
	c.Header("X-Pagination-Page", strconv.Itoa(int(page)))
	c.Header("X-Pagination-Limit", strconv.Itoa(int(limit)))
}

func (s middleware) Permissions(session *db.AuthSessionDto) []enum.Permission {
	permissions := enum.PermissionRootValues()
	if session.IsTenant {
		permissions = enum.PermissionTenantValues()
	}
	return permissions
}

func (s middleware) AuditLog(c *gin.Context, name string, action enum.DataAction, data, domain any, domain_id string) {
	session := s.Session(c)
	byteData, _ := json.Marshal(data)
	byteDomain, _ := json.Marshal(domain)

	audit := &db.AuditLogDomain{}
	audit.Name = gopkg.Pointer(name)
	audit.Url = gopkg.Pointer(c.Request.URL.String())
	audit.Method = gopkg.Pointer(c.Request.Method)
	audit.Data = gopkg.Pointer(byteData)
	audit.Domain = gopkg.Pointer(byteDomain)
	audit.DomainId = gopkg.Pointer(domain_id)
	audit.Action = gopkg.Pointer(action)
	audit.TenantId = gopkg.Pointer(session.TenantId)
	audit.UpdatedBy = gopkg.Pointer(session.Username)

	s.store.Db.AuditLog.Save(c.Request.Context(), audit)
}
