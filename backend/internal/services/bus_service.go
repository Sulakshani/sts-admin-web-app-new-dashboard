package services

import (
	"fmt"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

// GetAllBuses retrieves all buses using the repository
func GetAllBuses() ([]models.Bus, error) {
	repo := database.NewBusRepository(database.DB)
	return repo.GetAllBuses()
}

// GetPendingBuses retrieves all pending buses using the repository
func GetPendingBuses() ([]models.Bus, error) {
	repo := database.NewBusRepository(database.DB)
	return repo.GetPendingBuses()
}

// CreateBus creates a new bus using the repository
func CreateBus(bus *models.Bus) error {
	bus.VerificationStatus = ensurePendingApprovalStatus(bus.VerificationStatus)

	repo := database.NewBusRepository(database.DB)
	if err := repo.CreateBus(bus); err != nil {
		return err
	}

	if isPendingApprovalStatus(bus.VerificationStatus) {
		notifyApprovalRequest(
			"bus",
			fmt.Sprintf("Bus number: %s", bus.BusNumber),
			fmt.Sprintf("Company: %s", bus.CompanyName),
			fmt.Sprintf("Permit number: %s", bus.PermitNumber),
			fmt.Sprintf("License plate: %s", bus.LicensePlate),
		)
	}

	return nil
}

// UpdateBus updates an existing bus using the repository
func UpdateBus(bus *models.Bus) error {
	repo := database.NewBusRepository(database.DB)
	return repo.UpdateBus(bus)
}

// DeleteBus deletes a bus using the repository
func DeleteBus(id string) error {
	repo := database.NewBusRepository(database.DB)
	return repo.DeleteBus(id)
}

// GetBusByID retrieves a bus by ID using the repository
func GetBusByID(id string) (*models.Bus, error) {
	repo := database.NewBusRepository(database.DB)
	return repo.GetBusByID(id)
}

// UpdateBusVerification updates the verification status of a bus
func UpdateBusVerification(id string, status string, documents string) error {
	repo := database.NewBusRepository(database.DB)
	if err := repo.UpdateBusVerification(id, status, documents); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		bus, err := repo.GetBusByID(id)
		if err != nil {
			return err
		}
		if bus != nil {
			notifyApprovalDecision(
				bus.BusinessPhone,
				"bus",
				"approved",
				formatDecisionDetail("Bus number", bus.BusNumber),
				formatDecisionDetail("Company", bus.CompanyName),
			)
		}
	}

	return nil
}
