package handlers

import (
	"net/http"
	"strings"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetBuses(c *gin.Context) {
	buses, err := services.GetAllBuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buses)
}

func GetBusById(c *gin.Context) {
	id := c.Param("id")
	bus, err := services.GetBusByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bus == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bus not found"})
		return
	}
	c.JSON(http.StatusOK, bus)
}

func GetPendingBuses(c *gin.Context) {
	buses, err := services.GetPendingBuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buses)
}

func CreateBus(c *gin.Context) {
	var bus models.Bus
	if err := c.ShouldBindJSON(&bus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values if needed
	if bus.VerificationStatus == "" {
		bus.VerificationStatus = "Pending"
	}
	// Convert status to lowercase for database constraint
	if bus.Status == "" || bus.Status == "Active" {
		bus.Status = "active"
	} else {
		bus.Status = strings.ToLower(bus.Status)
	}

	err := services.CreateBus(&bus)
	if err != nil {
		// Log the error for debugging
		println("Error creating bus:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bus)
}

func UpdateBus(c *gin.Context) {
	id := c.Param("id")
	var bus models.Bus
	if err := c.ShouldBindJSON(&bus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bus.ID = id
	
	// Convert status to lowercase for database constraint
	if bus.Status != "" {
		bus.Status = strings.ToLower(bus.Status)
	}
	
	err := services.UpdateBus(&bus)
	if err != nil {
		println("Error updating bus:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bus)
}

type BusVerificationRequest struct {
	Status    string `json:"status"`
	Documents string `json:"documents"`
}

func VerifyBus(c *gin.Context) {
	id := c.Param("id")
	var req BusVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdateBusVerification(id, req.Status, req.Documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bus verification updated"})
}
