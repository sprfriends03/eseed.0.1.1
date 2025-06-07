package db

import (
	"app/pkg/enum"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthSessionDto struct {
	Name        string            `json:"name"`
	Phone       string            `json:"phone"`
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	UserId      string            `json:"user_id"`
	TenantId    enum.Tenant       `json:"tenant_id"`
	Permissions []enum.Permission `json:"permissions"`
	IsRoot      bool              `json:"is_root"`
	IsTenant    bool              `json:"is_tenant"`
	AccessToken string            `json:"access_token"`

	// Member-specific session fields
	IsMember         *bool   `json:"is_member,omitempty"`
	MembershipID     *string `json:"membership_id,omitempty"`
	MembershipStatus *string `json:"membership_status,omitempty"`
	KYCStatus        *string `json:"kyc_status,omitempty"`
}

type AuthTokenDto struct {
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthRefreshTokenData struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Keycode      string `json:"keycode" validate:"required"`
}

type AuthLoginData struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Keycode  string `json:"keycode" validate:"required"`
}

// AuthRegisterData is data for user registration
// AuthRegisterData godoc
// @Description AuthRegisterData is data for user registration
type AuthRegisterData struct {
	Keycode  string `json:"keycode" binding:"required" example:"tenant_A"`
	Username string `json:"username" binding:"required,alphanum,min=3,max=30" example:"john_doe"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"SecurePassword123"`
}

type AuthChangePasswordData struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type MemberRegisterData struct {
	Keycode     string `json:"keycode" binding:"required" example:"tenant_A"`
	Username    string `json:"username" binding:"required,alphanum,min=3,max=30" example:"member_user"`
	Password    string `json:"password" binding:"required,min=8,max=100" example:"SecurePassword123"`
	Email       string `json:"email" binding:"required,email" example:"member@example.com"`
	FirstName   string `json:"first_name" binding:"required" example:"John"`
	LastName    string `json:"last_name" binding:"required" example:"Doe"`
	DateOfBirth string `json:"date_of_birth" binding:"required" example:"1990-01-15"` // Expects YYYY-MM-DD format
	Phone       string `json:"phone" binding:"required,e164" example:"+12125552368"`
	// Add other fields like Address if they are collected at registration
}

type M map[string]any

type Query struct {
	Page   int64  `json:"page" form:"page" validate:"required,min=1" default:"1"`
	Limit  int64  `json:"limit" form:"limit" validate:"required,min=1,max=100" default:"10"`
	Sorts  string `json:"sorts" form:"sorts" validate:"omitempty" default:"created_at.desc"`
	Filter M      `json:"-"`
}

type BaseDomain struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt *time.Time         `json:"created_at,omitempty" validate:"omitempty"`
	UpdatedAt *time.Time         `json:"updated_at,omitempty" validate:"omitempty"`
	CreatedBy *string            `json:"created_by,omitempty" validate:"omitempty"`
	UpdatedBy *string            `json:"updated_by,omitempty" validate:"omitempty"`
}

func (s *BaseDomain) BeforeSave() {
	if s.ID.IsZero() {
		s.CreatedAt = gopkg.Pointer(time.Now())
	}
	s.UpdatedAt = gopkg.Pointer(time.Now())
}
