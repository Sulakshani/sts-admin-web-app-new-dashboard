package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var escalationService *services.EscalationService

// SetEscalationService sets the escalation service for handlers
func SetEscalationService(service *services.EscalationService) {
	escalationService = service
}

func normalizeAccessValue(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func GetComplaints(c *gin.Context) {
	claims, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	pageQuery, hasPage := c.GetQuery("page")
	pageSizeQuery, hasPageSize := c.GetQuery("page_size")
	reporterRole := normalizeAccessValue(c.Query("role"))

	if hasPage || hasPageSize {
		page := 1
		pageSize := 20

		if hasPage {
			if parsed, parseErr := strconv.Atoi(strings.TrimSpace(pageQuery)); parseErr == nil && parsed > 0 {
				page = parsed
			}
		}
		if hasPageSize {
			if parsed, parseErr := strconv.Atoi(strings.TrimSpace(pageSizeQuery)); parseErr == nil && parsed > 0 {
				pageSize = parsed
			}
		}

		if pageSize > 100 {
			pageSize = 100
		}

		paged, pageErr := services.GetComplaintsForAdminPaginated(services.ComplaintAdminContext{
			AdminID:  claims.AdminID,
			Role:     claims.Role,
			AppScope: claims.AppScope,
		}, reporterRole, page, pageSize)
		if pageErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": pageErr.Error()})
			return
		}

		c.JSON(http.StatusOK, paged)
		return
	}

	complaints, err := services.GetComplaintsForAdmin(services.ComplaintAdminContext{
		AdminID:  claims.AdminID,
		Role:     claims.Role,
		AppScope: claims.AppScope,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, complaints)
}

func GetComplaintById(c *gin.Context) {
	id := c.Param("id")
	claims, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	complaint, err := services.GetComplaintByIDForAdmin(id, services.ComplaintAdminContext{
		AdminID:  claims.AdminID,
		Role:     claims.Role,
		AppScope: claims.AppScope,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if complaint == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Complaint not found"})
		return
	}
	c.JSON(http.StatusOK, complaint)
}

func UpdateComplaintStatus(c *gin.Context) {
	id := c.Param("id")
	claims, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Status          string  `json:"status" binding:"required"`
		ResolutionNotes *string `json:"resolution_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := normalizeAccessValue(claims.Role)
	scope := normalizeAccessValue(claims.AppScope)

	if role != "super_admin" {
		if scope != "driver" {
			c.JSON(http.StatusForbidden, gin.H{"error": "only driver app admins can change complaint status"})
			return
		}

		escalation, escErr := escalationService.GetComplaintEscalation(id)
		if escErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": escErr.Error()})
			return
		}
		if escalation == nil {
			if initErr := escalationService.InitializeEscalation(id, ""); initErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": initErr.Error()})
				return
			}
			escalation, escErr = escalationService.GetComplaintEscalation(id)
			if escErr != nil || escalation == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve complaint escalation state"})
				return
			}
		}

		if escalation.SourceApp != "driver" {
			c.JSON(http.StatusForbidden, gin.H{"error": "status changes are limited to driver complaints"})
			return
		}

		expectedTeam := normalizeAccessValue(scope + "_" + role)
		if expectedTeam == "_" || expectedTeam == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the currently assigned app role can change this complaint"})
			return
		}

		if escalation.CurrentTeam != expectedTeam {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the currently assigned app role can change this complaint"})
			return
		}
	}

	err = services.UpdateComplaintStatus(id, req.Status, claims.AdminID, req.ResolutionNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Complaint status updated successfully"})
}

// ManualEscalateComplaint manually escalates a complaint to the next level
func ManualEscalateComplaint(c *gin.Context) {
	if escalationService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Escalation service not initialized"})
		return
	}

	claims, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	role := normalizeAccessValue(claims.Role)
	if role != "super_admin" && role != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only supervisor or super admin can manually escalate"})
		return
	}

	id := c.Param("id")

	var req struct {
		EscalatedBy string `json:"escalated_by"` // Admin user ID or name
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get complaint details in caller scope
	complaint, err := services.GetComplaintByIDForAdmin(id, services.ComplaintAdminContext{
		AdminID:  claims.AdminID,
		Role:     claims.Role,
		AppScope: claims.AppScope,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if complaint == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Complaint not found"})
		return
	}

	// Get current escalation level
	escalation, err := escalationService.GetComplaintEscalation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	currentLevel := 0
	if escalation != nil {
		currentLevel = escalation.CurrentLevel
	}

	// Escalate to next level
	escalatedBy := claims.AdminID
	if req.EscalatedBy != "" {
		escalatedBy = req.EscalatedBy
	}

	err = escalationService.EscalateToNextLevel(id, "", currentLevel, escalatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Complaint escalated successfully",
		"complaint_id": id,
	})
}

// GetComplaintEscalation returns escalation information for a complaint
func GetComplaintEscalation(c *gin.Context) {
	if escalationService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Escalation service not initialized"})
		return
	}

	_, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	escalation, err := escalationService.GetComplaintEscalation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if escalation == nil {
		if initErr := escalationService.InitializeEscalation(id, ""); initErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Escalation info not found"})
			return
		}

		escalation, err = escalationService.GetComplaintEscalation(id)
		if err != nil || escalation == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Escalation info not found"})
			return
		}
	}

	c.JSON(http.StatusOK, escalation)
}

// GetEscalationStats returns escalation statistics
func GetEscalationStats(c *gin.Context) {
	if escalationService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Escalation service not initialized"})
		return
	}

	claims, err := getAuthenticatedAdminClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if normalizeAccessValue(claims.Role) != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin access required"})
		return
	}

	stats, err := escalationService.GetEscalationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

type authAdminClaims struct {
	AdminID  string
	Role     string
	AppScope string
}

func getAuthenticatedAdminClaims(c *gin.Context) (*authAdminClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("authorization bearer token required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	adminID, _ := mapClaims["admin_id"].(string)
	role, _ := mapClaims["role"].(string)
	appScope, _ := mapClaims["app_scope"].(string)
	role = normalizeAccessValue(role)
	appScope = normalizeAccessValue(appScope)
	if adminID == "" || role == "" {
		return nil, fmt.Errorf("missing admin claims")
	}

	return &authAdminClaims{AdminID: adminID, Role: role, AppScope: appScope}, nil
}
