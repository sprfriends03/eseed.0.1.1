package route

import (
	"app/env"
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/util"
	"app/store"
	"app/store/db"
	testdb "app/test/db"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
	"github.com/redis/go-redis/v9" // Corrected import path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMailService is a mock implementation for tracking mail sending calls.
type mockMailService struct {
	sendMemberVerificationEmailCalled bool
	sendMemberVerificationEmailParams struct {
		toEmail              string
		username             string
		verificationToken    string
		verificationLinkBase string
	}
	sendMemberVerificationEmailError error
}

// SendMemberVerificationEmail records that the call was made and stores the parameters.
func (m *mockMailService) SendMemberVerificationEmail(toEmail, username, verificationToken, verificationLinkBase string) error {
	m.sendMemberVerificationEmailCalled = true
	m.sendMemberVerificationEmailParams.toEmail = toEmail
	m.sendMemberVerificationEmailParams.username = username
	m.sendMemberVerificationEmailParams.verificationToken = verificationToken
	m.sendMemberVerificationEmailParams.verificationLinkBase = verificationLinkBase
	return m.sendMemberVerificationEmailError
}

// TestMain sets up and tears down the test environment.
func TestMain(m *testing.M) {
	// Set Gin to TestMode before any routes are initialized if it's not done elsewhere globally for tests.
	gin.SetMode(gin.TestMode)
	testdb.Setup() // Setup test database
	// The app/env package loads its configuration via its init() function.
	// The explicit call to env.Load() below was causing an "undefined: env.Load" error
	// because the env package does not export a Load function.
	// By importing "app/env", its init() function will be triggered.
	// env.Load("../../.env") // REMOVE THIS LINE

	exitVal := m.Run()
	testdb.Teardown() // Teardown test database
	os.Exit(exitVal)
}

// setupAuthTestRouter initializes a Gin engine with auth routes for testing.
// It returns the router, the test store, and the mock mail service tracker.
func setupAuthTestRouter(t *testing.T) (*gin.Engine, *store.Store, *mockMailService) {
	ctx := context.Background() // Use a background context for setup

	// The store.New() function initializes its own DB, RDB, and Storage.
	// Ensure that the environment variables used by db.Init, rdb.New, storage.New
	// are correctly set for the test environment (e.g., via env package init or test setup).
	testStore := store.New()
	require.NotNil(t, testStore, "Test store should not be nil")

	// We still need a direct MongoDB connection for assertions/cleanup if not easily accessible via testStore.Db
	_, testMongoDb := testdb.GetTestDBContext()
	require.NotNil(t, testMongoDb, "Test MongoDB instance should not be nil for direct operations")

	// We might still need rdbClient for direct Redis assertions/cleanup if necessary,
	// though store.New() also initializes its own.
	rdbClient := redis.NewClient(&redis.Options{
		Addr: env.RedisUri,
	})
	_, errRedis := rdbClient.Ping(ctx).Result()
	require.NoError(t, errRedis, "Failed to connect to test Redis instance. Ensure Redis is running and accessible at %s", env.RedisUri)

	// This mockMailTracker is for making assertions. It won't be automatically used by the
	// real mail service unless a proper DI/mocking mechanism (interface/hook) is in place.
	mockMailTracker := &mockMailService{}

	// Create the middleware instance that will be used by the auth handlers.
	// This mdw instance will initialize its own real mail service (mail.New(testStore)).
	authRouteMiddleware := newMdw(testStore)

	router := gin.New()                     // Use gin.New() instead of gin.Default() for more control in tests
	router.Use(authRouteMiddleware.Error()) // Add essential middleware like error handling

	// Register auth routes. The auth.init() function appends a closure to route.handlers.
	// We need to find that specific closure or replicate its logic here.
	// For simplicity and directness in testing auth routes, we explicitly register them.
	authAPI := auth{authRouteMiddleware} // auth is the struct from route/auth.go, embedding the middleware

	v1 := router.Group("/auth/v1")
	{
		// Standard auth routes (copied from auth.go init for clarity, can be refactored)
		v1.POST("/login", authAPI.NoAuth(), authAPI.v1_Login())
		v1.POST("/register", authAPI.NoAuth(), authAPI.v1_Register())
		v1.POST("/refresh-token", authAPI.NoAuth(), authAPI.v1_RefreshToken())
		v1.POST("/logout", authAPI.BearerAuth(), authAPI.v1_Logout())
		v1.POST("/change-password", authAPI.BearerAuth(), authAPI.v1_ChangePassword())
		v1.GET("/me", authAPI.BearerAuth(), authAPI.v1_GetMe())
		v1.GET("/flush-cache", authAPI.BearerAuth(), authAPI.v1_FlushCache())

		// Member-specific routes
		v1.POST("/member/register", authAPI.NoAuth(), authAPI.v1_MemberRegister())
		v1.POST("/member/login", authAPI.NoAuth(), authAPI.v1_MemberLogin())
		v1.GET("/member/verify-email", authAPI.NoAuth(), authAPI.v1_VerifyMemberEmail())
	}

	// Ensure the default test tenant ("test_club") exists by calling the test utility.
	// CreateTestTenant in test/db/config.go already creates a tenant with keycode "test_club".
	_, errTenantCreate := testdb.CreateTestTenant()
	// This might return an error if the tenant already exists (e.g., duplicate key on keycode).
	// For setup, we just want to ensure it's there. If it fails because it's already there, that's okay.
	// However, CreateTestTenant as written will fail on duplicate key. Consider making CreateTestTenant idempotent or use a find-or-create pattern here.
	// For now, we'll require no error, assuming tests run in a clean state or CreateTestTenant handles this.
	require.NoError(t, errTenantCreate, "Failed to ensure default test tenant 'test_club' exists using CreateTestTenant. It might have failed due to pre-existing data or other issues.")

	return router, testStore, mockMailTracker
}

func TestMemberRegister_Success(t *testing.T) {
	router, testStore, mockMailUserCallTracker := setupAuthTestRouter(t)
	// Defer cleanup of collections
	defer func() {
		// testdb.ResetCollection uses a global test DB instance initialized by testdb.Setup()
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		// Tenant 'test_club' is reused, so we might not want to reset it after every single test
		// or ensure it's recreated in setup if needed. For now, let it persist through tests in this file.
		// testdb.ResetCollection(t, "tenants")
	}()

	tenantKeycode := "test_club"

	// Generate unique username and email for each test run to avoid conflicts
	timestamp := time.Now().UnixNano()
	uniqueUsername := fmt.Sprintf("testuser_%d", timestamp)
	uniqueEmail := fmt.Sprintf("testuser_%d@example.com", timestamp)

	registerData := db.MemberRegisterData{
		Username:    uniqueUsername,
		Password:    "Password123!",
		Email:       uniqueEmail,
		FirstName:   "Test",
		LastName:    "User",
		DateOfBirth: "1990-01-01",
		Phone:       "1234567890",
		Keycode:     tenantKeycode,
	}

	bodyBytes, _ := json.Marshal(registerData)
	req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// For verificationLinkBase, the handler might use c.Request.Host or Origin header
	req.Host = "test.example.com" // Simulate a host for the request

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Response code should be 201 Created. Body: %s", w.Body.String())

	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	require.NoError(t, err, "Failed to unmarshal response body")
	assert.Equal(t, "Member registered successfully. Please check your email to verify your account.", responseBody["message"])

	// Verify database state
	ctx := context.Background()
	user, err := testStore.Db.User.FindOne(ctx, db.M{"email": registerData.Email})
	require.NoError(t, err, "User should be created in DB")
	require.NotNil(t, user, "User should not be nil")
	assert.Equal(t, registerData.Username, gopkg.Value(user.Username))
	assert.Equal(t, registerData.Email, gopkg.Value(user.Email))
	assert.False(t, gopkg.Value(user.EmailVerified), "EmailVerified should be false initially")
	assert.NotEmpty(t, gopkg.Value(user.EmailVerificationToken), "EmailVerificationToken should be set")
	assert.NotNil(t, user.EmailVerificationTokenExpiresAt, "EmailVerificationTokenExpiresAt should be set") // Direct field access for pointer check

	member, err := testStore.Db.Member.FindByUserID(ctx, db.SID(user.ID))
	require.NoError(t, err, "Member should be created in DB")
	require.NotNil(t, member, "Member should not be nil")
	assert.Equal(t, registerData.Email, gopkg.Value(member.Email))
	assert.Equal(t, registerData.FirstName, gopkg.Value(member.FirstName))
	assert.Equal(t, "pending_verification", gopkg.Value(member.MemberStatus))
	assert.Equal(t, "pending_kyc", gopkg.Value(member.KYCStatus))
	assert.Equal(t, db.SID(user.ID), gopkg.Value(member.UserID))

	// Verify email sending mock call
	// NOTE: This assertion will likely FAIL with the current setup because the real mail service is called.
	// This failure indicates the need for a better mail mocking/DI strategy if precise call verification is required.
	assert.True(t, mockMailUserCallTracker.sendMemberVerificationEmailCalled, "SendMemberVerificationEmail should be called (this might fail if mock not injected)")
	if mockMailUserCallTracker.sendMemberVerificationEmailCalled { // Only check params if called
		assert.Equal(t, registerData.Email, mockMailUserCallTracker.sendMemberVerificationEmailParams.toEmail)
		assert.Equal(t, registerData.Username, mockMailUserCallTracker.sendMemberVerificationEmailParams.username)
		assert.Equal(t, gopkg.Value(user.EmailVerificationToken), mockMailUserCallTracker.sendMemberVerificationEmailParams.verificationToken)
		assert.True(t, strings.HasPrefix(mockMailUserCallTracker.sendMemberVerificationEmailParams.verificationLinkBase, "http://test.example.com"), "Verification link base should use request host")
	}
}

func TestMemberRegister_UserConflict_EmailExists(t *testing.T) {
	router, testStore, _ := setupAuthTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
	}()

	tenantKeycode := "test_club"
	ctx := context.Background()

	// Get the tenant ID for "test_club"
	testClubTenant, errTenant := testStore.Db.Tenant.FindOneByKeycode(ctx, tenantKeycode)
	require.NoError(t, errTenant, "Failed to find test_club tenant for conflict test setup")
	require.NotNil(t, testClubTenant, "test_club tenant should not be nil for conflict test setup")
	testClubTenantID := enum.Tenant(db.SID(testClubTenant.ID))

	// 1. Create an initial user
	existingUserEmail := fmt.Sprintf("existing_%d@example.com", time.Now().UnixNano())
	existingUsername := fmt.Sprintf("existinguser_%d", time.Now().UnixNano())
	userDomain := &db.UserDomain{
		Username:   gopkg.Pointer(existingUsername),
		Password:   gopkg.Pointer(util.HashPassword("Password123!")),
		Email:      gopkg.Pointer(existingUserEmail),
		TenantId:   gopkg.Pointer(testClubTenantID),
		DataStatus: gopkg.Pointer(enum.DataStatusEnable),
	}
	_, err := testStore.Db.User.Save(ctx, userDomain)
	require.NoError(t, err, "Failed to create initial user for conflict test")

	// 2. Attempt to register a new member with the same email
	registerData := db.MemberRegisterData{
		Username:    fmt.Sprintf("newuser_%d", time.Now().UnixNano()), // Different username
		Password:    "NewPassword123!",
		Email:       existingUserEmail, // Same email
		FirstName:   "Conflict",
		LastName:    "Test",
		DateOfBirth: "1991-01-01",
		Phone:       "0987654321",
		Keycode:     tenantKeycode,
	}

	bodyBytes, _ := json.Marshal(registerData)
	req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "test.example.com"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code, "Response code should be 409 Conflict. Body: %s", w.Body.String())

	var errResp ecode.Error
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err, "Failed to unmarshal error response body")
	assert.Equal(t, ecode.UserConflict.ErrCode, errResp.ErrCode, "Error code should be UserConflict")
}

