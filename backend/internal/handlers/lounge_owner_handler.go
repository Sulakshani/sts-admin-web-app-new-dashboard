package handlers

import (
	"net/http"
	"strings"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetPendingLoungeOwners(c *gin.Context) {
	owners, err := services.GetPendingLoungeOwners()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, owners)
}

func GetLoungeOwnerById(c *gin.Context) {
	id := c.Param("id")
	owner, err := services.GetLoungeOwnerByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if owner == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lounge owner not found"})
		return
	}

	c.JSON(http.StatusOK, owner)
}

type LoungeOwnerVerificationRequest struct {
	Status             string `json:"status"`
	VerificationStatus string `json:"verification_status"`
	Documents          string `json:"documents"`
	VerificationNotes  string `json:"verification_notes"`
}

func VerifyLoungeOwner(c *gin.Context) {
	id := c.Param("id")
	var req LoungeOwnerVerificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = req.VerificationStatus
	}
	status = strings.ToLower(status)

	notes := req.Documents
	if notes == "" {
		notes = req.VerificationNotes
	}

	if status == "verified" {
		status = "approved"
	}
	if status != "approved" && status != "rejected" && status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification status"})
		return
	}

	err := services.VerifyLoungeOwner(id, status, notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lounge owner verification updated successfully"})
}
