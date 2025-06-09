package db

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/pkg/validate"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlantSlotDomain struct {
	BaseDomain   `bson:",inline"`
	SlotNumber   *int    `json:"slot_number" bson:"slot_number" validate:"required,gte=1"`      // Unique slot number
	MemberID     *string `json:"member_id" bson:"member_id" validate:"required,len=24"`         // Link to Member
	MembershipID *string `json:"membership_id" bson:"membership_id" validate:"required,len=24"` // Link to Membership
	Status       *string `json:"status" bson:"status" validate:"required"`                      // available, occupied, maintenance
	Location     *string `json:"location" bson:"location" validate:"required"`                  // greenhouse-1, greenhouse-2, etc.
	Position     *struct {
		Row    *int `json:"row" bson:"row" validate:"required,gte=0"`
		Column *int `json:"column" bson:"column" validate:"required,gte=0"`
	} `json:"position" bson:"position" validate:"required"`
	Notes          *string `json:"notes" bson:"notes" validate:"omitempty"`
	MaintenanceLog *[]struct {
		Date        *time.Time `json:"date" bson:"date" validate:"required"`
		Description *string    `json:"description" bson:"description" validate:"required"`
		PerformedBy *string    `json:"performed_by" bson:"performed_by" validate:"required,len=24"` // Staff member ID
	} `json:"maintenance_log" bson:"maintenance_log" validate:"omitempty,dive"`
	LastCleanDate *time.Time   `json:"last_clean_date" bson:"last_clean_date" validate:"omitempty"`
	TenantId      *enum.Tenant `json:"tenant_id" bson:"tenant_id" validate:"required,len=24"`
}

func (s *PlantSlotDomain) Validate() error {
	s.BeforeSave()
	return validate.New().Validate(s)
}

// DTO structures following MembershipDomain pattern
type PlantSlotBaseDto struct {
	ID         string `json:"id"`
	SlotNumber int    `json:"slot_number"`
	MemberID   string `json:"member_id,omitempty"`
	Status     string `json:"status"`
	Location   string `json:"location"`
	Position   *struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	} `json:"position"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlantSlotDetailDto struct {
	ID           string `json:"id"`
	SlotNumber   int    `json:"slot_number"`
	MemberID     string `json:"member_id,omitempty"`
	MembershipID string `json:"membership_id,omitempty"`
	Status       string `json:"status"`
	Location     string `json:"location"`
	Position     *struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	} `json:"position"`
	Notes          string `json:"notes,omitempty"`
	MaintenanceLog []struct {
		Date        time.Time `json:"date"`
		Description string    `json:"description"`
		PerformedBy string    `json:"performed_by"`
	} `json:"maintenance_log,omitempty"`
	LastCleanDate time.Time `json:"last_clean_date,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PlantSlotExpiringDto struct {
	ID                 string    `json:"id"`
	SlotNumber         int       `json:"slot_number"`
	MemberID           string    `json:"member_id"`
	MemberEmail        string    `json:"member_email,omitempty"`
	Status             string    `json:"status"`
	Location           string    `json:"location"`
	LastCleanDate      time.Time `json:"last_clean_date,omitempty"`
	DaysSinceLastClean int       `json:"days_since_last_clean"`
}

// DTO methods following MembershipDomain pattern
func (s PlantSlotDomain) BaseDto() *PlantSlotBaseDto {
	dto := &PlantSlotBaseDto{
		ID:         SID(s.ID),
		SlotNumber: getValue(s.SlotNumber, 0),
		Status:     getValue(s.Status, ""),
		Location:   getValue(s.Location, ""),
		UpdatedAt:  getValue(s.UpdatedAt, time.Time{}),
	}

	if s.MemberID != nil {
		dto.MemberID = *s.MemberID
	}

	if s.Position != nil {
		dto.Position = &struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		}{
			Row:    getValue(s.Position.Row, 0),
			Column: getValue(s.Position.Column, 0),
		}
	}

	return dto
}

func (s PlantSlotDomain) DetailDto() *PlantSlotDetailDto {
	dto := &PlantSlotDetailDto{
		ID:         SID(s.ID),
		SlotNumber: getValue(s.SlotNumber, 0),
		Status:     getValue(s.Status, ""),
		Location:   getValue(s.Location, ""),
		Notes:      getValue(s.Notes, ""),
		CreatedAt:  getValue(s.CreatedAt, time.Time{}),
		UpdatedAt:  getValue(s.UpdatedAt, time.Time{}),
	}

	if s.MemberID != nil {
		dto.MemberID = *s.MemberID
	}

	if s.MembershipID != nil {
		dto.MembershipID = *s.MembershipID
	}

	if s.Position != nil {
		dto.Position = &struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		}{
			Row:    getValue(s.Position.Row, 0),
			Column: getValue(s.Position.Column, 0),
		}
	}

	if s.MaintenanceLog != nil {
		dto.MaintenanceLog = make([]struct {
			Date        time.Time `json:"date"`
			Description string    `json:"description"`
			PerformedBy string    `json:"performed_by"`
		}, len(*s.MaintenanceLog))

		for i, log := range *s.MaintenanceLog {
			dto.MaintenanceLog[i] = struct {
				Date        time.Time `json:"date"`
				Description string    `json:"description"`
				PerformedBy string    `json:"performed_by"`
			}{
				Date:        getValue(log.Date, time.Time{}),
				Description: getValue(log.Description, ""),
				PerformedBy: getValue(log.PerformedBy, ""),
			}
		}
	}

	if s.LastCleanDate != nil {
		dto.LastCleanDate = *s.LastCleanDate
	}

	return dto
}

