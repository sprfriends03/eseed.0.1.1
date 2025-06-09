package route

import (
	"app/pkg/ecode"
	"app/pkg/enum"
	"app/store/db"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhnghia272/gopkg"
)

type plant struct {
	*middleware
}

func init() {
	handlers = append(handlers, func(m *middleware, r *gin.Engine) {
		s := plant{m}

		v1 := r.Group("/plants/v1")
		{
			// Member endpoints (7 endpoints)
			v1.GET("/my-plants", s.BearerAuth(enum.PermissionPlantView), s.v1_getMyPlants())
			v1.POST("/create", s.BearerAuth(enum.PermissionPlantCreate), s.v1_createPlant())
			v1.GET("/:id", s.BearerAuth(enum.PermissionPlantView), s.v1_getPlantDetails())
			v1.PUT("/:id/status", s.BearerAuth(enum.PermissionPlantUpdate), s.v1_updatePlantStatus())
			v1.PUT("/:id/care", s.BearerAuth(enum.PermissionPlantCare), s.v1_recordCare())
			v1.POST("/:id/images", s.BearerAuth(enum.PermissionPlantUpdate), s.v1_uploadPlantImage())
			v1.POST("/:id/harvest", s.BearerAuth(enum.PermissionPlantHarvest), s.v1_harvestPlant())

			// Admin endpoints (5 endpoints)
			admin := v1.Group("/admin")
			{
				admin.GET("/all", s.BearerAuth(enum.PermissionPlantManage), s.v1_getAllPlants())
				admin.GET("/analytics", s.BearerAuth(enum.PermissionPlantManage), s.v1_getPlantAnalytics())
				admin.GET("/health-alerts", s.BearerAuth(enum.PermissionPlantManage), s.v1_getHealthAlerts())
				admin.PUT("/:id/force-status", s.BearerAuth(enum.PermissionPlantManage), s.v1_forceStatusUpdate())
				admin.GET("/harvest-ready", s.BearerAuth(enum.PermissionPlantManage), s.v1_getHarvestReady())
			}
		}
	})
}

// Request/Response structures following plant_slot pattern
type CreatePlantRequest struct {
	PlantSlotID string `json:"plant_slot_id" binding:"required,len=24"`
	PlantTypeID string `json:"plant_type_id" binding:"required,len=24"`
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Notes       string `json:"notes" validate:"omitempty,max=500"`
}

type UpdatePlantStatusRequest struct {
	Status string  `json:"status" binding:"required" validate:"oneof=seedling vegetative flowering harvested dead"`
	Reason *string `json:"reason" validate:"omitempty,max=255"`
}

type RecordCareRequest struct {
	CareType     string            `json:"care_type" binding:"required" validate:"oneof=watering fertilizing pruning inspection pest_control"`
	Notes        string            `json:"notes" validate:"omitempty,max=500"`
	Measurements *CareMeasurements `json:"measurements" validate:"omitempty"`
	Products     []string          `json:"products" validate:"omitempty,dive,required"`
}

type CareMeasurements struct {
	Temperature *float64 `json:"temperature" validate:"omitempty,gte=-10,lte=50"`
	Humidity    *float64 `json:"humidity" validate:"omitempty,gte=0,lte=100"`
	SoilPH      *float64 `json:"soil_ph" validate:"omitempty,gte=0,lte=14"`
	WaterAmount *float64 `json:"water_amount" validate:"omitempty,gte=0,lte=10000"`
}

type HarvestPlantRequest struct {
	Weight         float64 `json:"weight" binding:"required,gt=0"`
	Quality        int     `json:"quality" binding:"required,gte=1,lte=10"`
	Notes          string  `json:"notes" validate:"omitempty,max=500"`
	ProcessingType string  `json:"processing_type" binding:"required" validate:"oneof=self_process sell_to_seedeg"`
}

