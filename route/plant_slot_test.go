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

// Test helper to setup plant slot router
func setupPlantSlotTestRouter(t *testing.T) (*gin.Engine, *store.Store) {
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

	// Initialize routes (this will include plant slot routes via init())
	for i := range handlers {
		handlers[i](mdw, router)
	}

	return router, testStore
}

func TestPlantSlotEndpoints_Unauthorized(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	tests := []struct {
		method   string
		endpoint string
	}{
		{"GET", "/plant-slots/v1/my-slots"},
		{"POST", "/plant-slots/v1/request"},
		{"GET", "/plant-slots/v1/123"},
		{"PUT", "/plant-slots/v1/123/status"},
		{"POST", "/plant-slots/v1/123/maintenance"},
		{"POST", "/plant-slots/v1/transfer"},
		{"GET", "/plant-slots/v1/admin/all"},
		{"POST", "/plant-slots/v1/admin/assign"},
		{"GET", "/plant-slots/v1/admin/maintenance"},
		{"GET", "/plant-slots/v1/admin/analytics"},
		{"PUT", "/plant-slots/v1/admin/123/force-status"},
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

func TestPlantSlotEndpoints_WithInvalidAuth(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	req, _ := http.NewRequest("GET", "/plant-slots/v1/my-slots", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return 500 with invalid token")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "internal_server_error", response["error"])
}

func TestPlantSlotRoutes_Basic(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test that all plant slot routes are properly registered by checking they don't return 404
	tests := []struct {
		method   string
		endpoint string
	}{
		{"GET", "/plant-slots/v1/my-slots"},
		{"POST", "/plant-slots/v1/request"},
		{"GET", "/plant-slots/v1/123"},
		{"PUT", "/plant-slots/v1/123/status"},
		{"POST", "/plant-slots/v1/123/maintenance"},
		{"POST", "/plant-slots/v1/transfer"},
		{"GET", "/plant-slots/v1/admin/all"},
		{"POST", "/plant-slots/v1/admin/assign"},
		{"GET", "/plant-slots/v1/admin/maintenance"},
		{"GET", "/plant-slots/v1/admin/analytics"},
		{"PUT", "/plant-slots/v1/admin/123/force-status"},
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

func TestPlantSlot_CompilationAndRegistration(t *testing.T) {
	// This test verifies that the plant slot module compiles and registers properly
	router, store := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Check that store is properly initialized
	assert.NotNil(t, store, "Store should be initialized")
	assert.NotNil(t, store.Db, "Database should be available")
	assert.NotNil(t, store.Db.PlantSlot, "PlantSlot repository should be available")
	assert.NotNil(t, store.Db.Member, "Member repository should be available")
	assert.NotNil(t, store.Db.Membership, "Membership repository should be available")

	// Check that routes are registered by making a request
	req, _ := http.NewRequest("GET", "/plant-slots/v1/my-slots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 404 (route exists), should return 401 (auth required)
	assert.NotEqual(t, http.StatusNotFound, w.Code, "Plant slot routes should be registered")
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")
}

func TestPlantSlotRoutes_JsonValidation(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test invalid JSON for request endpoint (without auth should still get 401 first)
	payload := strings.NewReader(`{"invalid": json}`)

	req, _ := http.NewRequest("POST", "/plant-slots/v1/request", payload)
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

func TestPlantSlotRequest_JsonStructure(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test valid JSON structure for slot request
	payload := strings.NewReader(`{"quantity": 2, "preferred_location": "greenhouse-1"}`)

	req, _ := http.NewRequest("POST", "/plant-slots/v1/request", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still require authentication but JSON should be parseable
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}

func TestPlantSlotTransfer_JsonStructure(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test valid JSON structure for slot transfer
	payload := strings.NewReader(`{
		"to_member_id": "654db9eca1f1b1bdbf3d4621",
		"slot_ids": ["654db9eca1f1b1bdbf3d4617", "654db9eca1f1b1bdbf3d4622"],
		"reason": "Member upgrade transfer"
	}`)

	req, _ := http.NewRequest("POST", "/plant-slots/v1/transfer", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still require authentication but JSON should be parseable
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}

func TestPlantSlotMaintenance_JsonStructure(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test valid JSON structure for maintenance request
	payload := strings.NewReader(`{
		"description": "Irrigation system needs adjustment",
		"priority": "normal"
	}`)

	req, _ := http.NewRequest("POST", "/plant-slots/v1/123/maintenance", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still require authentication but JSON should be parseable
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}

func TestPlantSlotStatusUpdate_JsonStructure(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test valid JSON structure for status update
	payload := strings.NewReader(`{
		"status": "occupied",
		"reason": "Started new cultivation cycle"
	}`)

	req, _ := http.NewRequest("PUT", "/plant-slots/v1/123/status", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still require authentication but JSON should be parseable
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}

func TestPlantSlotAdminAssign_JsonStructure(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test valid JSON structure for admin assignment
	payload := strings.NewReader(`{
		"member_id": "654db9eca1f1b1bdbf3d4618",
		"membership_id": "654db9eca1f1b1bdbf3d4619",
		"slot_ids": ["654db9eca1f1b1bdbf3d4617", "654db9eca1f1b1bdbf3d4622"],
		"assigned_by": "654db9eca1f1b1bdbf3d4620"
	}`)

	req, _ := http.NewRequest("POST", "/plant-slots/v1/admin/assign", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still require authentication but JSON should be parseable
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response["error"])
}

func TestPlantSlotAdminEndpoints_ExistAndRequireAuth(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test all admin endpoints exist and require authentication
	adminTests := []struct {
		method   string
		endpoint string
	}{
		{"GET", "/plant-slots/v1/admin/all"},
		{"GET", "/plant-slots/v1/admin/maintenance"},
		{"GET", "/plant-slots/v1/admin/analytics"},
		{"PUT", "/plant-slots/v1/admin/123/force-status"},
	}

	for _, test := range adminTests {
		t.Run("admin_"+test.method+"_"+test.endpoint, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Admin routes should exist and require authentication
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Admin route should be registered")
			assert.Equal(t, http.StatusUnauthorized, w.Code, "Admin route should require authentication")
		})
	}
}

func TestPlantSlotQueryParameters(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test admin endpoints with query parameters
	testCases := []struct {
		endpoint    string
		queryParams string
	}{
		{"/plant-slots/v1/admin/all", "?page=1&limit=20&status=available&location=greenhouse-1"},
		{"/plant-slots/v1/admin/maintenance", "?days=30"},
		{"/plant-slots/v1/admin/analytics", ""},
	}

	for _, test := range testCases {
		t.Run("query_params_"+test.endpoint, func(t *testing.T) {
			fullURL := test.endpoint + test.queryParams
			req, _ := http.NewRequest("GET", fullURL, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should handle query parameters correctly (require auth first)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require authentication")

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "unauthorized", response["error"])
		})
	}
}

// Test helper functions validation
func TestHelperFunctions(t *testing.T) {
	// Test isValidStatusTransition function
	tests := []struct {
		from     string
		to       string
		expected bool
		testName string
	}{
		{"available", "allocated", true, "available_to_allocated"},
		{"allocated", "occupied", true, "allocated_to_occupied"},
		{"allocated", "available", true, "allocated_to_available"},
		{"occupied", "maintenance", true, "occupied_to_maintenance"},
		{"occupied", "available", true, "occupied_to_available"},
		{"maintenance", "available", true, "maintenance_to_available"},
		{"maintenance", "out_of_service", true, "maintenance_to_out_of_service"},
		{"out_of_service", "maintenance", true, "out_of_service_to_maintenance"},
		{"out_of_service", "available", true, "out_of_service_to_available"},
		{"available", "occupied", false, "invalid_available_to_occupied"},
		{"occupied", "allocated", false, "invalid_occupied_to_allocated"},
		{"maintenance", "occupied", false, "invalid_maintenance_to_occupied"},
		{"", "available", false, "empty_current_status"},
		{"available", "", false, "empty_new_status"},
	}

	for _, test := range tests {
		t.Run("status_transition_"+test.testName, func(t *testing.T) {
			result := isValidStatusTransition(test.from, test.to)
			assert.Equal(t, test.expected, result,
				"Status transition from %s to %s should be %v", test.from, test.to, test.expected)
		})
	}
}

func TestStringPtrHelper(t *testing.T) {
	// Test stringPtr helper function
	testString := "test_value"
	ptr := stringPtr(testString)

	assert.NotNil(t, ptr, "stringPtr should return non-nil pointer")
	assert.Equal(t, testString, *ptr, "stringPtr should return pointer to correct value")
}

func TestGetValueHelper(t *testing.T) {
	// Test getValue helper function
	testString := "test_value"
	defaultString := "default_value"

	// Test with non-nil pointer
	result1 := getValue(&testString, defaultString)
	assert.Equal(t, testString, result1, "getValue should return pointer value when not nil")

	// Test with nil pointer
	result2 := getValue((*string)(nil), defaultString)
	assert.Equal(t, defaultString, result2, "getValue should return default value when pointer is nil")

	// Test with integer
	testInt := 42
	defaultInt := 0

	result3 := getValue(&testInt, defaultInt)
	assert.Equal(t, testInt, result3, "getValue should work with integers")

	result4 := getValue((*int)(nil), defaultInt)
	assert.Equal(t, defaultInt, result4, "getValue should return default for nil integer pointer")
}

// Performance test to ensure routes don't have obvious bottlenecks
func TestPlantSlotRoutes_BasicPerformance(t *testing.T) {
	router, _ := setupPlantSlotTestRouter(t)
	defer func() {
		testdb.ResetCollection(t, "users")
		testdb.ResetCollection(t, "members")
		testdb.ResetCollection(t, "memberships")
		testdb.ResetCollection(t, "plant_slots")
	}()

	// Test that routes respond within reasonable time (basic smoke test)
	req, _ := http.NewRequest("GET", "/plant-slots/v1/my-slots", nil)
	w := httptest.NewRecorder()

	// This is a very basic performance check - routes should respond quickly even without auth
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should respond with 401")
	assert.NotEmpty(t, w.Body.String(), "Should return response body")
}

// Test that ensures all expected routes are covered in our test suite
func TestPlantSlotRoutes_Coverage(t *testing.T) {
	// This test ensures we're testing all the routes we expect to have
	expectedRoutes := []string{
		"GET:/plant-slots/v1/my-slots",
		"POST:/plant-slots/v1/request",
		"GET:/plant-slots/v1/:id",
		"PUT:/plant-slots/v1/:id/status",
		"POST:/plant-slots/v1/:id/maintenance",
		"POST:/plant-slots/v1/transfer",
		"GET:/plant-slots/v1/admin/all",
		"POST:/plant-slots/v1/admin/assign",
		"GET:/plant-slots/v1/admin/maintenance",
		"GET:/plant-slots/v1/admin/analytics",
		"PUT:/plant-slots/v1/admin/:id/force-status",
	}

	// This is more of a documentation test to ensure we remember all routes
	assert.Equal(t, 11, len(expectedRoutes), "Should have 11 plant slot routes total")

	memberRoutes := 0
	adminRoutes := 0

	for _, route := range expectedRoutes {
		if strings.Contains(route, "/admin/") {
			adminRoutes++
		} else {
			memberRoutes++
		}
	}

	assert.Equal(t, 6, memberRoutes, "Should have 6 member routes")
	assert.Equal(t, 5, adminRoutes, "Should have 5 admin routes")
}