// Helper function for getting values with defaults
func getValue[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// PlantSlotQuery following MembershipQuery pattern
type PlantSlotQuery struct {
	Query               `bson:",inline"`
	Status              *string      `json:"status" form:"status"`
	MemberID            *string      `json:"member_id" form:"member_id"`
	MembershipID        *string      `json:"membership_id" form:"membership_id"`
	Location            *string      `json:"location" form:"location"`
	AvailableOnly       *bool        `json:"available_only" form:"available_only"`
	MaintenanceRequired *bool        `json:"maintenance_required" form:"maintenance_required"`
	TenantId            *enum.Tenant `json:"tenant_id" form:"tenant_id"`
}

func (s *PlantSlotQuery) Build() *PlantSlotQuery {
	query := Query{
		Page:   s.Page,
		Limit:  s.Limit,
		Sorts:  s.Sorts,
		Filter: M{},
	}

	if s.Status != nil {
		query.Filter["status"] = *s.Status
	}
	if s.MemberID != nil {
		query.Filter["member_id"] = *s.MemberID
	}
	if s.MembershipID != nil {
		query.Filter["membership_id"] = *s.MembershipID
	}
	if s.Location != nil {
		query.Filter["location"] = *s.Location
	}
	if s.AvailableOnly != nil && *s.AvailableOnly {
		query.Filter["status"] = "available"
	}
	if s.MaintenanceRequired != nil && *s.MaintenanceRequired {
		query.Filter["status"] = "maintenance"
	}
	if s.TenantId != nil {
		query.Filter["tenant_id"] = *s.TenantId
	}

	s.Query = query
	return s
}

type plantSlot struct {
	repo *repo
}

func newPlantSlot(ctx context.Context, collection *mongo.Collection) *plantSlot {
	// Set up indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "member_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "membership_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "location", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "slot_number", Value: 1},
				{Key: "tenant_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "position.row", Value: 1},
				{Key: "position.column", Value: 1},
				{Key: "location", Value: 1},
				{Key: "tenant_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logrus.Errorln("Failed to create plant slot indexes:", err)
	}

	return &plantSlot{repo: newrepo(collection)}
}

// Enhanced business logic methods following task requirements
func (s *plantSlot) AllocateToMember(ctx context.Context, memberID string, membershipID string, quantity int) ([]*PlantSlotDomain, error) {
	// Validate allocation capacity first
	if err := s.ValidateAllocation(ctx, memberID, quantity); err != nil {
		return nil, err
	}

	// Find available slots
	availableSlots, err := s.FindByStatus(ctx, "available", "")
	if err != nil {
		return nil, err
	}

	if len(availableSlots) < quantity {
		return nil, NewPlantSlotError("plant_slot_insufficient_slots")
	}

	// Allocate the requested number of slots
	allocatedSlots := make([]*PlantSlotDomain, 0, quantity)

	for i := 0; i < quantity && i < len(availableSlots); i++ {
		slot := availableSlots[i]
		slot.MemberID = &memberID
		slot.MembershipID = &membershipID
		slot.Status = stringPtr("allocated")

		savedSlot, err := s.Save(ctx, slot)
		if err != nil {
			// Rollback previously allocated slots on error
			for _, allocated := range allocatedSlots {
				s.UpdateStatus(ctx, SID(allocated.ID), "available")
			}
			return nil, err
		}
		allocatedSlots = append(allocatedSlots, savedSlot)
	}

	return allocatedSlots, nil
}

func (s *plantSlot) ValidateAllocation(ctx context.Context, memberID string, requestedSlots int) error {
	// Check if member already has allocated slots
	existingSlots, err := s.FindByMemberID(ctx, memberID)
	if err != nil {
		return err
	}

	activeSlots := 0
	for _, slot := range existingSlots {
		if slot.Status != nil && (*slot.Status == "allocated" || *slot.Status == "occupied") {
			activeSlots++
		}
	}

	if activeSlots > 0 {
		return NewPlantSlotError("plant_slot_already_allocated")
	}

	return nil
}

func (s *plantSlot) ReleaseSlots(ctx context.Context, membershipID string) error {
	slots, err := s.FindByMembershipID(ctx, membershipID)
	if err != nil {
		return err
	}

	for _, slot := range slots {
		// Only release slots that are not occupied (have plants)
		if slot.Status != nil && *slot.Status != "occupied" {
			slot.MemberID = nil
			slot.MembershipID = nil
			slot.Status = stringPtr("available")
			slot.Notes = stringPtr("Released due to membership expiry")

			_, err := s.Save(ctx, slot)
			if err != nil {
				logrus.Errorf("Failed to release slot %s: %v", SID(slot.ID), err)
				continue
			}
		}
	}

	return nil
}

func (s *plantSlot) TransferSlots(ctx context.Context, fromMemberID, toMemberID string, slotIDs []string) error {
	// Validate that toMemberID has an active membership
	// This would typically call membership service, but for now we'll add a placeholder

	for _, slotID := range slotIDs {
		slot, err := s.FindByID(ctx, slotID)
		if err != nil {
			return err
		}

		// Verify ownership
		if slot.MemberID == nil || *slot.MemberID != fromMemberID {
			return NewPlantSlotError("slot_not_owned")
		}

		// Cannot transfer occupied slots
		if slot.Status != nil && *slot.Status == "occupied" {
			return NewPlantSlotError("plant_slot_occupied_cannot_transfer")
		}

		// Transfer ownership
		slot.MemberID = &toMemberID

		_, err = s.Save(ctx, slot)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *plantSlot) FindSlotsRequiringMaintenance(ctx context.Context, daysThreshold int, tenantID enum.Tenant) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	thresholdDate := time.Now().AddDate(0, 0, -daysThreshold)

	query := Query{
		Filter: M{
			"tenant_id": tenantID,
			"$or": []M{
				{
					"status": "maintenance",
				},
				{
					"last_clean_date": M{"$lt": thresholdDate},
					"status":          M{"$ne": "available"},
				},
			},
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper error function using ecode package
func NewPlantSlotError(code string) error {
	switch code {
	case "plant_slot_insufficient_slots":
		return ecode.PlantSlotInsufficientSlots
	case "plant_slot_already_allocated":
		return ecode.PlantSlotAlreadyAllocated
	case "slot_not_owned":
		return ecode.PlantSlotNotFound
	case "plant_slot_occupied_cannot_transfer":
		return ecode.PlantSlotOccupiedCannotTransfer
	default:
		return ecode.InternalServerError
	}
}

func (s *plantSlot) Save(ctx context.Context, domain *PlantSlotDomain, opts ...*options.UpdateOptions) (*PlantSlotDomain, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}

	if domain.ID.IsZero() {
		domain.ID = primitive.NewObjectID()
	}

	id, err := s.repo.Save(ctx, domain.ID, domain, opts...)
	if err != nil {
		return nil, err
	}
	domain.ID = id

	return s.FindByID(ctx, SID(id))
}

func (s *plantSlot) Create(ctx context.Context, domain *PlantSlotDomain) error {
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plantSlot) Update(ctx context.Context, id string, domain *PlantSlotDomain) error {
	domain.ID = OID(id)
	_, err := s.Save(ctx, domain)
	return err
}

func (s *plantSlot) FindByID(ctx context.Context, id string) (*PlantSlotDomain, error) {
	var domain PlantSlotDomain
	err := s.repo.FindOne(ctx, M{"_id": OID(id)}, &domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *plantSlot) FindByMemberID(ctx context.Context, memberID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"member_id": memberID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByMembershipID(ctx context.Context, membershipID string) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{"membership_id": membershipID},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByStatus(ctx context.Context, status string, tenantID enum.Tenant) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{
			"status":    status,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) FindByLocation(ctx context.Context, location string, tenantID enum.Tenant) ([]*PlantSlotDomain, error) {
	var domains []*PlantSlotDomain

	query := Query{
		Filter: M{
			"location":  location,
			"tenant_id": tenantID,
		},
	}

	return domains, s.repo.FindAll(ctx, query, &domains)
}

func (s *plantSlot) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{"$set": M{"status": status, "updated_at": time.Now()}},
	)
}

func (s *plantSlot) AddMaintenanceLog(ctx context.Context, id string, description string, staffID string) error {
	maintenance := M{
		"date":         time.Now(),
		"description":  description,
		"performed_by": staffID,
	}

	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$push": M{"maintenance_log": maintenance},
			"$set": M{
				"status":     "maintenance",
				"updated_at": time.Now(),
			},
		},
	)
}

func (s *plantSlot) MarkCleaned(ctx context.Context, id string) error {
	return s.repo.UpdateOne(ctx,
		M{"_id": OID(id)},
		M{
			"$set": M{
				"last_clean_date": time.Now(),
				"status":          "available",
				"updated_at":      time.Now(),
			},
		},
	)
}

func (s *plantSlot) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteOne(ctx, M{"_id": OID(id)})
}

func (s *plantSlot) Count(ctx context.Context, filter M) (int64, error) {
	return s.repo.CountDocuments(ctx, Query{Filter: filter}), nil
}