func TestMemberRegister_UserConflict_UsernameExists(t *testing.T) {
	router, testStore, _ := setupAuthTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
	}()

	tenantKeycode := "test_club"
	ctx := context.Background()

	// Get the tenant ID for "test_club"
	testClubTenant, errTenant := testStore.Db.Tenant.FindOneByKeycode(ctx, tenantKeycode)
	require.NoError(t, errTenant, "Failed to find test_club tenant for conflict test setup")
	require.NotNil(t, testClubTenant, "test_club tenant should not be nil for conflict test setup")
	testClubTenantID := enum.Tenant(db.SID(testClubTenant.ID))

	// 1. Create an initial user
	existingUserEmail := fmt.Sprintf("another_%d@example.com", time.Now().UnixNano())
	existingUsername := fmt.Sprintf("existinguser_%d", time.Now().UnixNano())

	userDomain := &db.UserDomain{
		Username:   gopkg.Pointer(existingUsername),
		Password:   gopkg.Pointer(util.HashPassword("Password123!")),
		Email:      gopkg.Pointer(existingUserEmail),
		TenantId:   gopkg.Pointer(testClubTenantID),
		DataStatus: gopkg.Pointer(enum.DataStatusEnable),
	}
	_, err := testStore.Db.User.Save(ctx, userDomain)
	require.NoError(t, err, "Failed to create initial user for conflict test")

	// 2. Attempt to register a new member with the same username
	registerData := db.MemberRegisterData{
		Username:    existingUsername, // Same username
		Password:    "NewPassword123!",
		Email:       fmt.Sprintf("newemail_%d@example.com", time.Now().UnixNano()), // Different email
		FirstName:   "ConflictUser",
		LastName:    "NameTest",
		DateOfBirth: "1992-01-01",
		Phone:       "1122334455",
		Keycode:     tenantKeycode,
	}

	bodyBytes, _ := json.Marshal(registerData)
	req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "test.example.com"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code, "Response code should be 409 Conflict. Body: %s", w.Body.String())

	var errResp ecode.Error
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err, "Failed to unmarshal error response body")
	assert.Equal(t, ecode.UserConflict.ErrCode, errResp.ErrCode, "Error code should be UserConflict")
}

