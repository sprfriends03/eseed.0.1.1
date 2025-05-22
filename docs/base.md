# Backend Features Documentation

## Core Infrastructure

### Authentication & Authorization
**Implementation Files:**
- `pkg/oauth/index.go`: Core OAuth implementation
- `route/auth.go`: Authentication endpoints
- `pkg/enum/index.go`: Permission definitions
- `route/index.go`: Middleware for auth

**Key Components:**
```go
// Authentication Middleware (route/index.go)
func (s middleware) BearerAuth(permissions ...enum.Permission) gin.HandlerFunc
func (s middleware) BasicAuth() gin.HandlerFunc
func (s middleware) NoAuth() gin.HandlerFunc

// Token Management (pkg/oauth/index.go)
func (s *Oauth) GenerateToken(ctx context.Context, uid string) (*db.AuthTokenDto, error)
func (s *Oauth) ValidateToken(ctx context.Context, access string) (*db.AuthSessionDto, error)
func (s *Oauth) RefreshToken(ctx context.Context, refresh string) (*db.AuthTokenDto, error)
```

**Implementation Rules:**
- All protected routes must use `BearerAuth` middleware with required permissions
- Token validation includes signature, expiry, and revocation checks
- Rate limiting applies to all authentication attempts
- Password hashing is mandatory before storage

### API Layer
**Implementation Files:**
- `route/index.go`: Core routing and middleware setup
- `pkg/ecode/index.go`: Error definitions
- `docs/swagger.yaml`: API documentation

**Key Components:**
```go
// Middleware Chain (route/index.go)
app.Use(
    mdw.Cors(),
    mdw.Compress(),
    mdw.Trace(),
    mdw.Logger(),
    mdw.Recover(),
    mdw.Error()
)

// Error Handling (pkg/ecode/index.go)
type Error struct {
    Status   int    `json:"-"`
    ErrCode  string `json:"error"`
    ErrDesc  string `json:"error_description"`
    ErrStack string `json:"-"`
}
```

**Implementation Rules:**
- All responses must use standard error format
- Rate limiting configuration varies by endpoint type
- All endpoints must be documented with Swagger annotations
- CORS headers must be properly set for web clients

### Data Storage & Caching
**Implementation Files:**
- `store/db/index.go`: MongoDB connection and models
- `store/rdb/index.go`: Redis operations
- `store/storage/index.go`: File storage operations
- `pkg/ws/index.go`: WebSocket handling

**Key Components:**
```go
// Store Structure (store/index.go)
type Store struct {
    Db      *db.Db
    Rdb     *rdb.Rdb
    Storage *storage.Storage
}

// Cache Operations (store/rdb/index.go)
func (s Rdb) Get(ctx context.Context, key string, value any) error
func (s Rdb) Set(ctx context.Context, key string, value any, exp time.Duration) error
```

**Configuration Examples:**
```yaml
# MongoDB Configuration
mongodb:
  uri: mongodb://localhost:27017
  database: app
  options:
    maxPoolSize: 100
    connectTimeoutMS: 5000

# Redis Configuration
redis:
  uri: redis://localhost:6379
  database: 0
  options:
    maxRetries: 3
    poolSize: 100

# MinIO Configuration
minio:
  endpoint: localhost:9000
  accessKey: minioadmin
  secretKey: minioadmin
  useSSL: false
  bucketName: app
```

**Cache Implementation Example:**
```go
// Cache Implementation (store/rdb/index.go)
func (s Rdb) GetWithRefresh(ctx context.Context, key string, value any, refresh func() (any, error), exp time.Duration) error {
    // Try getting from cache first
    if err := s.Get(ctx, key, value); err == nil {
        return nil
    }

    // If not in cache, refresh the data
    data, err := refresh()
    if err != nil {
        return err
    }

    // Store in cache and return
    if err := s.Set(ctx, key, data, exp); err != nil {
        return err
    }
    
    return copier.Copy(value, data)
}
```

## Feature Modules

### User Management
**Implementation Files:**
- `route/user.go`: User endpoints
- `store/db/user.go`: User data models and queries
- `pkg/mail/index.go`: Email notifications

**Key Endpoints:**
```go
// User Management Routes (route/user.go)
v1cms.GET("", s.BearerAuth(enum.PermissionUserView), s.v1cms_List())
v1cms.POST("", s.BearerAuth(enum.PermissionUserCreate), s.v1cms_Create())
v1cms.PUT("/:user_id", s.BearerAuth(enum.PermissionUserUpdate), s.v1cms_Update())
v1cms.POST("/:user_id/reset-password", s.BearerAuth(enum.PermissionUserUpdate), s.v1cms_ResetPassword())
```

**Implementation Rules:**
- User passwords must be hashed before storage
- Email notifications required for password resets
- User data must be tenant-isolated
- Role assignments must be validated

