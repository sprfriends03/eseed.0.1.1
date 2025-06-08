package db

import (
	"context"
	"strings"
	"time"

	"app/env"
	"app/pkg/encryption"
	"app/pkg/enum"
	"app/pkg/util"

	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"errors" // Moved import "errors" here
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// --- Repository Interfaces ---

type ClientRepo interface {
	FindOneByClientId(ctx context.Context, clientId string) (*ClientDomain, error)
	FindAllByTenant(ctx context.Context, tenant enum.Tenant) ([]*ClientDomain, error)
	Save(ctx context.Context, domain *ClientDomain, opts ...*options.UpdateOptions) (*ClientDomain, error)
	DeleteOne(ctx context.Context, domain *ClientDomain) error
	Count(ctx context.Context, q *ClientQuery, opts ...*options.CountOptions) int64
	FindAll(ctx context.Context, q *ClientQuery, opts ...*options.FindOptions) ([]*ClientDomain, error)
	CollectionName() string
}

type RoleRepo interface {
	FindAllByIds(ctx context.Context, ids []string) ([]*RoleDomain, error)
	// Adding methods that might be used elsewhere or for completeness, based on db/role.go
	Save(ctx context.Context, domain *RoleDomain, opts ...*options.UpdateOptions) (*RoleDomain, error)
	FindOneById(ctx context.Context, id string) (*RoleDomain, error)
	Count(ctx context.Context, q *RoleQuery, opts ...*options.CountOptions) int64
	FindAll(ctx context.Context, q *RoleQuery, opts ...*options.FindOptions) ([]*RoleDomain, error)
	CollectionName() string
}

type TenantRepo interface {
	FindOneById(ctx context.Context, id string) (*TenantDomain, error)
	Save(ctx context.Context, domain *TenantDomain, opts ...*options.UpdateOptions) (*TenantDomain, error)
	Count(ctx context.Context, q *TenantQuery, opts ...*options.CountOptions) int64
	FindAll(ctx context.Context, q *TenantQuery, opts ...*options.FindOptions) ([]*TenantDomain, error)
	CollectionName() string
	FindOneByKeycode(ctx context.Context, keycode string) (*TenantDomain, error) // Added from db/tenant.go
}

type UserRepo interface {
	FindOneById(ctx context.Context, id string) (*UserDomain, error)
	FindOneByTenant_Username(ctx context.Context, tenant enum.Tenant, username string) (*UserDomain, error)
	FindAllByTenant(ctx context.Context, tenant enum.Tenant) ([]*UserDomain, error)
	Save(ctx context.Context, domain *UserDomain, opts ...*options.UpdateOptions) (*UserDomain, error)
	CollectionName() string
	// Adding methods that might be used elsewhere or for completeness, based on db/user.go
	UpdateOne(ctx context.Context, filter M, update M, opts ...*options.UpdateOptions) error
	Count(ctx context.Context, q *UserQuery, opts ...*options.CountOptions) int64
	FindAll(ctx context.Context, q *UserQuery, opts ...*options.FindOptions) ([]*UserDomain, error)
	FindAllByRole(ctx context.Context, roleId string) ([]*UserDomain, error)
	FindOneByEmailVerificationToken(ctx context.Context, token string) (*UserDomain, error)
	IncrementVersionToken(ctx context.Context, id string) error
}

// --- DB Struct and Methods ---

var _db *DB

// Define collections
const (
	userCollection         = "users"
	roleCollection         = "roles"
	memberCollection       = "members"
	membershipCollection   = "memberships"
	plantTypeCollection    = "plant_types"
	plantSlotCollection    = "plant_slots"
	plantCollection        = "plants"
	careRecordCollection   = "care_records"
	harvestCollection      = "harvests"
	notificationCollection = "notifications"
	tenantCollection       = "tenant"
	clientCollection       = "client"
	auditLogCollection     = "audit_log"
)

// DB represents the database layer
type DB struct {
	// Store *mongo.Database instead of context and client,
	// as individual repos will use collections from this database.
	db *mongo.Database

	client ClientRepo // Renamed from Client
	role   RoleRepo   // Renamed from Role
	tenant TenantRepo // Renamed from Tenant
	user   UserRepo   // Renamed from User
	Member       *member
	Membership   *membership
	PlantType    *plantType
	PlantSlot    *plantSlot
	Plant        *plant
	CareRecord   *careRecord
	Harvest      *harvest
	Notification *notification
	AuditLog     *audit_log
	// Removed ctx and client fields, assuming they are managed within Init or by individual repos through *mongo.Database

	mongoClient *mongo.Client // Keep a reference to the client for closing
}

// Create a MongoDB client
func createMongoClient(ctx context.Context) (*mongo.Client, error) {
	uri := env.MongoUri
	connectTimeout := 10 * time.Second

	if uri == "" {
		logrus.Fatal("MongoDB URI is empty - check env configuration")
		// For testability, perhaps return an error instead of Fatal
		// return nil, errors.New("MongoDB URI is empty")
	}

	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(connectTimeout).
		SetReadConcern(readconcern.Majority()).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		logrus.WithError(err).Errorln("Failed to connect to MongoDB") // Error instead of Fatal
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logrus.WithError(err).Errorln("Failed to ping MongoDB") // Error instead of Fatal
		return nil, err
	}

	logrus.Infoln("Connected to MongoDB at", uri)
	return client, nil
}

