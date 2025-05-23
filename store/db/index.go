package db

import (
	"app/env"
	"app/pkg/encryption"
	"app/pkg/enum"
	"app/pkg/util"
	"context"
	"fmt"
	"time"

	"github.com/nhnghia272/gopkg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type Db struct {
	client          *mongo.Client
	Bucket          *gridfs.Bucket
	AuditLog        *audit_log
	Client          *client
	Role            *role
	Tenant          *tenant
	User            *user
	Member          *member
	Membership      *membership
	PlantSlot       *plantSlot
	Plant           *plant
	CareRecord      *careRecord
	Harvest         *harvest
	PlantType       *plantType
	SeasonalCatalog *seasonalCatalog
	Payment         *payment
	NFTRecord       *nftRecord
}

func New() *Db {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := connstring.ParseAndValidate(env.MongoUri)
	if err != nil {
		logrus.Fatalln("Mongo", err)
	}

	opts := &options.BSONOptions{UseJSONStructTags: true, NilMapAsEmpty: true, NilSliceAsEmpty: true, NilByteSliceAsEmpty: true}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conn.Original).SetBSONOptions(opts).SetReadPreference(readpref.SecondaryPreferred()))
	if err != nil {
		logrus.Fatalln("Mongo", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logrus.Fatalln("Mongo", err)
	}

	mongodb := client.Database(conn.Database)

	bucket, err := gridfs.NewBucket(mongodb)
	if err != nil {
		logrus.Fatalln("Mongo", err)
	}

	fmt.Printf("Mongo connected %v\n", env.MongoUri)

	db := &Db{
		client:          client,
		Bucket:          bucket,
		AuditLog:        newAuditLog(ctx, mongodb.Collection("audit_log")),
		Client:          newClient(ctx, mongodb.Collection("client")),
		Role:            newRole(ctx, mongodb.Collection("role")),
		Tenant:          newTenant(ctx, mongodb.Collection("tenant")),
		User:            newUser(ctx, mongodb.Collection("user")),
		Member:          newMember(ctx, mongodb.Collection("member")),
		Membership:      newMembership(ctx, mongodb.Collection("membership")),
		PlantSlot:       newPlantSlot(ctx, mongodb.Collection("plant_slot")),
		Plant:           newPlant(ctx, mongodb.Collection("plant")),
		CareRecord:      newCareRecord(ctx, mongodb.Collection("care_record")),
		Harvest:         newHarvest(ctx, mongodb.Collection("harvest")),
		PlantType:       newPlantType(ctx, mongodb.Collection("plant_type")),
		SeasonalCatalog: newSeasonalCatalog(ctx, mongodb.Collection("seasonal_catalog")),
		Payment:         newPayment(ctx, mongodb.Collection("payment")),
		NFTRecord:       newNFTRecord(ctx, mongodb.Collection("nft_record")),
	}

	return db.initialize(ctx)
}

func (s Db) Uri() string {
	return env.MongoUri
}

func (s Db) Instance() *mongo.Client {
	return s.client
}

func (s Db) Session() (mongo.Session, error) {
	return s.client.StartSession()
}

func (s *Db) initialize(ctx context.Context) *Db {
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

func Regex(v string) M {
	return M{"$regex": v, "$options": "sim"}
}

func OID(id string) primitive.ObjectID {
	oid, _ := primitive.ObjectIDFromHex(id)
	return oid
}

func OIDs(ids []string) []primitive.ObjectID {
	return gopkg.MapFunc(ids, func(id string) primitive.ObjectID { return OID(id) })
}

func SID(oid primitive.ObjectID) string {
	return oid.Hex()
}

func SIDs(oids []primitive.ObjectID) []string {
	return gopkg.MapFunc(oids, func(oid primitive.ObjectID) string { return SID(oid) })
}
