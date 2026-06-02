package handlers

import (
	"log"
	"net/http"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetLoungeBookings handles GET /api/lounge-bookings
func GetLoungeBookings(c *gin.Context) {
	bookings, err := services.GetAllLoungeBookings()
	if err != nil {
		// Log the detailed error
		log.Printf("Error getting lounge bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// GetLoungeBookingByID handles GET /api/lounge-bookings/:id
func GetLoungeBookingByID(c *gin.Context) {
	id := c.Param("id")
	booking, err := services.GetLoungeBookingByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, booking)
}

// CreateLoungeBooking handles POST /api/lounge-bookings
func CreateLoungeBooking(c *gin.Context) {
	var booking models.LoungeBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set default values if not provided
	if booking.PaymentStatus == "" {
		booking.PaymentStatus = "pending"
	}
	if booking.Status == "" {
		booking.Status = "pending"
	}

	if err := services.CreateLoungeBooking(&booking); err != nil {
		log.Printf("Error creating lounge booking: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Lounge booking created successfully",
		"id":      booking.LoungeBookingID,
	})
}

// UpdateLoungeBooking handles PUT /api/lounge-bookings/:id
func UpdateLoungeBooking(c *gin.Context) {
	id := c.Param("id")

	var booking models.LoungeBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	booking.LoungeBookingID = id

	if err := services.UpdateLoungeBooking(&booking); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lounge booking updated successfully"})
}

// UpdateLoungeBookingPaymentStatus handles PATCH /api/lounge-bookings/:id/payment-status
func UpdateLoungeBookingPaymentStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := services.UpdateLoungeBookingPaymentStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment status updated successfully"})
}

// UpdateLoungeBookingStatus handles PATCH /api/lounge-bookings/:id/booking-status
func UpdateLoungeBookingStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := services.UpdateLoungeBookingStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}

// DeleteLoungeBooking handles DELETE /api/lounge-bookings/:id
func DeleteLoungeBooking(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeleteLoungeBooking(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lounge booking deleted successfully"})
}
