package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type plantSlot struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := plantSlot{m}

		v1 := r.Group("/plant-slots/v1")
		{
			// Member endpoints
			v1.GET("/my-slots", s.BearerAuth(enum.PermissionPlantSlotView), s.v1_getMySlots())
			v1.POST("/request", s.BearerAuth(enum.PermissionPlantSlotCreate), s.v1_requestSlots())
			v1.GET("/:id", s.BearerAuth(enum.PermissionPlantSlotView), s.v1_getSlotDetails())
			v1.PUT("/:id/status", s.BearerAuth(enum.PermissionPlantSlotUpdate), s.v1_updateSlotStatus())
			v1.POST("/:id/maintenance", s.BearerAuth(enum.PermissionPlantSlotUpdate), s.v1_reportMaintenance())

			// Transfer functionality
			v1.POST("/transfer", s.BearerAuth(enum.PermissionPlantSlotTransfer), s.v1_transferSlots())

			// Admin endpoints
			admin := v1.Group("/admin")
			{
				admin.GET("/all", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getAllSlots())
				admin.POST("/assign", s.BearerAuth(enum.PermissionPlantSlotAssign), s.v1_assignSlots())
				admin.GET("/maintenance", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getMaintenanceSlots())
				admin.GET("/analytics", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_getSlotAnalytics())
				admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionPlantSlotManage), s.v1_forceStatusUpdate())
			}
		}
	})
}

// Request/Response structures following membership pattern
type SlotRequestRequest struct {
	Quantity          int    `json:"quantity" binding:"required,min=1,max=10"`
	PreferredLocation string `json:"preferred_location" validate:"omitempty"`
}

type TransferSlotsRequest struct {
	ToMemberID string   `json:"to_member_id" binding:"required,len=24"`
	SlotIDs    []string `json:"slot_ids" binding:"required,min=1,dive,len=24"`
	Reason     string   `json:"reason" binding:"required"`
}

type MaintenanceRequest struct {
	Description string `json:"description" binding:"required"`
	Priority    string `json:"priority" validate:"oneof=low normal high"`
}

type UpdateSlotStatusRequest struct {
	Status string  `json:"status" binding:"required" validate:"oneof=available allocated occupied maintenance out_of_service"`
	Reason *string `json:"reason" validate:"omitempty"`
}

type AssignSlotsRequest struct {
	MemberID     string   `json:"member_id" binding:"required,len=24"`
	MembershipID string   `json:"membership_id" binding:"required,len=24"`
	SlotIDs      []string `json:"slot_ids" binding:"required,min=1,dive,len=24"`
	AssignedBy   string   `json:"assigned_by" binding:"required,len=24"`
}

// v1_getMySlots handles GET /plant-slots/v1/my-slots
func (s plantSlot) v1_getMySlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get member's slots
		slots, err := s.store.Db.PlantSlot.FindByMemberID(ctx, db.SID(member.ID))
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Convert to DTO
		slotDtos := make([]*db.PlantSlotBaseDto, len(slots))
		for i, slot := range slots {
			slotDtos[i] = slot.BaseDto()
		}

		c.JSON(http.StatusOK, gin.H{
			"slots": slotDtos,
			"total": len(slotDtos),
		})
	}
}

// v1_requestSlots handles POST /plant-slots/v1/request
func (s plantSlot) v1_requestSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		var req SlotRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify member has active membership
		membership, err := s.store.Db.Membership.FindActiveByMemberID(ctx, db.SID(member.ID))
		if err != nil {
			c.Error(ecode.PlantSlotMembershipRequired.Desc(fmt.Errorf("Active membership required to request plant slots")))
			return
		}

		// Check if member already has allocated slots
		if err := s.store.Db.PlantSlot.ValidateAllocation(ctx, db.SID(member.ID), req.Quantity); err != nil {
			c.Error(err)
			return
		}

		// Allocate slots
		allocatedSlots, err := s.store.Db.PlantSlot.AllocateToMember(ctx,
			db.SID(member.ID),
			db.SID(membership.ID),
			req.Quantity)
		if err != nil {
			c.Error(err)
			return
		}

		// Convert to DTO
		slotDtos := make([]*db.PlantSlotDetailDto, len(allocatedSlots))
		for i, slot := range allocatedSlots {
			slotDtos[i] = slot.DetailDto()
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Plant slots allocated successfully",
			"slots":   slotDtos,
			"total":   len(slotDtos),
		})
	}
}

