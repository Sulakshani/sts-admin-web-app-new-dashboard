package services

import (
	"fmt"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetDrivers() ([]models.Driver, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetDrivers()
}

func GetPendingDrivers() ([]models.Driver, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetPendingDrivers()
}

func GetDriverByID(id string) (*models.Driver, error) {
	repo := database.NewStaffRepository(database.DB)
	return repo.GetDriverByID(id)
}

func CreateDriver(driver *models.Driver) error {
	driver.VerificationStatus = ensurePendingApprovalStatus(driver.VerificationStatus)

	repo := database.NewStaffRepository(database.DB)
	if err := repo.CreateDriver(driver); err != nil {
		return err
	}

	if isPendingApprovalStatus(driver.VerificationStatus) {
		notifyApprovalRequest(
			"driver",
			fmt.Sprintf("Name: %s", driver.Name),
			fmt.Sprintf("Contact: %s", driver.ContactNumber),
			fmt.Sprintf("License number: %s", driver.LicenseNumber),
			fmt.Sprintf("Employment status: %s", driver.Status),
		)
	}

	return nil
}

func UpdateDriver(driver *models.Driver) error {
	repo := database.NewStaffRepository(database.DB)
	return repo.UpdateDriver(driver)
}

func UpdateDriverVerification(id string, status string, documents string) error {
	repo := database.NewStaffRepository(database.DB)
	if err := repo.UpdateDriverVerification(id, status, documents); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		driver, err := repo.GetDriverByID(id)
		if err != nil {
			return err
		}
		if driver != nil {
			notifyApprovalDecision(
				driver.ContactNumber,
				"driver",
				"approved",
				formatDecisionDetail("Name", driver.Name),
				formatDecisionDetail("License", driver.LicenseNumber),
			)
		}
	}

	return nil
}
