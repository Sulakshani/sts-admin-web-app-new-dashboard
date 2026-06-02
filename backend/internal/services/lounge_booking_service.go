package services

import (
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetAllLoungeBookings() ([]models.LoungeBooking, error) {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.GetLoungeBookings()
}

func GetLoungeBookingByID(id string) (*models.LoungeBooking, error) {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.GetLoungeBookingByID(id)
}

func CreateLoungeBooking(lb *models.LoungeBooking) error {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.CreateLoungeBooking(lb)
}

func UpdateLoungeBooking(lb *models.LoungeBooking) error {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.UpdateLoungeBooking(lb)
}

func UpdateLoungeBookingPaymentStatus(id string, status string) error {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.UpdatePaymentStatus(id, status)
}

func UpdateLoungeBookingStatus(id string, status string) error {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.UpdateBookingStatus(id, status)
}

func DeleteLoungeBooking(id string) error {
	repo := database.NewLoungeBookingRepository(database.DB)
	return repo.DeleteLoungeBooking(id)
}