// v1_getSlotDetails handles GET /plant-slots/v1/:id
func (s plantSlot) v1_getSlotDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		slotID := c.Param("id")

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get slot details
		slot, err := s.store.Db.PlantSlot.FindByID(ctx, slotID)
		if err != nil {
			c.Error(ecode.PlantSlotNotFound.Desc(fmt.Errorf("Plant slot not found")))
			return
		}

		// Verify ownership (or admin)
		if slot.MemberID == nil || *slot.MemberID != db.SID(member.ID) {
			c.Error(ecode.Forbidden.Desc(fmt.Errorf("Not authorized to view this slot")))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"slot": slot.DetailDto(),
		})
	}
}

// v1_updateSlotStatus handles PUT /plant-slots/v1/:id/status
func (s plantSlot) v1_updateSlotStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		slotID := c.Param("id")

		var req UpdateSlotStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get slot and verify ownership
		slot, err := s.store.Db.PlantSlot.FindByID(ctx, slotID)
		if err != nil {
			c.Error(ecode.PlantSlotNotFound.Desc(fmt.Errorf("Plant slot not found")))
			return
		}

		if slot.MemberID == nil || *slot.MemberID != db.SID(member.ID) {
			c.Error(ecode.Forbidden.Desc(fmt.Errorf("Not authorized to update this slot")))
			return
		}

		// Validate status transitions (business logic)
		if !isValidStatusTransition(getValue(slot.Status, ""), req.Status) {
			c.Error(ecode.New(http.StatusBadRequest, "invalid_status_transition").Desc(
				fmt.Errorf("Invalid status transition from %s to %s",
					getValue(slot.Status, ""), req.Status)))
			return
		}

		// Update status
		err = s.store.Db.PlantSlot.UpdateStatus(ctx, slotID, req.Status)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Slot status updated successfully",
			"slot_id":    slotID,
			"new_status": req.Status,
		})
	}
}

// v1_reportMaintenance handles POST /plant-slots/v1/:id/maintenance
func (s plantSlot) v1_reportMaintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		slotID := c.Param("id")

		var req MaintenanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get slot and verify ownership
		slot, err := s.store.Db.PlantSlot.FindByID(ctx, slotID)
		if err != nil {
			c.Error(ecode.PlantSlotNotFound.Desc(fmt.Errorf("Plant slot not found")))
			return
		}

		if slot.MemberID == nil || *slot.MemberID != db.SID(member.ID) {
			c.Error(ecode.Forbidden.Desc(fmt.Errorf("Not authorized to report maintenance for this slot")))
			return
		}

		// Add maintenance log
		err = s.store.Db.PlantSlot.AddMaintenanceLog(ctx, slotID, req.Description, session.UserId)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Maintenance request recorded successfully",
			"slot_id": slotID,
		})
	}
}

// v1_transferSlots handles POST /plant-slots/v1/transfer
func (s plantSlot) v1_transferSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		var req TransferSlotsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Verify recipient member exists and has membership
		_, err = s.store.Db.Member.FindByID(ctx, req.ToMemberID)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "recipient_member_not_found").Desc(err))
			return
		}

		_, err = s.store.Db.Membership.FindActiveByMemberID(ctx, req.ToMemberID)
		if err != nil {
			c.Error(ecode.PlantSlotMembershipRequired.Desc(fmt.Errorf("Recipient member must have active membership")))
			return
		}

		// Perform transfer
		err = s.store.Db.PlantSlot.TransferSlots(ctx, db.SID(member.ID), req.ToMemberID, req.SlotIDs)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":           "Slots transferred successfully",
			"from_member_id":    db.SID(member.ID),
			"to_member_id":      req.ToMemberID,
			"transferred_slots": len(req.SlotIDs),
			"reason":            req.Reason,
		})
	}
}

// Admin endpoints

// v1_getAllSlots handles GET /plant-slots/v1/admin/all
func (s plantSlot) v1_getAllSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
		status := c.Query("status")
		location := c.Query("location")

		// Build query
		query := &db.PlantSlotQuery{
			Query: db.Query{
				Page:  page,
				Limit: limit,
			},
			TenantId: &session.TenantId,
		}

		if status != "" {
			query.Status = &status
		}
		if location != "" {
			query.Location = &location
		}

		query.Build()

		// For MVP - simplified implementation
		count, _ := s.store.Db.PlantSlot.Count(ctx, db.M{
			"tenant_id": session.TenantId,
		})

		c.JSON(http.StatusOK, gin.H{
			"slots": []interface{}{}, // Simplified for MVP
			"total": count,
			"page":  page,
			"limit": limit,
		})
	}
}