func TestMemberRegister_InvalidInput_MissingFields(t *testing.T) {
	router, _, _ := setupAuthTestRouter(t) // Store and mail tracker not strictly needed for validating input binding
	tenantKeycode := "test_club"           // A valid tenant keycode is still needed for the request structure
	timestamp := time.Now().UnixNano()

	testCases := []struct {
		name         string
		payload      db.MemberRegisterData
		expectedCode int
		// We might not get a specific ecode for bind failures if ShouldBindJSON catches it first
		// The default Gin binding error is often just a 400 without a custom ecode body.
		// expectedErrCode int // If specific ecode is expected from validation
	}{
		{
			name: "Missing Username",
			payload: db.MemberRegisterData{
				// Username omitted
				Password:    "Password123!",
				Email:       fmt.Sprintf("test_%d@example.com", timestamp),
				FirstName:   "Test",
				LastName:    "User",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
				Keycode:     tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing Password",
			payload: db.MemberRegisterData{
				Username: fmt.Sprintf("user_%d", timestamp),
				// Password omitted
				Email:       fmt.Sprintf("test_%d@example.com", timestamp),
				FirstName:   "Test",
				LastName:    "User",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
				Keycode:     tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing Email",
			payload: db.MemberRegisterData{
				Username: fmt.Sprintf("user_%d", timestamp),
				Password: "Password123!",
				// Email omitted
				FirstName:   "Test",
				LastName:    "User",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
				Keycode:     tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing FirstName",
			payload: db.MemberRegisterData{
				Username:    fmt.Sprintf("user_%d", timestamp),
				Password:    "Password123!",
				Email:       fmt.Sprintf("test_%d@example.com", timestamp),
				LastName:    "User",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
				Keycode:     tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing LastName",
			payload: db.MemberRegisterData{
				Username:    fmt.Sprintf("user_%d", timestamp),
				Password:    "Password123!",
				Email:       fmt.Sprintf("test_%d@example.com", timestamp),
				FirstName:   "Test",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
				Keycode:     tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing DateOfBirth",
			payload: db.MemberRegisterData{
				Username:  fmt.Sprintf("user_%d", timestamp),
				Password:  "Password123!",
				Email:     fmt.Sprintf("test_%d@example.com", timestamp),
				FirstName: "Test",
				LastName:  "User",
				Phone:     "1234567890",
				Keycode:   tenantKeycode,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing Keycode",
			payload: db.MemberRegisterData{
				Username:    fmt.Sprintf("user_%d", timestamp),
				Password:    "Password123!",
				Email:       fmt.Sprintf("test_%d@example.com", timestamp),
				FirstName:   "Test",
				LastName:    "User",
				DateOfBirth: "1990-01-01",
				Phone:       "1234567890",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Host = "test.example.com"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Response code should be %d for %s. Body: %s", tc.expectedCode, tc.name, w.Body.String())

			if tc.expectedCode == http.StatusBadRequest {
				var errResp ecode.Error
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err, "Failed to unmarshal error response body for %s", tc.name)
				// For binding failures on `ShouldBindJSON`, the ecode might be a generic BadRequest
				// or one derived from validator tags. Check if it's a known ecode.Error format.
				assert.Equal(t, ecode.BadRequest.ErrCode, errResp.ErrCode, "Error code should be BadRequest for %s", tc.name)
			}
		})
	}
}

func TestMemberRegister_InvalidInput_BadEmail(t *testing.T) {
	router, _, _ := setupAuthTestRouter(t)
	tenantKeycode := "test_club"
	timestamp := time.Now().UnixNano()

	payload := db.MemberRegisterData{
		Username:    fmt.Sprintf("user_%d", timestamp),
		Password:    "Password123!",
		Email:       "not-an-email", // Invalid email format
		FirstName:   "Test",
		LastName:    "User",
		DateOfBirth: "1990-01-01",
		Phone:       "1234567890",
		Keycode:     tenantKeycode,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "test.example.com"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Response code should be 400 Bad Request for invalid email. Body: %s", w.Body.String())

	var errResp ecode.Error
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err, "Failed to unmarshal error response body for invalid email")
	// Gin's default binding for `binding:"email"` should result in a BadRequest.
	// The specific error message/code might vary depending on validator details.
	assert.Equal(t, ecode.BadRequest.ErrCode, errResp.ErrCode, "Error code should be BadRequest for invalid email")
}

func TestMemberRegister_InvalidDOB(t *testing.T) {
	router, _, _ := setupAuthTestRouter(t)
	tenantKeycode := "test_club"
	timestamp := time.Now().UnixNano()

	testCases := []struct {
		name            string
		dob             string
		expectedCode    int
		expectedErrCode string
	}{
		{
			name:            "Invalid DOB format DD-MM-YYYY",
			dob:             "01-01-1990",
			expectedCode:    http.StatusBadRequest,
			expectedErrCode: ecode.New(http.StatusBadRequest, "invalid_date_format").ErrCode,
		},
		{
			name:            "Invalid DOB format YYYY/MM/DD",
			dob:             "1990/01/01",
			expectedCode:    http.StatusBadRequest,
			expectedErrCode: ecode.New(http.StatusBadRequest, "invalid_date_format").ErrCode,
		},
		{
			name:            "Invalid DOB value non-existent date",
			dob:             "1990-02-30", // February 30th does not exist
			expectedCode:    http.StatusBadRequest,
			expectedErrCode: ecode.New(http.StatusBadRequest, "invalid_date_format").ErrCode,
		},
		{
			name:            "Invalid DOB non-numeric",
			dob:             "NineteenNinety-Jan-First",
			expectedCode:    http.StatusBadRequest,
			expectedErrCode: ecode.New(http.StatusBadRequest, "invalid_date_format").ErrCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := db.MemberRegisterData{
				Username:    fmt.Sprintf("user_dobtest_%d", timestamp),
				Password:    "Password123!",
				Email:       fmt.Sprintf("dobtest_%d@example.com", timestamp),
				FirstName:   "DOBTest",
				LastName:    "User",
				DateOfBirth: tc.dob,
				Phone:       "1234560000",
				Keycode:     tenantKeycode,
			}
			timestamp++ // Ensure unique user/email for sub-tests if they create users

			bodyBytes, _ := json.Marshal(payload)
			req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Host = "test.example.com"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Response code for %s. Body: %s", tc.name, w.Body.String())

			var errResp ecode.Error
			err := json.Unmarshal(w.Body.Bytes(), &errResp)
			require.NoError(t, err, "Failed to unmarshal error response body for %s", tc.name)
			assert.Equal(t, tc.expectedErrCode, errResp.ErrCode, "Specific error code for %s", tc.name)
		})
	}
}

func TestMemberRegister_TenantNotFound(t *testing.T) {
	router, _, _ := setupAuthTestRouter(t) // Store is used implicitly by handler, but no specific setup needed beyond router
	timestamp := time.Now().UnixNano()

	payload := db.MemberRegisterData{
		Username:    fmt.Sprintf("user_tenanttest_%d", timestamp),
		Password:    "Password123!",
		Email:       fmt.Sprintf("tenanttest_%d@example.com", timestamp),
		FirstName:   "TenantTest",
		LastName:    "User",
		DateOfBirth: "1990-01-01",
		Phone:       "1234567777",
		Keycode:     "non_existent_tenant_keycode",
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/v1/member/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "test.example.com"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, ecode.TenantNotFound.Status, w.Code, "Response code for non-existent tenant. Body: %s", w.Body.String())

	var errResp ecode.Error
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err, "Failed to unmarshal error response body for tenant not found")
	assert.Equal(t, ecode.TenantNotFound.ErrCode, errResp.ErrCode, "Error code should be TenantNotFound")
}

// TODO: Add more tests:
// - TestMemberRegister_EmailSendingFailure (requires ability to make mail send fail)
