package services

import (
	"fmt"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetConductors() ([]models.Conductor, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetConductors()
}

func GetPendingConductors() ([]models.Conductor, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetPendingConductors()
}

func GetConductorByID(id string) (*models.Conductor, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetConductorByID(id)
}

func CreateConductor(conductor *models.Conductor) error {
	conductor.VerificationStatus = ensurePendingApprovalStatus(conductor.VerificationStatus)

	repo := database.NewStaffRepository(database.DB)
	if err := repo.CreateConductor(conductor); err != nil {
		return err
	}

	if isPendingApprovalStatus(conductor.VerificationStatus) {
		notifyApprovalRequest(
			"conductor",
			fmt.Sprintf("Name: %s", conductor.Name),
			fmt.Sprintf("Contact: %s", conductor.ContactNumber),
			fmt.Sprintf("License number: %s", conductor.LicenseNumber),
			fmt.Sprintf("Employment status: %s", conductor.Status),
		)
	}

	return nil
}

func UpdateConductor(conductor *models.Conductor) error {
	repo := database.NewStaffRepository(database.DB)
	return repo.UpdateConductor(conductor)
}

func UpdateConductorVerification(id string, status string, documents string) error {
	repo := database.NewStaffRepository(database.DB)
	if err := repo.UpdateConductorVerification(id, status, documents); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		conductor, err := repo.GetConductorByID(id)
		if err != nil {
			return err
		}
		if conductor != nil {
			notifyApprovalDecision(
				conductor.ContactNumber,
				"conductor",
				"approved",
				formatDecisionDetail("Name", conductor.Name),
				formatDecisionDetail("License", conductor.LicenseNumber),
			)
		}
	}

	return nil
}
