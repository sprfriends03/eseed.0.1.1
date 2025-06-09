package route

import (
	"app/env"
	"app/store"
	testdb "app/test/db"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper to setup membership router
func setupMembershipTestRouter(t *testing.T) (*gin.Engine, *store.Store) {
	ctx := context.Background()

	testStore := store.New()
	require.NotNil(t, testStore, "Test store should not be nil")

	_, testMongoDb := testdb.GetTestDBContext()
	require.NotNil(t, testMongoDb, "Test MongoDB instance should not be nil")

	// Redis connection
	opts, err := redis.ParseURL(env.RedisUri)
	require.NoError(t, err, "Failed to parse Redis URI")

	rdbClient := redis.NewClient(opts)
	_, errRedis := rdbClient.Ping(ctx).Result()
	require.NoError(t, errRedis, "Failed to connect to test Redis instance")

	// Create middleware
	mdw := newMdw(testStore)

	router := gin.New()
	router.Use(mdw.Error())

	// Initialize routes (this will include membership routes via init())
	for i := range handlers {
		handlers[i](mdw, router)
	}

	return router, testStore
}

func TestMembershipEndpoints_Unauthorized(t *testing.T) {
	router, _ := setupMembershipTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
	}()

	tests := []struct {
		method   string
		endpoint string
	}{
		{"POST", "/membership/v1/purchase"},
		{"GET", "/membership/v1/status"},
		{"POST", "/membership/v1/renew"},
		{"GET", "/membership/v1/history"},
		{"DELETE", "/membership/v1/123"},
		{"GET", "/membership/v1/admin/pending"},
		{"GET", "/membership/v1/admin/expiring"},
		{"PUT", "/membership/v1/admin/123/status"},
		{"GET", "/membership/v1/admin/analytics"},
	}

	for _, test := range tests {
		t.Run(test.method+"_"+test.endpoint, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 without auth")

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "unauthorized", response["error"])
		})
	}
}

func TestMembershipEndpoints_WithInvalidAuth(t *testing.T) {
	router, _ := setupMembershipTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
	}()

	req, _ := http.NewRequest("GET", "/membership/v1/status", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return 500 with invalid token")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "internal_server_error", response["error"])
}

func TestMembershipRoutes_Basic(t *testing.T) {
	router, _ := setupMembershipTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
	}()

	// Test that all membership routes are properly registered by checking they don't return 404
	tests := []struct {
		method   string
		endpoint string
	}{
		{"POST", "/membership/v1/purchase"},
		{"GET", "/membership/v1/status"},
		{"POST", "/membership/v1/renew"},
		{"GET", "/membership/v1/history"},
		{"DELETE", "/membership/v1/123"},
		{"GET", "/membership/v1/admin/pending"},
		{"GET", "/membership/v1/admin/expiring"},
		{"PUT", "/membership/v1/admin/123/status"},
		{"GET", "/membership/v1/admin/analytics"},
	}

	for _, test := range tests {
		t.Run(test.method+"_"+test.endpoint+"_registered", func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Routes should be registered (not 404) and require auth (401)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should be registered")
			assert.Equal(t, http.StatusUnauthorized, w.Code, "Route should require authentication")
		})
	}
}

func TestMembership_CompilationAndRegistration(t *testing.T) {
	// This test verifies that the membership module compiles and registers properly
	router, store := setupMembershipTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
	}()

	// Check that store is properly initialized
	assert.NotNil(t, store, "Store should be initialized")
	assert.NotNil(t, store.Db, "Database should be available")
	assert.NotNil(t, store.Db.Membership, "Membership repository should be available")
	assert.NotNil(t, store.Db.Member, "Member repository should be available")

	// Check that routes are registered by making a request
	req, _ := http.NewRequest("GET", "/membership/v1/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 404 (route exists), should return 401 (auth required)
	assert.NotEqual(t, http.StatusNotFound, w.Code, "Membership routes should be registered")
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")
}

func TestMembershipRoutes_JsonValidation(t *testing.T) {
	router, _ := setupMembershipTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
	}()

	// Test invalid JSON for purchase endpoint (without auth should still get 401 first)
	payload := strings.NewReader(`{"invalid": json}`)

	req, _ := http.NewRequest("POST", "/membership/v1/purchase", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should get unauthorized since no auth token provided
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication first")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}
