package handlers

import (
	"database/sql"
	"net/http"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type EscalationHandler struct {
	escalationService *services.EscalationService
}

func NewEscalationHandler(db *sql.DB, cfg *config.Config) *EscalationHandler {
	return &EscalationHandler{
		escalationService: services.NewEscalationService(db, cfg),
	}
}

// GetEscalationConfig returns the escalation configuration for all categories
// GET /api/escalation/config
func (h *EscalationHandler) GetEscalationConfig(c *gin.Context) {
	configs := models.GetEscalationConfigs()
	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
	})
}

// GetEscalationConfigForCategory returns escalation config for a specific category
// GET /api/escalation/config/:category
func (h *EscalationHandler) GetEscalationConfigForCategory(c *gin.Context) {
	category := c.Param("category")
	
	config := models.GetEscalationConfigForCategory(category)
	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// GetComplaintEscalation returns escalation info for a specific complaint
// GET /api/escalation/complaint/:id
func (h *EscalationHandler) GetComplaintEscalation(c *gin.Context) {
	complaintID := c.Param("id")
	
	escalation, err := h.escalationService.GetComplaintEscalation(complaintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get escalation info",
		})
		return
	}
	
	if escalation == nil {
		if initErr := h.escalationService.InitializeEscalation(complaintID, ""); initErr != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No escalation record found",
			})
			return
		}

		escalation, err = h.escalationService.GetComplaintEscalation(complaintID)
		if err != nil || escalation == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No escalation record found",
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, escalation)
}

// EscalateComplaint manually escalates a complaint to the next level
// POST /api/escalation/complaint/:id/escalate
func (h *EscalationHandler) EscalateComplaint(c *gin.Context) {
	complaintID := c.Param("id")
	
	var request struct {
		Category     string `json:"category" binding:"required"`
		CurrentLevel int    `json:"current_level" binding:"required"`
		EscalatedBy  string `json:"escalated_by"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}
	
	escalatedBy := request.EscalatedBy
	if escalatedBy == "" {
		escalatedBy = "admin" // Default to admin if not specified
	}
	
	err := h.escalationService.EscalateToNextLevel(
		complaintID, 
		request.Category, 
		request.CurrentLevel, 
		escalatedBy,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Complaint escalated successfully",
	})
}

// AssignComplaint assigns a complaint to an app role/scope team
// POST /api/escalation/complaint/:id/assign
func (h *EscalationHandler) AssignComplaint(c *gin.Context) {
	complaintID := c.Param("id")
	
	var request struct {
		AppRole  string `json:"app_role"`
		AppScope string `json:"app_scope"`
		AdminID  string `json:"admin_id"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if request.AppRole == "" && request.AdminID != "" {
		var role, scope string
		err := database.DB.QueryRow(`
			SELECT COALESCE(role, ''), COALESCE(app_scope, '')
			FROM admin_users
			WHERE id::text = $1
		`, request.AdminID).Scan(&role, &scope)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin_id for assignment"})
			return
		}
		request.AppRole = role
		if request.AppScope == "" {
			request.AppScope = scope
		}
	}

	if request.AppRole == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_role is required"})
		return
	}
	
	err := h.escalationService.AssignComplaintToRole(complaintID, request.AppRole, request.AppScope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Complaint assigned successfully by role",
	})
}

// GetEscalationHistory returns the escalation history for a complaint
// GET /api/escalation/complaint/:id/history
func (h *EscalationHandler) GetEscalationHistory(c *gin.Context) {
	complaintID := c.Param("id")
	
	history, err := h.escalationService.GetEscalationHistory(complaintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get escalation history",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// GetEscalationStats returns escalation statistics
// GET /api/escalation/stats
func (h *EscalationHandler) GetEscalationStats(c *gin.Context) {
	stats, err := h.escalationService.GetEscalationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get escalation stats",
		})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// InitializeComplaintEscalation initializes escalation for a new complaint
// POST /api/escalation/complaint/:id/initialize
func (h *EscalationHandler) InitializeComplaintEscalation(c *gin.Context) {
	complaintID := c.Param("id")
	
	var request struct {
		Category string `json:"category" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}
	
	err := h.escalationService.InitializeEscalation(complaintID, request.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Escalation initialized successfully",
	})
}
