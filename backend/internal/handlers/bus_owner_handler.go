package handlers

import (
	"fmt"
	"net/http"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetBusOwners(c *gin.Context) {
	owners, err := services.GetAllBusOwners()
	if err != nil {
		fmt.Printf("ERROR in GetBusOwners: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, owners)
}

func GetBusOwnerById(c *gin.Context) {
	id := c.Param("id")
	owner, err := services.GetBusOwnerByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if owner == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bus owner not found"})
		return
	}
	c.JSON(http.StatusOK, owner)
}

func GetPendingBusOwners(c *gin.Context) {
	owners, err := services.GetPendingBusOwners()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, owners)
}

func CreateBusOwner(c *gin.Context) {
	var owner models.BusOwner
	if err := c.ShouldBindJSON(&owner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if owner.VerificationStatus == "" {
		owner.VerificationStatus = "pending"
	}
	if owner.Country == "" {
		owner.Country = "Sri Lanka"
	}
	if owner.TotalBuses == 0 {
		owner.TotalBuses = 0
	}

	err := services.CreateBusOwner(&owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, owner)
}

func UpdateBusOwner(c *gin.Context) {
	id := c.Param("id")
	var owner models.BusOwner
	if err := c.ShouldBindJSON(&owner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner.ID = id

	err := services.UpdateBusOwner(&owner)
	if err != nil {
		fmt.Printf("ERROR in UpdateBusOwner: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, owner)
}

func DeleteBusOwner(c *gin.Context) {
	id := c.Param("id")

	err := services.DeleteBusOwner(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bus owner deleted successfully"})
}

type BusOwnerVerificationRequest struct {
	Status                string      `json:"status"`
	VerificationStatus    string      `json:"verification_status"`
	Documents             interface{} `json:"documents"`
	VerificationDocuments interface{} `json:"verification_documents"`
}

func VerifyBusOwner(c *gin.Context) {
	id := c.Param("id")
	var req BusOwnerVerificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Support both status and verification_status fields
	status := req.Status
	if status == "" {
		status = req.VerificationStatus
	}

	// Support both documents and verification_documents fields
	documents := req.Documents
	if documents == nil {
		documents = req.VerificationDocuments
	}

	// Validate status
	if status != "verified" && status != "rejected" && status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification status"})
		return
	}

	err := services.VerifyBusOwner(id, status, documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bus owner verification updated successfully"})
}
