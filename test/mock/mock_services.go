package mock

import (
	"app/pkg/enum"
	"app/store/db"
	"context"
	"time"

	"github.com/nhnghia272/gopkg"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APIError represents an API error response
type APIError struct {
	Message string `json:"message"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// MockMemberService is a mock implementation of the MemberService
type MockMemberService struct {
	members map[string]*db.MemberDomain
}

// NewMockMemberService creates a new instance of MockMemberService
func NewMockMemberService() *MockMemberService {
	return &MockMemberService{
		members: make(map[string]*db.MemberDomain),
	}
}

// Create adds a new member
func (m *MockMemberService) Create(ctx context.Context, tenantID enum.Tenant, userID string, member *db.MemberDomain) (string, error) {
	id := primitive.NewObjectID()
	now := time.Now()

	member.ID = id
	member.CreatedAt = gopkg.Pointer(now)
	member.UpdatedAt = gopkg.Pointer(now)
	member.CreatedBy = gopkg.Pointer(userID)
	member.UpdatedBy = gopkg.Pointer(userID)
	member.TenantId = gopkg.Pointer(tenantID)

	m.members[id.Hex()] = member
	return id.Hex(), nil
}

// Get retrieves a member by ID
func (m *MockMemberService) Get(ctx context.Context, tenantID enum.Tenant, id string) (*db.MemberDomain, error) {
	member, exists := m.members[id]
	if !exists {
		return nil, ErrNotFound
	}

	if *member.TenantId != tenantID {
		return nil, ErrNotFound
	}

	return member, nil
}

// Update modifies an existing member
func (m *MockMemberService) Update(ctx context.Context, tenantID enum.Tenant, userID string, id string, updates map[string]interface{}) error {
	member, exists := m.members[id]
	if !exists {
		return ErrNotFound
	}

	if *member.TenantId != tenantID {
		return ErrNotFound
	}

	// Apply updates
	if name, ok := updates["first_name"].(string); ok {
		member.FirstName = gopkg.Pointer(name)
	}

	if email, ok := updates["email"].(string); ok {
		member.Email = gopkg.Pointer(email)
	}

	if phone, ok := updates["phone"].(string); ok {
		member.Phone = gopkg.Pointer(phone)
	}

	if status, ok := updates["member_status"].(string); ok {
		member.MemberStatus = gopkg.Pointer(status)
	}

	member.UpdatedAt = gopkg.Pointer(time.Now())
	member.UpdatedBy = gopkg.Pointer(userID)

	return nil
}

// Delete softly removes a member
func (m *MockMemberService) Delete(ctx context.Context, tenantID enum.Tenant, userID string, id string) error {
	member, exists := m.members[id]
	if !exists {
		return ErrNotFound
	}

	if *member.TenantId != tenantID {
		return ErrNotFound
	}

	// Set data status to disabled (soft delete)
	member.UpdatedAt = gopkg.Pointer(time.Now())
	member.UpdatedBy = gopkg.Pointer(userID)

	return nil
}

// List returns members that match the given filters
func (m *MockMemberService) List(ctx context.Context, tenantID enum.Tenant, filters map[string]interface{}, pagination *db.Query) ([]*db.MemberDomain, int64, error) {
	var results []*db.MemberDomain

	for _, member := range m.members {
		if *member.TenantId != tenantID {
			continue
		}

		// Apply filters
		match := true
		for key, value := range filters {
			switch key {
			case "member_status":
				if status, ok := value.(string); ok && member.MemberStatus != nil && *member.MemberStatus != status {
					match = false
				}
			case "first_name":
				if name, ok := value.(string); ok && member.FirstName != nil && *member.FirstName != name {
					match = false
				}
			case "email":
				if email, ok := value.(string); ok && member.Email != nil && *member.Email != email {
					match = false
				}
			}
		}

		if match {
			results = append(results, member)
		}
	}

	// Apply pagination
	if pagination != nil {
		start := pagination.Page * pagination.Limit
		end := start + pagination.Limit

		if start < int64(len(results)) {
			if end > int64(len(results)) {
				end = int64(len(results))
			}
			results = results[start:end]
		} else {
			results = []*db.MemberDomain{}
		}
	}

	return results, int64(len(results)), nil
}

// Custom errors
var (
	ErrNotFound = &APIError{Message: "Resource not found"}
	ErrInternal = &APIError{Message: "Internal server error"}
)

// MockPlantService is a mock implementation of the PlantService
type MockPlantService struct {
	plants map[string]*db.PlantDomain
}

// NewMockPlantService creates a new instance of MockPlantService
func NewMockPlantService() *MockPlantService {
	return &MockPlantService{
		plants: make(map[string]*db.PlantDomain),
	}
}

// Create adds a new plant
func (m *MockPlantService) Create(ctx context.Context, tenantID enum.Tenant, userID string, plant *db.PlantDomain) (string, error) {
	id := primitive.NewObjectID()
	now := time.Now()

	plant.ID = id
	plant.CreatedAt = gopkg.Pointer(now)
	plant.UpdatedAt = gopkg.Pointer(now)
	plant.CreatedBy = gopkg.Pointer(userID)
	plant.UpdatedBy = gopkg.Pointer(userID)
	plant.TenantId = gopkg.Pointer(tenantID)

	if plant.Status == nil {
		plant.Status = gopkg.Pointer(string(enum.PlantStatusGrowing))
	}

	if plant.PlantedDate == nil {
		plant.PlantedDate = gopkg.Pointer(now)
	}

	m.plants[id.Hex()] = plant
	return id.Hex(), nil
}

// Get retrieves a plant by ID
func (m *MockPlantService) Get(ctx context.Context, tenantID enum.Tenant, id string) (*db.PlantDomain, error) {
	plant, exists := m.plants[id]
	if !exists {
		return nil, ErrNotFound
	}

	if *plant.TenantId != tenantID {
		return nil, ErrNotFound
	}

	return plant, nil
}

// Update modifies an existing plant
func (m *MockPlantService) Update(ctx context.Context, tenantID enum.Tenant, userID string, id string, updates map[string]interface{}) error {
	plant, exists := m.plants[id]
	if !exists {
		return ErrNotFound
	}

	if *plant.TenantId != tenantID {
		return ErrNotFound
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		plant.Name = gopkg.Pointer(name)
	}

	if status, ok := updates["status"].(string); ok {
		plant.Status = gopkg.Pointer(status)
	}

	plant.UpdatedAt = gopkg.Pointer(time.Now())
	plant.UpdatedBy = gopkg.Pointer(userID)

	return nil
}

// Delete softly removes a plant
func (m *MockPlantService) Delete(ctx context.Context, tenantID enum.Tenant, userID string, id string) error {
	plant, exists := m.plants[id]
	if !exists {
		return ErrNotFound
	}

	if *plant.TenantId != tenantID {
		return ErrNotFound
	}

	// Soft delete
	plant.UpdatedAt = gopkg.Pointer(time.Now())
	plant.UpdatedBy = gopkg.Pointer(userID)

	return nil
}

// List returns plants that match the given filters
func (m *MockPlantService) List(ctx context.Context, tenantID enum.Tenant, filters map[string]interface{}, pagination *db.Query) ([]*db.PlantDomain, int64, error) {
	var results []*db.PlantDomain

	for _, plant := range m.plants {
		if *plant.TenantId != tenantID {
			continue
		}

		// Apply filters
		match := true
		for key, value := range filters {
			switch key {
			case "status":
				if status, ok := value.(string); ok && plant.Status != nil && *plant.Status != status {
					match = false
				}
			case "member_id":
				if memberID, ok := value.(string); ok && plant.MemberID != nil && *plant.MemberID != memberID {
					match = false
				}
			case "plant_type_id":
				if typeID, ok := value.(string); ok && plant.PlantTypeID != nil && *plant.PlantTypeID != typeID {
					match = false
				}
			}
		}

		if match {
			results = append(results, plant)
		}
	}

	// Apply pagination
	if pagination != nil {
		start := pagination.Page * pagination.Limit
		end := start + pagination.Limit

		if start < int64(len(results)) {
			if end > int64(len(results)) {
				end = int64(len(results))
			}
			results = results[start:end]
		} else {
			results = []*db.PlantDomain{}
		}
	}

	return results, int64(len(results)), nil
}

// MockNotificationService is a mock implementation of the NotificationService
type MockNotificationService struct {
	notifications map[string]*db.NotificationDomain
}

// NewMockNotificationService creates a new instance of MockNotificationService
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make(map[string]*db.NotificationDomain),
	}
}

// Create adds a new notification
func (m *MockNotificationService) Create(ctx context.Context, tenantID enum.Tenant, notification *db.NotificationDomain) (string, error) {
	id := primitive.NewObjectID()
	now := time.Now()

	notification.ID = id
	notification.CreatedAt = gopkg.Pointer(now)
	notification.TenantId = gopkg.Pointer(tenantID)

	if notification.Status == nil {
		notification.Status = gopkg.Pointer(string(enum.NotificationStatusUnread))
	}

	m.notifications[id.Hex()] = notification
	return id.Hex(), nil
}

// MarkAsRead marks a notification as read
func (m *MockNotificationService) MarkAsRead(ctx context.Context, tenantID enum.Tenant, id string) error {
	notification, exists := m.notifications[id]
	if !exists {
		return ErrNotFound
	}

	if *notification.TenantId != tenantID {
		return ErrNotFound
	}

	notification.Status = gopkg.Pointer(string(enum.NotificationStatusRead))
	notification.ReadAt = gopkg.Pointer(time.Now())

	return nil
}

// List returns notifications for a user
func (m *MockNotificationService) List(ctx context.Context, tenantID enum.Tenant, userID string, pagination *db.Query) ([]*db.NotificationDomain, int64, error) {
	var results []*db.NotificationDomain

	for _, notification := range m.notifications {
		if *notification.TenantId != tenantID {
			continue
		}

		if notification.MemberID != nil && *notification.MemberID == userID {
			results = append(results, notification)
		}
	}

	// Apply pagination
	if pagination != nil {
		start := pagination.Page * pagination.Limit
		end := start + pagination.Limit

		if start < int64(len(results)) {
			if end > int64(len(results)) {
				end = int64(len(results))
			}
			results = results[start:end]
		} else {
			results = []*db.NotificationDomain{}
		}
	}

	return results, int64(len(results)), nil
}
