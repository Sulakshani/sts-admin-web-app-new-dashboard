package services

import (
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

// GetAllBookings retrieves all bus bookings using the repository
func GetAllBookings() ([]models.BusBooking, error) {
	repo := database.NewBookingRepository(database.DB)
	return repo.GetAllBookings()
}

// CreateBooking creates a new bus booking
func CreateBooking(booking *models.BusBooking) error {
	repo := database.NewBookingRepository(database.DB)
	return repo.CreateBooking(booking)
}

// GetBookingByID retrieves a booking by ID using the repository
func GetBookingByID(id string) (*models.BusBooking, error) {
	repo := database.NewBookingRepository(database.DB)
	return repo.GetBookingByID(id)
}

// UpdateBookingStatus updates the booking status using the repository
func UpdateBookingStatus(id string, status string) error {
	repo := database.NewBookingRepository(database.DB)
	return repo.UpdateBookingStatus(id, status)
}

// UpdatePaymentStatus updates the payment status using the repository
func UpdatePaymentStatus(id string, status string) error {
	repo := database.NewBookingRepository(database.DB)
	return repo.UpdatePaymentStatus(id, status)
}

// UpdateBooking updates a complete booking record using the repository
func UpdateBooking(booking *models.BusBooking) error {
	repo := database.NewBookingRepository(database.DB)
	return repo.UpdateBooking(booking)
}

// GetBookingsByStatus retrieves bookings by status using the repository
func GetBookingsByStatus(status string) ([]models.BusBooking, error) {
	repo := database.NewBookingRepository(database.DB)
	return repo.GetBookingsByStatus(status)
}

// SearchBookings searches bookings using the repository
func SearchBookings(searchTerm string) ([]models.BusBooking, error) {
	repo := database.NewBookingRepository(database.DB)
	return repo.SearchBookings(searchTerm)
}

// DeleteBooking cancels a booking using the repository
func DeleteBooking(id string) error {
	repo := database.NewBookingRepository(database.DB)
	return repo.DeleteBooking(id)
}