// v1_assignSlots handles POST /plant-slots/v1/admin/assign
func (s plantSlot) v1_assignSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var req AssignSlotsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Verify member and membership exist
		_, err := s.store.Db.Member.FindByID(ctx, req.MemberID)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		_, err = s.store.Db.Membership.FindByID(ctx, req.MembershipID)
		if err != nil {
			c.Error(ecode.MembershipNotFound.Desc(err))
			return
		}

		// Assign slots (simplified for MVP)
		assignedCount := 0
		for _, slotID := range req.SlotIDs {
			slot, err := s.store.Db.PlantSlot.FindByID(ctx, slotID)
			if err != nil {
				continue
			}

			if slot.Status != nil && *slot.Status == "available" {
				slot.MemberID = &req.MemberID
				slot.MembershipID = &req.MembershipID
				slot.Status = stringPtr("allocated")

				_, err = s.store.Db.PlantSlot.Save(ctx, slot)
				if err == nil {
					assignedCount++
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Slots assigned successfully",
			"assigned_count": assignedCount,
			"member_id":      req.MemberID,
		})
	}
}

// v1_getMaintenanceSlots handles GET /plant-slots/v1/admin/maintenance
func (s plantSlot) v1_getMaintenanceSlots() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		daysThreshold, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

		// Get slots requiring maintenance
		slots, err := s.store.Db.PlantSlot.FindSlotsRequiringMaintenance(ctx, daysThreshold, session.TenantId)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Convert to DTO
		slotDtos := make([]*db.PlantSlotDetailDto, len(slots))
		for i, slot := range slots {
			slotDtos[i] = slot.DetailDto()
		}

		c.JSON(http.StatusOK, gin.H{
			"slots":          slotDtos,
			"total":          len(slotDtos),
			"days_threshold": daysThreshold,
		})
	}
}

// v1_getSlotAnalytics handles GET /plant-slots/v1/admin/analytics
func (s plantSlot) v1_getSlotAnalytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Basic analytics for MVP
		totalSlots, _ := s.store.Db.PlantSlot.Count(ctx, db.M{
			"tenant_id": session.TenantId,
		})

		allocatedSlots, _ := s.store.Db.PlantSlot.Count(ctx, db.M{
			"tenant_id": session.TenantId,
			"status":    "allocated",
		})

		occupiedSlots, _ := s.store.Db.PlantSlot.Count(ctx, db.M{
			"tenant_id": session.TenantId,
			"status":    "occupied",
		})

		maintenanceSlots, _ := s.store.Db.PlantSlot.Count(ctx, db.M{
			"tenant_id": session.TenantId,
			"status":    "maintenance",
		})

		c.JSON(http.StatusOK, gin.H{
			"total_slots":       totalSlots,
			"allocated_slots":   allocatedSlots,
			"occupied_slots":    occupiedSlots,
			"maintenance_slots": maintenanceSlots,
			"available_slots":   totalSlots - allocatedSlots - occupiedSlots - maintenanceSlots,
			"utilization_rate":  float64(allocatedSlots+occupiedSlots) / float64(totalSlots) * 100,
		})
	}
}

// v1_forceStatusUpdate handles PUT /plant-slots/v1/admin/:id/force-status
func (s plantSlot) v1_forceStatusUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		slotID := c.Param("id")

		var req UpdateSlotStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Admin can force any status change
		err := s.store.Db.PlantSlot.UpdateStatus(ctx, slotID, req.Status)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Slot status updated successfully (admin override)",
			"slot_id":    slotID,
			"new_status": req.Status,
		})
	}
}

// Helper functions

// isValidStatusTransition validates status transitions following business rules
func isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"available":      {"allocated"},
		"allocated":      {"occupied", "available"},
		"occupied":       {"maintenance", "available"},
		"maintenance":    {"available", "out_of_service"},
		"out_of_service": {"maintenance", "available"},
	}

	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// Helper function for string pointer (already exists in plant_slot.go)
func stringPtr(s string) *string {
	return &s
}

// Helper function for getting values with defaults (duplicate from plant_slot.go)
func getValue[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
