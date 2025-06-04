package testhelpers

import (
	"app/pkg/enum"
	"app/store/db"
	testdb "app/test/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestContext provides common context values used in tests
type TestContext struct {
	Context    context.Context
	Database   *mongo.Database
	TenantID   enum.Tenant
	UserID     string
	RoleID     string
	Collection string
	T          *testing.T
}

// SetupTest initializes a test context with all necessary values
func SetupTest(t *testing.T, collection string) *TestContext {
	testdb.Setup()

	// Create a test tenant
	tenantID, err := testdb.CreateTestTenant()
	require.NoError(t, err, "Failed to create test tenant")

	ctx, testDB := testdb.GetTestDBContext()
	testdb.ResetCollection(t, collection)

	return &TestContext{
		Context:    ctx,
		Database:   testDB,
		TenantID:   tenantID,
		Collection: collection,
		T:          t,
	}
}

// CleanupTest performs cleanup after a test completes
func CleanupTest(t *testing.T) {
	testdb.ResetCollection(t, "tenant")
	testdb.Teardown()
}

// LoadTestData loads test data from a JSON file
func LoadTestData(t *testing.T, filename string, data interface{}) {
	path := filepath.Join("testdata", filename)
	bytes, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read test data file")

	err = json.Unmarshal(bytes, data)
	require.NoError(t, err, "Failed to unmarshal test data")
}

// InsertTestDocument inserts a document into the test collection
func (tc *TestContext) InsertTestDocument(data interface{}) string {
	id, err := testdb.InsertTestData(tc.T, tc.Collection, data)
	require.NoError(tc.T, err)
	return id
}

// CreateTestGinContext creates a test Gin context with the specified method and path
func CreateTestGinContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create HTTP request
	var req *http.Request
	if body != nil {
		bodyJSON, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, strings.NewReader(string(bodyJSON)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	c.Request = req
	return c, w
}

// SetupAuthContext adds authentication context for testing authenticated endpoints
func SetupAuthContext(c *gin.Context, tenantID enum.Tenant, userID, roleID string) {
	c.Set("tenant_id", tenantID)
	c.Set("user_id", userID)
	c.Set("role_id", roleID)
}

// CreateTestMember creates a test member in the database
func (tc *TestContext) CreateTestMember() (string, *db.MemberDomain) {
	id := primitive.NewObjectID()
	now := time.Now()

	member := &db.MemberDomain{
		BaseDomain: db.BaseDomain{
			ID:        id,
			CreatedAt: gopkg.Pointer(now),
			UpdatedAt: gopkg.Pointer(now),
			CreatedBy: gopkg.Pointer("test_user"),
			UpdatedBy: gopkg.Pointer("test_user"),
		},
		FirstName:    gopkg.Pointer("Test"),
		LastName:     gopkg.Pointer("Member"),
		Email:        gopkg.Pointer("test@example.com"),
		MemberStatus: gopkg.Pointer("active"),
		Phone:        gopkg.Pointer("555-1234"),
		TenantId:     gopkg.Pointer(tc.TenantID),
	}

	_, err := tc.Database.Collection("member").InsertOne(tc.Context, member)
	require.NoError(tc.T, err, "Failed to create test member")

	return id.Hex(), member
}

// CreateTestPlant creates a test plant in the database
func (tc *TestContext) CreateTestPlant(memberID string) (string, *db.PlantDomain) {
	id := primitive.NewObjectID()
	now := time.Now()

	plant := &db.PlantDomain{
		BaseDomain: db.BaseDomain{
			ID:        id,
			CreatedAt: gopkg.Pointer(now),
			UpdatedAt: gopkg.Pointer(now),
			CreatedBy: gopkg.Pointer("test_user"),
			UpdatedBy: gopkg.Pointer("test_user"),
		},
		Name:        gopkg.Pointer("Test Plant"),
		MemberID:    gopkg.Pointer(memberID),
		PlantTypeID: gopkg.Pointer(primitive.NewObjectID().Hex()),
		PlantSlotID: gopkg.Pointer(primitive.NewObjectID().Hex()),
		Status:      gopkg.Pointer(string(enum.PlantStatusGrowing)),
		PlantedDate: gopkg.Pointer(now),
		TenantId:    gopkg.Pointer(tc.TenantID),
	}

	_, err := tc.Database.Collection("plant").InsertOne(tc.Context, plant)
	require.NoError(tc.T, err, "Failed to create test plant")

	return id.Hex(), plant
}

// GetDocumentByID retrieves a document by its ID
func (tc *TestContext) GetDocumentByID(id string, result interface{}) {
	objectID, err := primitive.ObjectIDFromHex(id)
	require.NoError(tc.T, err, "Invalid object ID")

	err = tc.Database.Collection(tc.Collection).FindOne(tc.Context, bson.M{"_id": objectID}).Decode(result)
	require.NoError(tc.T, err, fmt.Sprintf("Failed to get document with ID %s", id))
}

// AssertCollectionCount checks that a collection has the expected number of documents
func (tc *TestContext) AssertCollectionCount(filter interface{}, expectedCount int) {
	if filter == nil {
		filter = bson.M{}
	}

	count, err := tc.Database.Collection(tc.Collection).CountDocuments(tc.Context, filter)
	require.NoError(tc.T, err, "Failed to count documents")
	require.Equal(tc.T, int64(expectedCount), count, "Collection count doesn't match expected")
}

// AssertDocumentExists verifies that a document with the given filter exists
func (tc *TestContext) AssertDocumentExists(filter interface{}) {
	count, err := tc.Database.Collection(tc.Collection).CountDocuments(tc.Context, filter)
	require.NoError(tc.T, err, "Failed to check document existence")
	require.Equal(tc.T, int64(1), count, "Document doesn't exist")
}

// AssertDocumentNotExists verifies that a document with the given filter doesn't exist
func (tc *TestContext) AssertDocumentNotExists(filter interface{}) {
	count, err := tc.Database.Collection(tc.Collection).CountDocuments(tc.Context, filter)
	require.NoError(tc.T, err, "Failed to check document existence")
	require.Equal(tc.T, int64(0), count, "Document exists when it shouldn't")
}
