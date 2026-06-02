package handlers

import (
	"net/http"
	"sts-backend/internal/models"
	"sts-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetBookings retrieves all bus bookings
func GetBookings(c *gin.Context) {
	bookings, err := services.GetAllBookings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// CreateBooking creates a new bus booking
func CreateBooking(c *gin.Context) {
	var booking models.BusBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.CreateBooking(&booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBookingByID retrieves a single booking by ID
func GetBookingByID(c *gin.Context) {
	id := c.Param("id")
	booking, err := services.GetBookingByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	c.JSON(http.StatusOK, booking)
}

// UpdateBookingStatusRequest represents the request body for updating booking status
type UpdateBookingStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateBookingStatus updates the booking status
func UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdateBookingStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}

// UpdatePaymentStatusRequest represents the request body for updating payment status
type UpdatePaymentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdatePaymentStatus updates the payment status
func UpdatePaymentStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdatePaymentStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment status updated successfully"})
}

// UpdateBooking updates a complete booking record
func UpdateBooking(c *gin.Context) {
	id := c.Param("id")
	var booking models.BusBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID from the URL matches the booking
	booking.BookingID = id

	err := services.UpdateBooking(&booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch and return the updated booking
	updatedBooking, err := services.GetBookingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Booking updated successfully"})
		return
	}

	c.JSON(http.StatusOK, updatedBooking)
}

// GetBookingsByStatus retrieves bookings filtered by status
func GetBookingsByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	bookings, err := services.GetBookingsByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// SearchBookings searches bookings by reference number or passenger details
func SearchBookings(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query parameter 'q' is required"})
		return
	}

	bookings, err := services.SearchBookings(searchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// CancelBooking cancels a booking
func CancelBooking(c *gin.Context) {
	id := c.Param("id")
	err := services.DeleteBooking(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}
