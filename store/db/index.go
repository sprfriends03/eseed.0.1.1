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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

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
	ctx    context.Context
	client *mongo.Client

	User         *user
	Role         *role
	Member       *member
	Membership   *membership
	PlantType    *plantType
	PlantSlot    *plantSlot
	Plant        *plant
	CareRecord   *careRecord
	Harvest      *harvest
	Notification *notification
	Tenant       *tenant
	Client       *client
	AuditLog     *audit_log
}

// Create a MongoDB client
func createMongoClient(ctx context.Context) *mongo.Client {
	uri := env.MongoUri
	connectTimeout := 10 * time.Second

	if uri == "" {
		logrus.Fatal("MongoDB URI is empty - check env configuration")
	}

	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(connectTimeout).
		SetReadConcern(readconcern.Majority()).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		logrus.WithError(err).Fatalln("Failed to connect to MongoDB")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logrus.WithError(err).Fatalln("Failed to ping MongoDB")
	}

	logrus.Infoln("Connected to MongoDB at", uri)
	return client
}

// Init initializes the database connections
func Init(ctx context.Context) *DB {
	client := createMongoClient(ctx)

	// Extract database name from the URI
	uri := env.MongoUri
	dbName := "app" // default fallback
	if uri != "" {
		// Parse the URI to extract database name
		// Format: mongodb://user:pass@host:port/database?options
		// Find the last slash before any query parameters
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
	db := client.Database(dbName)

	_db = &DB{
		ctx:          ctx,
		client:       client,
		User:         newUser(ctx, db.Collection(userCollection)),
		Role:         newRole(ctx, db.Collection(roleCollection)),
		Member:       newMember(ctx, db.Collection(memberCollection)),
		Membership:   newMembership(ctx, db.Collection(membershipCollection)),
		PlantType:    newPlantType(ctx, db.Collection(plantTypeCollection)),
		PlantSlot:    newPlantSlot(ctx, db.Collection(plantSlotCollection)),
		Plant:        newPlant(ctx, db.Collection(plantCollection)),
		CareRecord:   newCareRecord(ctx, db.Collection(careRecordCollection)),
		Harvest:      newHarvest(ctx, db.Collection(harvestCollection)),
		Notification: newNotification(ctx, db.Collection(notificationCollection)),
		Tenant:       newTenant(ctx, db.Collection(tenantCollection)),
		Client:       newClient(ctx, db.Collection(clientCollection)),
		AuditLog:     newAuditLog(ctx, db.Collection(auditLogCollection)),
	}

	// Initialize root tenant and client
	_db.initialize(ctx)

	return _db
}

// Get returns the database instance
func Get() *DB {
	return _db
}

// Close closes the database connection
func (db *DB) Close() {
	if db.client != nil {
		if err := db.client.Disconnect(db.ctx); err != nil {
			logrus.WithError(err).Warningln("Error disconnecting from MongoDB")
		}
	}
}

// SID converts an ObjectID to string
func SID(id primitive.ObjectID) string {
	return id.Hex()
}

// OID converts a string to an ObjectID
func OID(id string) primitive.ObjectID {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID
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

func (s DB) Uri() string {
	return env.MongoUri
}

func (s DB) Instance() *mongo.Client {
	return s.client
}

func (s DB) Session() (mongo.Session, error) {
	return s.client.StartSession()
}

func (s *DB) initialize(ctx context.Context) *DB {
	tenant := &TenantDomain{
		Name:       gopkg.Pointer(env.RootUser),
		Keycode:    gopkg.Pointer(env.RootUser),
		Username:   gopkg.Pointer(env.RootUser),
		DataStatus: gopkg.Pointer(enum.DataStatusEnable),
		IsRoot:     gopkg.Pointer(true),
	}
	s.Tenant.Save(ctx, tenant)

	if !tenant.ID.IsZero() {
		s.Client.Save(ctx, &ClientDomain{
			Name:         gopkg.Pointer(env.RootUser),
			ClientId:     gopkg.Pointer(env.ClientId),
			ClientSecret: gopkg.Pointer(encryption.Encrypt(env.ClientSecret, env.ClientId)),
			SecureKey:    gopkg.Pointer(encryption.Encrypt(util.RandomSecureKey(), env.ClientId)),
			IsRoot:       gopkg.Pointer(true),
			TenantId:     gopkg.Pointer(enum.Tenant(SID(tenant.ID))),
		})

		s.User.Save(ctx, &UserDomain{
			Name:       gopkg.Pointer(env.RootUser),
			Phone:      gopkg.Pointer(env.RootUser),
			Email:      gopkg.Pointer(env.RootUser),
			Username:   gopkg.Pointer(env.RootUser),
			Password:   gopkg.Pointer(util.HashPassword(env.RootPass)),
			DataStatus: gopkg.Pointer(enum.DataStatusEnable),
			RoleIds:    gopkg.Pointer([]string{}),
			IsRoot:     gopkg.Pointer(true),
			TenantId:   gopkg.Pointer(enum.Tenant(SID(tenant.ID))),
		})
	}

	return s
}

func OIDs(ids []string) []primitive.ObjectID {
	return gopkg.MapFunc(ids, func(id string) primitive.ObjectID { return OID(id) })
}

func SIDs(oids []primitive.ObjectID) []string {
	return gopkg.MapFunc(oids, func(oid primitive.ObjectID) string { return SID(oid) })
}