**Advanced Implementation Details:**

1. Password Management:
```go
// Password Hashing (pkg/util/password.go)
func HashPassword(password string) string {
    bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes)
}

func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

2. User Cache Management:
```go
// User Cache (store/index.go)
func (s Store) GetUser(ctx context.Context, id string) (*db.UserCache, error) {
    key := fmt.Sprintf("user:%v", id)
    cache := &db.UserCache{}

    // Try cache first
    if err := s.Rdb.Get(ctx, key, &cache); err == nil {
        return cache, nil
    }

    // Load from database
    obj, err := s.Db.User.FindOneById(ctx, id)
    if err == nil {
        cache = obj.Cache()
        s.Rdb.Set(ctx, key, cache, time.Hour*24)
    }

    return cache, err
}
```

3. Role Permission Validation:
```go
// Permission Validation (route/role.go)
if slices.ContainsFunc(data.Permissions, func(e enum.Permission) bool {
    return !slices.Contains(s.Permissions(session), e)
}) {
    c.Error(ecode.InvalidPermission)
    return
}
```

### Role Management
**Implementation Files:**
- `route/role.go`: Role endpoints
- `store/db/role.go`: Role data models
- `pkg/enum/index.go`: Permission definitions

**Key Endpoints:**
```go
// Role Management Routes (route/role.go)
v1cms.GET("", s.BearerAuth(enum.PermissionRoleView), s.v1cms_List())
v1cms.POST("", s.BearerAuth(enum.PermissionRoleCreate), s.v1cms_Create())
v1cms.PUT("/:role_id", s.BearerAuth(enum.PermissionRoleUpdate), s.v1cms_Update())
```

**Implementation Rules:**
- Roles must be tenant-specific
- Permission validation on role creation/update
- Cache invalidation on role updates
- Audit logging for all role changes

### Tenant Management
**Implementation Files:**
- `route/tenant.go`: Tenant endpoints
- `store/db/tenant.go`: Tenant data models
- `pkg/mail/index.go`: Email notifications

**Key Endpoints:**
```go
// Tenant Management Routes (route/tenant.go)
v1cms.GET("", s.BearerAuth(enum.PermissionTenantView), s.v1cms_List())
v1cms.POST("", s.BearerAuth(enum.PermissionTenantCreate), s.v1cms_Create())
v1cms.PUT("/:tenant_id", s.BearerAuth(enum.PermissionTenantUpdate), s.v1cms_Update())
v1cms.POST("/:tenant_id/reset-password", s.BearerAuth(enum.PermissionTenantUpdate), s.v1cms_ResetPassword())
```

**Implementation Rules:**
- Root tenant cannot be modified
- Tenant creation includes default admin user
- Cascade updates to related entities
- Email notifications for critical changes

### File Storage
**Implementation Files:**
- `route/storage.go`: Storage endpoints
- `store/storage/index.go`: Storage operations
- `pkg/validate/index.go`: File validation

**Key Endpoints:**
```go
// Storage Routes (route/storage.go)
v1.GET("/images/:filename", s.NoAuth(), s.v1_DownloadImage())
v1.GET("/videos/:filename", s.NoAuth(), s.v1_DownloadVideo())
v1.POST("/images", s.BearerAuth(), s.v1_UploadImage([]string{"image"}))
v1.POST("/videos", s.BearerAuth(), s.v1_UploadVideo([]string{"video"}))
```

**Implementation Rules:**
- Content type validation required
- File size limits enforced
- CDN integration for downloads
- Cache headers for optimization

### WebSocket Support
**Implementation Files:**
- `route/wss.go`: WebSocket endpoints
- `pkg/ws/index.go`: WebSocket implementation
- `pkg/ws/conn.go`: Connection handling

**Key Components:**
```go
// WebSocket Implementation (pkg/ws/index.go)
type Ws struct {
    sync.RWMutex
    store *store.Store
    users gopkg.CacheShard[[]*Conn]
}

func (s *Ws) Broadcast(data []byte)
func (s *Ws) EmitTo(users []string, data []byte)
```

**Implementation Rules:**
- Authentication required for connections
- Message validation before broadcast
- Rate limiting per connection
- Automatic reconnection support

### Security Implementation
**Implementation Files:**
- `route/index.go`: Security middleware
- `pkg/oauth/index.go`: Token management
- `pkg/validate/index.go`: Input validation
- `store/db/audit.go`: Audit logging

**Key Security Measures:**
```go
// Rate Limiting (route/index.go)
func (s middleware) Limiter(c *gin.Context, session *db.AuthSessionDto) bool {
    var (
        key   = c.Request.URL.Path + c.ClientIP()
        limit = redis_rate.PerMinute(100)
    )
    // ... rate limiting logic
}

