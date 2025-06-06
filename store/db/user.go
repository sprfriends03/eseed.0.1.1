package db

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"net/http"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDomain struct {
	BaseDomain                      `json:"inline"`
	Name                            *string          `json:"name,omitempty" validate:"omitempty"`
	Phone                           *string          `json:"phone,omitempty" validate:"omitempty,lowercase"`
	Email                           *string          `json:"email,omitempty" validate:"omitempty,lowercase"`
	Username                        *string          `json:"username,omitempty" validate:"omitempty,lowercase"`
	Password                        *string          `json:"password,omitempty" validate:"omitempty"`
	DataStatus                      *enum.DataStatus `json:"data_status,omitempty" validate:"omitempty,data_status"`
	RoleIds                         *[]string        `json:"role_ids,omitempty" validate:"omitempty,dive,len=24"`
	IsRoot                          *bool            `json:"is_root,omitempty" validate:"omitempty"`
	TenantId                        *enum.Tenant     `json:"tenant_id,omitempty" validate:"omitempty,len=24"`
	VersionToken                    *int64           `json:"version_token,omitempty" validate:"omitempty"`
	EmailVerified                   *bool            `json:"email_verified,omitempty" bson:"email_verified,omitempty"`
	EmailVerificationToken          *string          `json:"-" bson:"email_verification_token,omitempty"`
	EmailVerificationTokenExpiresAt *time.Time       `json:"-" bson:"email_verification_token_expires_at,omitempty"`
	DateOfBirth                     *time.Time       `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty" validate:"omitempty"`
	EmailVerifiedAt                 *time.Time       `json:"email_verified_at,omitempty" bson:"email_verified_at,omitempty"`
	PrivacyPreferences              *PrivacySettings `json:"privacy_preferences,omitempty" bson:"privacy_preferences,omitempty" validate:"omitempty"`
	ProfilePicture                  *string          `json:"profile_picture,omitempty" bson:"profile_picture,omitempty" validate:"omitempty"`
}

// Privacy settings for user profile
type PrivacySettings struct {
	ShowEmail          *bool `json:"show_email,omitempty" bson:"show_email,omitempty" validate:"omitempty"`
	ShowPhone          *bool `json:"show_phone,omitempty" bson:"show_phone,omitempty" validate:"omitempty"`
	ShowDateOfBirth    *bool `json:"show_date_of_birth,omitempty" bson:"show_date_of_birth,omitempty" validate:"omitempty"`
	ShowName           *bool `json:"show_name,omitempty" bson:"show_name,omitempty" validate:"omitempty"`
	ShowProfilePicture *bool `json:"show_profile_picture,omitempty" bson:"show_profile_picture,omitempty" validate:"omitempty"`
	IsPublic           *bool `json:"is_public,omitempty" bson:"is_public,omitempty" validate:"omitempty"`
}

