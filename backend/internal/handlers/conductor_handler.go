package handlers

import (
	"net/http"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetConductors(c *gin.Context) {
	conductors, err := services.GetConductors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conductors)
}

func GetPendingConductors(c *gin.Context) {
	conductors, err := services.GetPendingConductors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conductors)
}

func GetConductorById(c *gin.Context) {
	id := c.Param("id")
	conductor, err := services.GetConductorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conductor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conductor not found"})
		return
	}
	c.JSON(http.StatusOK, conductor)
}

func CreateConductor(c *gin.Context) {
	var conductor models.Conductor
	if err := c.ShouldBindJSON(&conductor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.CreateConductor(&conductor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conductor)
}

func UpdateConductor(c *gin.Context) {
	id := c.Param("id")
	var conductor models.Conductor
	if err := c.ShouldBindJSON(&conductor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conductor.ID = id
	err := services.UpdateConductor(&conductor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conductor)
}

type ConductorVerificationRequest struct {
	Status    string `json:"status"`
	Documents string `json:"documents"`
}

func VerifyConductor(c *gin.Context) {
	id := c.Param("id")
	var req ConductorVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdateConductorVerification(id, req.Status, req.Documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conductor verification updated"})
}
