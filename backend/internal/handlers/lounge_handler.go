package handlers

import (
	"net/http"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetLounges(c *gin.Context) {
	lounges, err := services.GetAllLounges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lounges)
}

func GetPendingLounges(c *gin.Context) {
	lounges, err := services.GetPendingLounges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lounges)
}

func GetLoungeById(c *gin.Context) {
	id := c.Param("id")
	lounge, err := services.GetLoungeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if lounge == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lounge not found"})
		return
	}
	c.JSON(http.StatusOK, lounge)
}

func CreateLounge(c *gin.Context) {
	var l models.Lounge
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := services.CreateLounge(l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Lounge created successfully"})
}

func UpdateLounge(c *gin.Context) {
	id := c.Param("id")
	var l models.Lounge
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.LoungeID = id
	if err := services.UpdateLounge(l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lounge updated successfully"})
}

func DeleteLounge(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteLounge(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lounge deleted successfully"})
}

type LoungeVerificationRequest struct {
	Status    string `json:"status"`
	Documents string `json:"documents"`
}

func VerifyLounge(c *gin.Context) {
	id := c.Param("id")
	var req LoungeVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdateLoungeVerification(id, req.Status, req.Documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lounge verification updated"})
}