func (s *UserDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

func (s UserDomain) CmsDto() *UserCmsDto {
	return &UserCmsDto{
		ID:         SID(s.ID),
		Name:       gopkg.Value(s.Name),
		Phone:      gopkg.Value(s.Phone),
		Email:      gopkg.Value(s.Email),
		Username:   gopkg.Value(s.Username),
		DataStatus: gopkg.Value(s.DataStatus),
		RoleIds:    gopkg.Value(s.RoleIds),
		UpdatedBy:  gopkg.Value(s.UpdatedBy),
		UpdatedAt:  gopkg.Value(s.UpdatedAt),
	}
}

func (s *UserDomain) Cache() *UserCache {
	// Create a cache object for privacy settings if it exists
	var privacySettings *PrivacySettings
	if s.PrivacyPreferences != nil {
		privacySettings = s.PrivacyPreferences
	}

	return &UserCache{
		ID:              SID(s.ID),
		Name:            gopkg.Value(s.Name),
		Phone:           gopkg.Value(s.Phone),
		Email:           gopkg.Value(s.Email),
		Username:        gopkg.Value(s.Username),
		DataStatus:      gopkg.Value(s.DataStatus),
		RoleIds:         gopkg.Value(s.RoleIds),
		IsRoot:          gopkg.Value(s.IsRoot),
		TenantId:        gopkg.Value(s.TenantId),
		VersionToken:    gopkg.Value(s.VersionToken),
		Permissions:     make([]enum.Permission, 0),
		EmailVerified:   gopkg.Value(s.EmailVerified),
		EmailVerifiedAt: s.EmailVerifiedAt,
		DateOfBirth:     s.DateOfBirth,
		PrivacySettings: privacySettings,
		ProfilePicture:  gopkg.Value(s.ProfilePicture),
	}
}

type UserCmsDto struct {
	ID         string          `json:"user_id" example:"671db9eca1f1b1bdbf3d4618"`
	Name       string          `json:"name" example:"Aloha"`
	Phone      string          `json:"phone" example:"0973123456"`
	Email      string          `json:"email" example:"aloha@email.com"`
	Username   string          `json:"username" example:"aloha"`
	DataStatus enum.DataStatus `json:"data_status"`
	RoleIds    []string        `json:"role_ids" example:"671db9eca1f1b1bdbf3d4617"`
	UpdatedBy  string          `json:"updated_by" example:"editor"`
	UpdatedAt  time.Time       `json:"updated_at" example:"2006-01-02T03:04:05Z"`
}

type UserCache struct {
	ID              string            `json:"user_id"`
	Name            string            `json:"name"`
	Phone           string            `json:"phone"`
	Email           string            `json:"email"`
	Username        string            `json:"username"`
	DataStatus      enum.DataStatus   `json:"data_status"`
	RoleIds         []string          `json:"role_ids"`
	IsRoot          bool              `json:"is_root"`
	TenantId        enum.Tenant       `json:"tenant_id"`
	VersionToken    int64             `json:"version_token"`
	IsTenant        bool              `json:"is_tenant"`
	Permissions     []enum.Permission `json:"permissions"`
	EmailVerified   bool              `json:"email_verified"`
	EmailVerifiedAt *time.Time        `json:"email_verified_at,omitempty"`
	DateOfBirth     *time.Time        `json:"date_of_birth,omitempty"`
	PrivacySettings *PrivacySettings  `json:"privacy_preferences,omitempty"`
	ProfilePicture  string            `json:"profile_picture,omitempty"`
}

type UserCmsData struct {
	Name       string          `json:"name" validate:"required" example:"Aloha"`
	Phone      string          `json:"phone" validate:"required,lowercase" example:"0973123456"`
	Email      string          `json:"email" validate:"required,lowercase,email" example:"aloha@email.com"`
	Username   string          `json:"username" validate:"required,lowercase" example:"aloha"`
	DataStatus enum.DataStatus `json:"data_status" validate:"required,data_status"`
	RoleIds    []string        `json:"role_ids" validate:"required,min=1,dive,len=24" example:"671db9eca1f1b1bdbf3d4617"`
}

func (s UserCmsData) Domain(domain *UserDomain) *UserDomain {
	domain.Name = gopkg.Pointer(s.Name)
	domain.Phone = gopkg.Pointer(s.Phone)
	domain.Email = gopkg.Pointer(s.Email)
	domain.Username = gopkg.Pointer(s.Username)
	domain.DataStatus = gopkg.Pointer(s.DataStatus)
	domain.RoleIds = gopkg.Pointer(s.RoleIds)
	return domain
}

type UserQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	RoleId     *string          `json:"role_id" form:"role_id" validate:"omitempty,len=24"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
	TenantId   *enum.Tenant     `json:"tenant_id" form:"tenant_id" validate:"omitempty,len=24"`
}

func (s *UserQuery) Build() *UserQuery {
	if s.Filter == nil {
		s.Filter = M{}
	}
	if s.Search != nil {
		s.Filter["$or"] = []M{{"name": Regex(gopkg.Value(s.Search))}, {"email": Regex(gopkg.Value(s.Search))}, {"phone": Regex(gopkg.Value(s.Search))}, {"username": Regex(gopkg.Value(s.Search))}}
	}
	if s.RoleId != nil {
		s.Filter["role_ids"] = s.RoleId
	}
	if s.DataStatus != nil {
		s.Filter["data_status"] = s.DataStatus
	}
	if s.TenantId != nil {
		s.Filter["tenant_id"] = s.TenantId
	}
	s.Filter["is_root"] = false
	return s
}

type UserCmsQuery struct {
	Query
	Search     *string          `json:"search" form:"search" validate:"omitempty"`
	RoleId     *string          `json:"role_id" form:"role_id" validate:"omitempty,len=24"`
	DataStatus *enum.DataStatus `json:"data_status" form:"data_status" validate:"omitempty,data_status"`
}

func (s *UserCmsQuery) BuildCore() *UserQuery {
	return &UserQuery{Query: s.Query, Search: s.Search, RoleId: s.RoleId, DataStatus: s.DataStatus}
}

type user struct {
	repo *repo
}

func newUser(ctx context.Context, col *mongo.Collection) *user {
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"username": bson.M{"$exists": true, "$gt": ""}}),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"phone": bson.M{"$exists": true, "$gt": ""}}),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"email": bson.M{"$exists": true, "$gt": ""}}),
	})
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email_verification_token", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetPartialFilterExpression(bson.M{"email_verification_token": bson.M{"$exists": true, "$gt": ""}}),
	})
	return &user{newrepo(col)}
}

func (s user) CollectionName() string { return s.repo.col.Name() }

func (s user) Save(ctx context.Context, domain *UserDomain, opts ...*options.UpdateOptions) (*UserDomain, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}
	id, err := s.repo.Save(ctx, domain.ID, domain, opts...)
	if err != nil {
		return nil, err
	}
	domain.ID = id
	return s.FindOneById(ctx, SID(id))
}

func (s user) FindOne(ctx context.Context, filter M, opts ...*options.FindOneOptions) (*UserDomain, error) {
	domain := &UserDomain{}
	return domain, s.repo.FindOne(ctx, filter, &domain, opts...)
}

func (s user) UpdateOne(ctx context.Context, filter M, update M, opts ...*options.UpdateOptions) error {
	return s.repo.UpdateOne(ctx, filter, update, opts...)
}

func (s user) Count(ctx context.Context, q *UserQuery, opts ...*options.CountOptions) int64 {
	return s.repo.CountDocuments(ctx, q.Build().Query, opts...)
}

func (s user) FindAll(ctx context.Context, q *UserQuery, opts ...*options.FindOptions) ([]*UserDomain, error) {
	domains := make([]*UserDomain, 0)
	return domains, s.repo.FindAll(ctx, q.Build().Query, &domains, opts...)
}

func (s user) FindOneById(ctx context.Context, id string) (*UserDomain, error) {
	return s.FindOne(ctx, M{"_id": OID(id)})
}

func (s user) FindAllByRole(ctx context.Context, roleId string) ([]*UserDomain, error) {
	return s.FindAll(ctx, &UserQuery{Query: Query{Filter: M{"role_ids": roleId}}})
}

func (s user) FindAllByTenant(ctx context.Context, tenant enum.Tenant) ([]*UserDomain, error) {
	return s.FindAll(ctx, &UserQuery{Query: Query{Filter: M{"tenant_id": tenant}}})
}

func (s user) FindOneByTenant_Username(ctx context.Context, tenant enum.Tenant, username string) (*UserDomain, error) {
	return s.FindOne(ctx, M{"tenant_id": tenant, "username": username})
}

func (s user) FindOneByEmailVerificationToken(ctx context.Context, token string) (*UserDomain, error) {
	if token == "" {
		return nil, ecode.New(http.StatusBadRequest, "empty_verification_token")
	}
	return s.FindOne(ctx, M{"email_verification_token": token})
}

func (s user) IncrementVersionToken(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx, M{"_id": OID(id)}, M{"$inc": M{"version_token": 1}})
}

// UserProfileDto for profile view
type UserProfileDto struct {
	ID              string           `json:"user_id" example:"671db9eca1f1b1bdbf3d4618"`
	Name            string           `json:"name,omitempty" example:"Aloha"`
	Phone           string           `json:"phone,omitempty" example:"0973123456"`
	Email           string           `json:"email,omitempty" example:"aloha@email.com"`
	Username        string           `json:"username" example:"aloha"`
	DateOfBirth     *time.Time       `json:"date_of_birth,omitempty"`
	EmailVerified   bool             `json:"email_verified"`
	EmailVerifiedAt *time.Time       `json:"email_verified_at,omitempty"`
	PrivacySettings *PrivacySettings `json:"privacy_preferences,omitempty"`
	ProfilePicture  string           `json:"profile_picture,omitempty"`
}

// Convert UserDomain to UserProfileDto with privacy settings applied
func (s UserDomain) ProfileDto(applyPrivacy bool) *UserProfileDto {
	profileDto := &UserProfileDto{
		ID:              SID(s.ID),
		Username:        gopkg.Value(s.Username),
		EmailVerified:   gopkg.Value(s.EmailVerified),
		EmailVerifiedAt: s.EmailVerifiedAt,
	}

	// Apply privacy settings if requested and available
	if applyPrivacy && s.PrivacyPreferences != nil {
		privacySettings := s.PrivacyPreferences

		if privacySettings.ShowEmail != nil && *privacySettings.ShowEmail {
			profileDto.Email = gopkg.Value(s.Email)
		}

		if privacySettings.ShowPhone != nil && *privacySettings.ShowPhone {
			profileDto.Phone = gopkg.Value(s.Phone)
		}

		if privacySettings.ShowName != nil && *privacySettings.ShowName {
			profileDto.Name = gopkg.Value(s.Name)
		}

		if privacySettings.ShowDateOfBirth != nil && *privacySettings.ShowDateOfBirth {
			profileDto.DateOfBirth = s.DateOfBirth
		}

		if privacySettings.ShowProfilePicture != nil && *privacySettings.ShowProfilePicture {
			profileDto.ProfilePicture = gopkg.Value(s.ProfilePicture)
		}
	} else {
		// If not applying privacy or no settings, show all fields for self-view
		profileDto.Email = gopkg.Value(s.Email)
		profileDto.Phone = gopkg.Value(s.Phone)
		profileDto.Name = gopkg.Value(s.Name)
		profileDto.DateOfBirth = s.DateOfBirth
		profileDto.PrivacySettings = s.PrivacyPreferences
		profileDto.ProfilePicture = gopkg.Value(s.ProfilePicture)
	}

	return profileDto
}

// UserProfileUpdateData for profile updates
type UserProfileUpdateData struct {
	Name           string     `json:"name" validate:"required" example:"Aloha"`
	Phone          string     `json:"phone" validate:"required,lowercase" example:"0973123456"`
	Email          string     `json:"email" validate:"required,lowercase,email" example:"aloha@email.com"`
	DateOfBirth    *time.Time `json:"date_of_birth" validate:"required"`
	ProfilePicture string     `json:"profile_picture,omitempty"`
}

// Apply updates to UserDomain
func (s UserProfileUpdateData) Domain(domain *UserDomain) *UserDomain {
	domain.Name = gopkg.Pointer(s.Name)
	domain.Phone = gopkg.Pointer(s.Phone)
	domain.Email = gopkg.Pointer(s.Email)
	domain.DateOfBirth = s.DateOfBirth
	if s.ProfilePicture != "" {
		domain.ProfilePicture = gopkg.Pointer(s.ProfilePicture)
	}
	return domain
}

// UserPrivacyUpdateData for privacy setting updates
type UserPrivacyUpdateData struct {
	ShowEmail          bool `json:"show_email" validate:"required"`
	ShowPhone          bool `json:"show_phone" validate:"required"`
	ShowDateOfBirth    bool `json:"show_date_of_birth" validate:"required"`
	ShowName           bool `json:"show_name" validate:"required"`
	ShowProfilePicture bool `json:"show_profile_picture" validate:"required"`
	IsPublic           bool `json:"is_public" validate:"required"`
}

// Apply privacy updates to UserDomain
func (s UserPrivacyUpdateData) Domain(domain *UserDomain) *UserDomain {
	// Create privacy settings object if it doesn't exist
	if domain.PrivacyPreferences == nil {
		domain.PrivacyPreferences = &PrivacySettings{}
	}

	domain.PrivacyPreferences.ShowEmail = gopkg.Pointer(s.ShowEmail)
	domain.PrivacyPreferences.ShowPhone = gopkg.Pointer(s.ShowPhone)
	domain.PrivacyPreferences.ShowDateOfBirth = gopkg.Pointer(s.ShowDateOfBirth)
	domain.PrivacyPreferences.ShowName = gopkg.Pointer(s.ShowName)
	domain.PrivacyPreferences.ShowProfilePicture = gopkg.Pointer(s.ShowProfilePicture)
	domain.PrivacyPreferences.IsPublic = gopkg.Pointer(s.IsPublic)

	return domain
}