type UploadImageRequest struct {
	ImageURL    string `json:"image_url" binding:"required"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

// v1_getMyPlants handles GET /plants/v1/my-plants
func (s plant) v1_getMyPlants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse query parameters
		var query db.PlantQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Set member filter
		query.MemberID = gopkg.Pointer(db.SID(member.ID))
		query.TenantId = gopkg.Pointer(session.TenantId)

		// Get plants with pagination
		plants, err := s.store.Db.Plant.FindAll(ctx, query.Build())
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Convert to DTOs
		plantDtos := make([]*db.PlantBaseDto, len(plants))
		for i, plant := range plants {
			plantDtos[i] = plant.BaseDto()
		}

		// Get total count for pagination
		totalCount, err := s.store.Db.Plant.Count(ctx, query.Filter)
		if err != nil {
			totalCount = int64(len(plants))
		}

		page := int64(1)
		if query.Page > 0 {
			page = query.Page
		}
		limit := int64(20)
		if query.Limit > 0 {
			limit = query.Limit
		}

		c.JSON(http.StatusOK, gin.H{
			"plants": plantDtos,
			"total":  len(plantDtos),
			"count":  totalCount,
			"page":   page,
			"limit":  limit,
		})
	}
}

// v1_createPlant handles POST /plants/v1/create
func (s plant) v1_createPlant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		var req CreatePlantRequest
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

		// Validate plant slot ownership and availability
		plantSlot, err := s.store.Db.PlantSlot.FindByID(ctx, req.PlantSlotID)
		if err != nil {
			c.Error(ecode.PlantSlotNotFound.Desc(err))
			return
		}

		// Verify slot ownership
		if plantSlot.MemberID == nil || *plantSlot.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Plant slot not owned by member")))
			return
		}

		// Verify slot is available
		if plantSlot.Status == nil || *plantSlot.Status != "allocated" {
			c.Error(ecode.PlantSlotOccupied.Desc(fmt.Errorf("Plant slot is not available for planting")))
			return
		}

		// Check if slot already has a plant
		existingPlant, err := s.store.Db.Plant.FindByPlantSlotID(ctx, req.PlantSlotID)
		if err == nil && existingPlant != nil {
			c.Error(ecode.PlantSlotOccupied.Desc(fmt.Errorf("Plant slot already has an active plant")))
			return
		}

		// Validate plant type availability
		plantType, err := s.store.Db.PlantType.FindByID(ctx, req.PlantTypeID)
		if err != nil {
			c.Error(ecode.PlantTypeNotAvailable.Desc(err))
			return
		}

		// Create plant domain
		now := time.Now()
		expectedHarvest := now.AddDate(0, 0, 90) // Default 90 days growth cycle
		if plantType.FloweringTime != nil {
			expectedHarvest = now.AddDate(0, 0, *plantType.FloweringTime)
		}

		plant := &db.PlantDomain{
			PlantTypeID:     &req.PlantTypeID,
			PlantSlotID:     &req.PlantSlotID,
			MemberID:        gopkg.Pointer(db.SID(member.ID)),
			Status:          gopkg.Pointer("seedling"),
			PlantedDate:     &now,
			ExpectedHarvest: &expectedHarvest,
			Name:            &req.Name,
			Health:          gopkg.Pointer(8), // Default healthy rating
			Strain:          plantType.Strain,
			Notes:           &req.Notes,
			TenantId:        &session.TenantId,
		}

		// Save plant
		savedPlant, err := s.store.Db.Plant.Save(ctx, plant)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Update plant slot status to occupied
		err = s.store.Db.PlantSlot.UpdateStatus(ctx, req.PlantSlotID, "occupied")
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to update plant slot status after plant creation: %v\n", err)
		}

		// Audit log
		s.AuditLog(c, "plant", enum.DataActionCreate, savedPlant, savedPlant, db.SID(savedPlant.ID))

		c.JSON(http.StatusCreated, gin.H{
			"message": "Plant created successfully",
			"plant":   savedPlant.DetailDto(),
		})
	}
}

// v1_getPlantDetails handles GET /plants/v1/:id
func (s plant) v1_getPlantDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		// Get member info
		member, err := s.store.Db.Member.FindByUserID(ctx, session.UserId)
		if err != nil {
			c.Error(ecode.New(http.StatusNotFound, "member_not_found").Desc(err))
			return
		}

		// Get plant details
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		// Verify ownership
		if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to view this plant")))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"plant": plant.DetailDto(),
		})
	}
}

// v1_updatePlantStatus handles PUT /plants/v1/:id/status
func (s plant) v1_updatePlantStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		var req UpdatePlantStatusRequest
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

		// Get plant and verify ownership
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to update this plant")))
			return
		}

		// Validate status transition
		currentStatus := gopkg.Value(plant.Status)
		if !isValidPlantStatusTransition(currentStatus, req.Status) {
			c.Error(ecode.PlantLifecycleViolation.Desc(fmt.Errorf("Invalid status transition from %s to %s", currentStatus, req.Status)))
			return
		}

		// Update plant status
		err = s.store.Db.Plant.UpdateStatus(ctx, plantID, req.Status)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// If plant is harvested, update slot status
		if req.Status == "harvested" || req.Status == "dead" {
			err = s.store.Db.PlantSlot.UpdateStatus(ctx, gopkg.Value(plant.PlantSlotID), "available")
			if err != nil {
				fmt.Printf("Failed to update plant slot status after plant harvest/death: %v\n", err)
			}
		}

		// Audit log
		s.AuditLog(c, "plant", enum.DataActionUpdate, req, plant, plantID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Plant status updated successfully",
		})
	}
}

// v1_recordCare handles PUT /plants/v1/:id/care
func (s plant) v1_recordCare() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		var req RecordCareRequest
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

		// Get plant and verify ownership
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to record care for this plant")))
			return
		}

		// Validate plant is not harvested or dead
		if plant.Status != nil && (*plant.Status == "harvested" || *plant.Status == "dead") {
			c.Error(ecode.PlantLifecycleViolation.Desc(fmt.Errorf("Cannot record care for harvested or dead plants")))
			return
		}

		// Create care record (simplified - would need actual CareRecord domain)
		now := time.Now()

		// Update plant's health based on care type
		updateData := map[string]interface{}{
			"updated_at": now,
		}

		// Health improvement logic based on care type
		if plant.Health != nil {
			currentHealth := *plant.Health
			switch req.CareType {
			case "watering":
				if currentHealth < 10 {
					updateData["health"] = currentHealth + 1
				}
			case "fertilizing":
				if currentHealth < 9 {
					updateData["health"] = currentHealth + 2
				}
			case "pest_control":
				if currentHealth < 8 {
					updateData["health"] = currentHealth + 3
				}
			}
		}

		// Update plant
		err = s.store.Db.Plant.UpdateFields(ctx, plantID, updateData)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Audit log
		s.AuditLog(c, "care_record", enum.DataActionCreate, req, plant, plantID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Care record added successfully",
		})
	}
}

// v1_uploadPlantImage handles POST /plants/v1/:id/images
func (s plant) v1_uploadPlantImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		var req UploadImageRequest
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

		// Get plant and verify ownership
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to upload images for this plant")))
			return
		}

		// Add image to plant
		err = s.store.Db.Plant.AddImage(ctx, plantID, req.ImageURL)
		if err != nil {
			c.Error(ecode.PlantImageUploadFailed.Desc(err))
			return
		}

		// Audit log
		s.AuditLog(c, "plant", enum.DataActionUpdate, req, plant, plantID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Image uploaded successfully",
		})
	}
}

// v1_harvestPlant handles POST /plants/v1/:id/harvest
func (s plant) v1_harvestPlant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		var req HarvestPlantRequest
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

		// Get plant and verify ownership
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		if plant.MemberID == nil || *plant.MemberID != db.SID(member.ID) {
			c.Error(ecode.PlantUnauthorizedOwner.Desc(fmt.Errorf("Not authorized to harvest this plant")))
			return
		}

		// Validate plant is ready for harvest
		if plant.Status == nil || *plant.Status != "flowering" {
			c.Error(ecode.PlantNotReadyForHarvest.Desc(fmt.Errorf("Plant must be in flowering stage to harvest")))
			return
		}

		if plant.ExpectedHarvest != nil && time.Now().Before(*plant.ExpectedHarvest) {
			c.Error(ecode.PlantNotReadyForHarvest.Desc(fmt.Errorf("Plant is not yet ready for harvest")))
			return
		}

		// Update plant with harvest information
		now := time.Now()
		err = s.store.Db.Plant.UpdateFields(ctx, plantID, map[string]interface{}{
			"status":         "harvested",
			"actual_harvest": now,
			"updated_at":     now,
		})
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Update plant slot status to available
		if plant.PlantSlotID != nil {
			err = s.store.Db.PlantSlot.UpdateStatus(ctx, *plant.PlantSlotID, "available")
			if err != nil {
				fmt.Printf("Failed to update plant slot status after harvest: %v\n", err)
			}
		}

		// Audit log
		s.AuditLog(c, "harvest", enum.DataActionCreate, req, plant, plantID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Plant harvested successfully",
		})
	}
}

// Admin endpoints

// v1_getAllPlants handles GET /plants/v1/admin/all
func (s plant) v1_getAllPlants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse query parameters
		var query db.PlantQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query.TenantId = &session.TenantId

		// Get plants with pagination
		plants, err := s.store.Db.Plant.FindAll(ctx, query.Build())
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// Convert to DTOs
		plantDtos := make([]*db.PlantDetailDto, len(plants))
		for i, plant := range plants {
			plantDtos[i] = plant.DetailDto()
		}

		// Get total count for pagination
		totalCount, err := s.store.Db.Plant.Count(ctx, query.Filter)
		if err != nil {
			totalCount = int64(len(plants))
		}

		adminPage := int64(1)
		if query.Page > 0 {
			adminPage = query.Page
		}
		adminLimit := int64(20)
		if query.Limit > 0 {
			adminLimit = query.Limit
		}

		c.JSON(http.StatusOK, gin.H{
			"plants": plantDtos,
			"total":  len(plantDtos),
			"count":  totalCount,
			"page":   adminPage,
			"limit":  adminLimit,
		})
	}
}

// v1_getPlantAnalytics handles GET /plants/v1/admin/analytics
func (s plant) v1_getPlantAnalytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse query parameters
		var query db.PlantAnalyticsQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		query.TenantId = &session.TenantId

		// Get plant statistics
		stats := map[string]interface{}{}

		// Total plants by status
		statusStats, err := s.store.Db.Plant.GetStatusStatistics(ctx, *query.TenantId, query.MemberID)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		stats["status_distribution"] = statusStats

		// Health distribution
		healthStats, err := s.store.Db.Plant.GetHealthStatistics(ctx, *query.TenantId, query.MemberID)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		stats["health_distribution"] = healthStats

		// Strain popularity
		strainStats, err := s.store.Db.Plant.GetStrainStatistics(ctx, *query.TenantId, query.MemberID)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		stats["strain_popularity"] = strainStats

		// Growth cycle metrics
		cycleStats, err := s.store.Db.Plant.GetGrowthCycleMetrics(ctx, *query.TenantId, query.TimeRange)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		stats["growth_metrics"] = cycleStats

		// Upcoming harvests
		upcomingHarvests, err := s.store.Db.Plant.GetUpcomingHarvests(ctx, *query.TenantId, 30) // Next 30 days
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}
		stats["upcoming_harvests"] = upcomingHarvests

		c.JSON(http.StatusOK, gin.H{
			"analytics":    stats,
			"generated_at": time.Now(),
		})
	}
}

// v1_getHealthAlerts handles GET /plants/v1/admin/health-alerts
func (s plant) v1_getHealthAlerts() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Get health alerts
		alerts, err := s.store.Db.Plant.GetHealthAlerts(ctx, session.TenantId)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"alerts": alerts,
			"total":  len(alerts),
		})
	}
}

// v1_forceStatusUpdate handles PUT /plants/v1/admin/:id/force-status
func (s plant) v1_forceStatusUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)
		plantID := c.Param("id")

		var req UpdatePlantStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(ecode.BadRequest.Desc(err))
			return
		}

		// Get plant
		plant, err := s.store.Db.Plant.FindByID(ctx, plantID)
		if err != nil {
			c.Error(ecode.PlantNotFound.Desc(err))
			return
		}

		// Verify tenant
		if plant.TenantId == nil || *plant.TenantId != session.TenantId {
			c.Error(ecode.Forbidden.Desc(fmt.Errorf("Plant not found in tenant")))
			return
		}

		// Admin can force any status change
		err = s.store.Db.Plant.UpdateStatus(ctx, plantID, req.Status)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		// If plant is harvested, update slot status
		if req.Status == "harvested" || req.Status == "dead" {
			if plant.PlantSlotID != nil {
				err = s.store.Db.PlantSlot.UpdateStatus(ctx, *plant.PlantSlotID, "available")
				if err != nil {
					fmt.Printf("Failed to update plant slot status after admin force status: %v\n", err)
				}
			}
		}

		// Audit log
		s.AuditLog(c, "plant", enum.DataActionUpdate, req, plant, plantID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Plant status forced successfully",
		})
	}
}

// v1_getHarvestReady handles GET /plants/v1/admin/harvest-ready
func (s plant) v1_getHarvestReady() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session := s.Session(c)

		// Parse optional days ahead parameter
		daysAhead := 7 // Default to next 7 days
		if daysParam := c.Query("days_ahead"); daysParam != "" {
			if parsed, err := strconv.Atoi(daysParam); err == nil && parsed > 0 {
				daysAhead = parsed
			}
		}

		// Get plants ready for harvest
		harvestReady, err := s.store.Db.Plant.GetUpcomingHarvests(ctx, session.TenantId, daysAhead)
		if err != nil {
			c.Error(ecode.InternalServerError.Desc(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"harvest_ready": harvestReady,
			"total":         len(harvestReady),
			"days_ahead":    daysAhead,
		})
	}
}

// Helper functions

// isValidPlantStatusTransition validates status transitions following plant lifecycle
func isValidPlantStatusTransition(current, new string) bool {
	validTransitions := map[string][]string{
		"seedling":   {"vegetative", "dead"},
		"vegetative": {"flowering", "dead"},
		"flowering":  {"harvested", "dead"},
		"harvested":  {"seedling"}, // New cycle (reusing slot)
		"dead":       {"seedling"}, // Replacement
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}