// Init initializes the database connections
func Init(ctx context.Context) *DB {
	mongoCli, err := createMongoClient(ctx)
	if err != nil {
		// Depending on application's desired behavior, panic, log fatal, or handle error
		logrus.WithError(err).Fatalln("Database initialization failed during mongo client creation")
		return nil // Or panic
	}

	// Extract database name from the URI
	uri := env.MongoUri
	dbName := "app" // default fallback
	if uri != "" {
		if idx := strings.LastIndex(uri, "/"); idx != -1 && idx < len(uri)-1 {
			afterSlash := uri[idx+1:]
			if qIdx := strings.Index(afterSlash, "?"); qIdx != -1 {
				dbName = afterSlash[:qIdx]
			} else if afterSlash != "" {
				dbName = afterSlash
			}
		}
	}

	logrus.Infoln("Using database name:", dbName, "extracted from URI:", uri)
	mongoDb := mongoCli.Database(dbName)

	// Keep other repos as they are for now if not specified in subtask
	// but ideally, they would also be interfaces.
	_db = &DB{
		db:           mongoDb,
		mongoClient:  mongoCli,
		user:         newUser(ctx, mongoDb.Collection(userCollection)),     // Use new field name
		role:         newRole(ctx, mongoDb.Collection(roleCollection)),     // Use new field name
		tenant:       newTenant(ctx, mongoDb.Collection(tenantCollection)), // Use new field name
		client:       newClient(ctx, mongoDb.Collection(clientCollection)), // Use new field name
		Member:       newMember(ctx, mongoDb.Collection(memberCollection)),
		Membership:   newMembership(ctx, mongoDb.Collection(membershipCollection)),
		PlantType:    newPlantType(ctx, mongoDb.Collection(plantTypeCollection)),
		PlantSlot:    newPlantSlot(ctx, mongoDb.Collection(plantSlotCollection)),
		Plant:        newPlant(ctx, mongoDb.Collection(plantCollection)),
		CareRecord:   newCareRecord(ctx, mongoDb.Collection(careRecordCollection)),
		Harvest:      newHarvest(ctx, mongoDb.Collection(harvestCollection)),
		Notification: newNotification(ctx, mongoDb.Collection(notificationCollection)),
		AuditLog:     newAuditLog(ctx, mongoDb.Collection(auditLogCollection)),
	}

	// Initialize root tenant and client
	// This uses the repository methods, which should be fine as they are assigned above.
	_db.initialize(ctx)

	return _db
}

// Get returns the database instance
func Get() *DB {
	return _db
}

// Close closes the database connection
func (d *DB) Close(ctx context.Context) { // Added ctx parameter
	if d.mongoClient != nil {
		if err := d.mongoClient.Disconnect(ctx); err != nil { // Use passed ctx
			logrus.WithError(err).Warningln("Error disconnecting from MongoDB")
		}
	}
}

// Accessor methods for DBer interface
func (d *DB) Client() ClientRepo     { return d.client } // Use new field name
func (d *DB) Role() RoleRepo         { return d.role }   // Use new field name
func (d *DB) Tenant() TenantRepo     { return d.tenant } // Use new field name
func (d *DB) User() UserRepo         { return d.user }   // Use new field name
func (d *DB) MongoInstance() *mongo.Client { return d.mongoClient }
func (d *DB) Database() *mongo.Database { return d.db }