// Audit Logging (route/index.go)
func (s middleware) AuditLog(c *gin.Context, name string, action enum.DataAction, data, domain any, domain_id string)
```

**Implementation Rules:**
- All user input must be validated
- Sensitive data must be encrypted
- Rate limiting on all endpoints
- Comprehensive audit logging
- CORS headers properly configured
- Secure session management

**Additional Security Measures:**

1. JWT Token Structure:
```go
// Token Claims (pkg/oauth/index.go)
type AccessClaims struct {
    jwt.RegisteredClaims
    JwtType  string `json:"jwt_type"`
    Version  int64  `json:"version"`
    Username string `json:"username"`
}

// Token Generation
func (s *Oauth) generateAccessToken(ctx context.Context, user *db.UserCache) (string, error) {
    now := time.Now()
    claims := &AccessClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Subject:   user.ID,
            ID:        uuid.NewString(),
        },
        JwtType:  JwtTypeAccess,
        Version:  user.VersionToken,
        Username: user.Username,
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}
```

2. Rate Limiting Configuration:
```go
// Rate Limit Configuration (route/index.go)
var (
    defaultLimit = redis_rate.PerMinute(100)
    authLimit    = redis_rate.Limit{
        Rate:   10,
        Burst:  3,
        Period: time.Minute,
    }
    uploadLimit = redis_rate.Limit{
        Rate:   50,
        Burst:  10,
        Period: time.Minute,
    }
)
```

## Development Guidelines

### Code Organization
- Routes defined in `route/` directory
- Business logic in `pkg/` directory
- Data access in `store/` directory
- Middleware in `route/index.go`
- Models in respective `store/db/` files

### Error Handling
- Use defined error codes from `pkg/ecode/index.go`
- Include appropriate HTTP status codes
- Provide descriptive error messages
- Log errors appropriately

### Testing
- Unit tests for business logic
- Integration tests for APIs
- Mock external dependencies
- Test security measures

### Documentation
- Swagger annotations for all endpoints
- Clear code comments
- Updated README.md
- API documentation maintenance

### Error Handling Examples
```go
// Standard Error Response
type Error struct {
    Status   int    `json:"-"`
    ErrCode  string `json:"error"`
    ErrDesc  string `json:"error_description"`
    ErrStack string `json:"-"`
}

// Error Creation
var (
    InternalServerError     = New(http.StatusInternalServerError, "internal_server_error")
    TooManyRequests        = New(http.StatusTooManyRequests, "too_many_requests")
    Unauthorized           = New(http.StatusUnauthorized, "unauthorized")
    Forbidden             = New(http.StatusForbidden, "forbidden")
)

