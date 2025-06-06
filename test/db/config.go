package testdb

import (
	"app/pkg/enum"
	"app/store/db"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Default test database connection details
	defaultTestMongoURI      = "mongodb://localhost:27017"
	defaultTestMongoDatabase = "cannabis_test"
	testMongoURIEnv          = "TEST_MONGODB_URI"
	testMongoDatabaseEnv     = "TEST_MONGODB_DATABASE"
)

var (
	testClient *mongo.Client
	testDB     *mongo.Database
)

// GetTestMongoURI returns the MongoDB URI for the test environment
func GetTestMongoURI() string {
	uri := os.Getenv(testMongoURIEnv)
	if uri == "" {
		return defaultTestMongoURI
	}
	return uri
}

// GetTestMongoDatabase returns the MongoDB database name for the test environment
func GetTestMongoDatabase() string {
	dbName := os.Getenv(testMongoDatabaseEnv)
	if dbName == "" {
		return defaultTestMongoDatabase
	}
	return dbName
}

// Setup initializes the test database connection
// This should be called in TestMain functions
func Setup() {
	uri := GetTestMongoURI()
	dbName := GetTestMongoDatabase()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		logrus.Fatalf("Failed to connect to test MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logrus.Fatalf("Failed to ping test MongoDB: %v", err)
	}

	testClient = client
	testDB = client.Database(dbName)

	logrus.Infof("Connected to test MongoDB at %s using database %s", uri, dbName)
}

// Teardown closes the test database connection
// This should be called at the end of TestMain functions
func Teardown() {
	if testClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := testClient.Disconnect(ctx); err != nil {
			logrus.Errorf("Failed to disconnect from test MongoDB: %v", err)
		}
	}
}

// ResetCollection drops and recreates a collection for testing
func ResetCollection(t *testing.T, collectionName string) {
	if testDB == nil {
		t.Fatalf("Test database not initialized. Call Setup() first.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testDB.Collection(collectionName).Drop(ctx); err != nil {
		// Ignore error if collection doesn't exist
		if err.Error() != "ns not found" {
			t.Logf("Warning while dropping collection %s: %v", collectionName, err)
		}
	}
}

// InsertTestData inserts test data into the specified collection
func InsertTestData(t *testing.T, collectionName string, data interface{}) (string, error) {
	if testDB == nil {
		t.Fatalf("Test database not initialized. Call Setup() first.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := testDB.Collection(collectionName).InsertOne(ctx, data)
	if err != nil {
		return "", fmt.Errorf("failed to insert test data: %w", err)
	}

	// Convert inserted ID to string
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return fmt.Sprintf("%v", result.InsertedID), nil
}

// GetTestDBContext returns a context and database for testing
func GetTestDBContext() (context.Context, *mongo.Database) {
	if testDB == nil {
		logrus.Fatal("Test database not initialized. Call Setup() first.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ctx, testDB
}

// CreateTestTenant creates a test tenant for testing
func CreateTestTenant() (enum.Tenant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := primitive.NewObjectID()
	tenant := &db.TenantDomain{
		BaseDomain: db.BaseDomain{
			ID: id,
		},
		Name:       gopkg.Pointer("Test Cannabis Club"),
		Keycode:    gopkg.Pointer("test_club"),
		Username:   gopkg.Pointer("test_admin"),
		DataStatus: gopkg.Pointer(enum.DataStatusEnable),
		IsRoot:     gopkg.Pointer(false),
	}

	_, err := testDB.Collection("tenant").InsertOne(ctx, tenant)
	if err != nil {
		return "", fmt.Errorf("failed to create test tenant: %w", err)
	}

	return enum.Tenant(id.Hex()), nil
}