// SID converts an ObjectID to string
func SID(id primitive.ObjectID) string {
	return id.Hex()
}

// OID converts a string to an ObjectID
func OID(id string) primitive.ObjectID {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID // Consider returning error as well
	}
	return objID
}

// Regex creates a case-insensitive regex pattern for MongoDB queries
func Regex(s string) primitive.Regex {
	return primitive.Regex{
		Pattern: ".*" + s + ".*",
		Options: "i",
	}
}

// Uri, Instance, Session methods might need to be re-evaluated based on new DB struct.
// For now, let's assume they might be used by other packages directly on _db.
// If they are part of an interface DB should satisfy, they need to be correct.
// Uri can remain the same. Instance now returns d.mongoClient. Session from d.mongoClient.

func (d *DB) Uri() string { // Changed receiver to pointer
	return env.MongoUri
}

func (d *DB) Instance() *mongo.Client { // Changed receiver to pointer
	return d.mongoClient
}

func (d *DB) Session() (mongo.Session, error) { // Changed receiver to pointer
	if d.mongoClient == nil {
		// Handle case where client is not initialized, though Init should prevent this.
		return nil, errors.New("mongo client not initialized")
	}
	return d.mongoClient.StartSession()
}


func (s *DB) initialize(ctx context.Context) *DB {
	// Ensure that s.Tenant, s.Client, s.User are not nil, which they shouldn't be after Init.
	// This method is called from Init, so the repo fields (like s.Tenant) are already populated.
	tenantDomain := &TenantDomain{ // Renamed variable to avoid conflict with field
		Name:       gopkg.Pointer(env.RootUser),
		Keycode:    gopkg.Pointer(env.RootUser),
		Username:   gopkg.Pointer(env.RootUser),
		DataStatus: gopkg.Pointer(enum.DataStatusEnable),
		IsRoot:     gopkg.Pointer(true),
	}
	// Use s.tenant which is TenantRepo interface.
	savedTenant, err := s.tenant.Save(ctx, tenantDomain) // Use new field name
	if err != nil {
		logrus.WithError(err).Errorln("Failed to save root tenant during initialization")
		return s
	}

	if !savedTenant.ID.IsZero() {
		_, err = s.client.Save(ctx, &ClientDomain{ // Use new field name
			Name:         gopkg.Pointer(env.RootUser),
			ClientId:     gopkg.Pointer(env.ClientId),
			ClientSecret: gopkg.Pointer(encryption.Encrypt(env.ClientSecret, env.ClientId)), // Make sure encryption and util packages are imported
			SecureKey:    gopkg.Pointer(encryption.Encrypt(util.RandomSecureKey(), env.ClientId)), // Make sure util is imported
			IsRoot:       gopkg.Pointer(true),
			TenantId:     gopkg.Pointer(enum.Tenant(SID(savedTenant.ID))),
		})
		if err != nil {
			logrus.WithError(err).Errorln("Failed to save root client during initialization")
		}

		_, err = s.user.Save(ctx, &UserDomain{ // Use new field name
			Name:       gopkg.Pointer(env.RootUser),
			Phone:      gopkg.Pointer(env.RootUser),
			Email:      gopkg.Pointer(env.RootUser), // Make sure gopkg is imported
			Username:   gopkg.Pointer(env.RootUser),
			Password:   gopkg.Pointer(util.HashPassword(env.RootPass)),
			DataStatus: gopkg.Pointer(enum.DataStatusEnable),
			RoleIds:    gopkg.Pointer([]string{}),
			IsRoot:     gopkg.Pointer(true),
			TenantId:   gopkg.Pointer(enum.Tenant(SID(savedTenant.ID))),
		})
		if err != nil {
			logrus.WithError(err).Errorln("Failed to save root user during initialization")
		}
	}

	return s
}

func OIDs(ids []string) []primitive.ObjectID {
	return gopkg.MapFunc(ids, func(id string) primitive.ObjectID { return OID(id) })
}

func SIDs(oids []primitive.ObjectID) []string {
	return gopkg.MapFunc(oids, func(oid primitive.ObjectID) string { return SID(oid) })
}

// Removed import "errors" from the bottom