// Error Usage
if err != nil {
    return ecode.InternalServerError.Desc(err)
}
```

### Testing Examples
```go
// Unit Test Example
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *UserCreateInput
        want    *User
        wantErr bool
    }{
        {
            name: "valid user creation",
            input: &UserCreateInput{
                Username: "testuser",
                Email:    "test@example.com",
            },
            want: &User{
                Username: "testuser",
                Email:    "test@example.com",
                Status:   StatusActive,
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Configuration Management
```go
// Environment Variables (env/index.go)
var (
    Port         = getEnv("PORT", "3000")
    Environment  = getEnv("ENVIRONMENT", "development")
    MongoUri     = getEnv("MONGO_URI", "mongodb://localhost:27017")
    RedisUri     = getEnv("REDIS_URI", "redis://localhost:6379")
    MinioUri     = getEnv("MINIO_URI", "localhost:9000")
    RootUser     = getEnv("ROOT_USER", "admin")
    RootPass     = getEnv("ROOT_PASS", "admin")
    ClientId     = getEnv("CLIENT_ID", "client")
    ClientSecret = getEnv("CLIENT_SECRET", "secret")
)

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

### Deployment Configuration

**Docker Compose Example:**
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - MONGO_URI=mongodb://mongo:27017
      - REDIS_URI=redis://redis:6379
      - MINIO_URI=minio:9000
    depends_on:
      - mongo
      - redis
      - minio

  mongo:
    image: mongo:latest
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  redis:
    image: redis:latest
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    ports:
      - "9000:9000"
    environment:
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
    command: server /data

volumes:
  mongo_data:
```

### Performance Optimization

1. Database Indexing:
```go
// MongoDB Indexes (store/db/index.go)
func (s *Db) createIndexes(ctx context.Context) error {
    // User indexes
    s.User.CreateIndexes(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "tenant_id", Value: 1}}},
    })

    // Role indexes
    s.Role.CreateIndexes(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "name", Value: 1}, {Key: "tenant_id", Value: 1}}, Options: options.Index().SetUnique(true)},
    })

    return nil
}
```

2. Cache Strategy:
```go
// Cache TTL Configuration
const (
    UserCacheTTL    = time.Hour * 24
    RoleCacheTTL    = time.Hour * 24
    SessionCacheTTL = time.Hour * 1
)

// Cache Key Patterns
const (
    UserCacheKey    = "user:%s"
    RoleCacheKey    = "role:%s"
    SessionCacheKey = "session:%s"
)
```

### Monitoring and Logging

1. Trace Implementation:
```go
// Trace Middleware (route/index.go)
func (s middleware) Trace() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        logrus.WithFields(logrus.Fields{
            "status":     status,
            "latency":    latency,
            "path":       path,
            "query":      query,
            "ip":        c.ClientIP(),
            "method":    c.Request.Method,
            "user_agent": c.Request.UserAgent(),
        }).Info("request")
    }
}
```

2. Health Check Implementation:
```go
// Health Check (route/docs.go)
func init() {
    handlers = append(handlers, func(m *middleware, r *gin.Engine) {
        r.GET("/healthz", func(c *gin.Context) {
            health := struct {
                Status    string `json:"status"`
                Timestamp string `json:"timestamp"`
            }{
                Status:    "ok",
                Timestamp: time.Now().Format(time.RFC3339),
            }
            c.JSON(http.StatusOK, health)
        })
    })
}
```

## Database Structure

### MongoDB Collections

#### 1. User Collection (`user`)
```javascript
{
    _id: ObjectId,
    name: String,
    phone: String,          // Unique per tenant, lowercase
    email: String,          // Unique per tenant, lowercase
    username: String,       // Unique per tenant, lowercase
    password: String,       // Hashed
    data_status: String,    // enum: ["enable", "disable"]
    role_ids: [String],     // Array of role ObjectIds
    is_root: Boolean,
    tenant_id: String,      // Reference to tenant
    version_token: Number,  // For token versioning
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1, username: 1} unique
- {tenant_id: 1, phone: 1} unique
- {tenant_id: 1, email: 1} unique
```

#### 2. Role Collection (`role`)
```javascript
{
    _id: ObjectId,
    name: String,
    permissions: [String],  // Array of permission enums
    data_status: String,    // enum: ["enable", "disable"]
    tenant_id: String,      // Reference to tenant
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
```

#### 3. Tenant Collection (`tenant`)
```javascript
{
    _id: ObjectId,
    name: String,
    keycode: String,       // Unique, lowercase
    username: String,      // lowercase
    phone: String,        // lowercase
    email: String,        // lowercase
    address: String,
    data_status: String,  // enum: ["enable", "disable"]
    is_root: Boolean,
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {keycode: 1} unique
```

#### 4. Client Collection (`client`)
```javascript
{
    _id: ObjectId,
    name: String,
    client_id: String,     // Unique
    client_secret: String, // Encrypted
    secure_key: String,    // Encrypted
    is_root: Boolean,
    tenant_id: String,     // Reference to tenant
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {client_id: 1} unique
- {tenant_id: 1}
```

#### 5. Audit Log Collection (`audit_log`)
```javascript
{
    _id: ObjectId,
    name: String,          // Module name
    url: String,           // API endpoint
    method: String,        // HTTP method
    data: Binary,          // Request/change data
    domain: Binary,        // Domain object data
    domain_id: String,     // Reference to affected document
    action: String,        // enum: DataAction
    tenant_id: String,     // Reference to tenant
    created_at: DateTime,
    updated_at: DateTime,
    created_by: String,
    updated_by: String
}

Indexes:
- {tenant_id: 1}
```

### Common Features Across Collections

1. **Base Domain Fields**
   - All collections inherit these fields:
   ```javascript
   {
       _id: ObjectId,
       created_at: DateTime,
       updated_at: DateTime,
       created_by: String,
       updated_by: String
   }
   ```

2. **Multi-tenancy**
   - Most collections have `tenant_id` field
   - Unique constraints are tenant-scoped
   - Root entities are marked with `is_root: true`

3. **Status Management**
   - `data_status` field for entity state
   - Common values: "enable", "disable"

4. **Audit Trail**
   - Creation and modification timestamps
   - User tracking for creates/updates
   - Detailed audit logging in `audit_log` collection

### Data Relationships

1. **User → Role**
   - Many-to-Many relationship
   - Users contain array of `role_ids`
   - Roles contain permissions

2. **User/Role/Client → Tenant**
   - Many-to-One relationship
   - Each entity belongs to one tenant
   - Referenced by `tenant_id`

3. **Audit Log → All**
   - One-to-Many relationship
   - Logs reference source entity via `domain_id`
   - Stores both before/after states

### Security Features

1. **Password Management**
   - Passwords are hashed before storage
   - Version token for session management

2. **Client Authentication**
   - Client secrets are encrypted
   - Secure keys for API access

3. **Data Isolation**
   - Tenant-based data segregation
   - Root-level access control
   - Permission-based authorization
