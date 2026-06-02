package services

import (
	"fmt"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetAllBusOwners() ([]models.BusOwner, error) {
	return database.GetAllBusOwners()
}

func GetBusOwnerByID(id string) (*models.BusOwner, error) {
	return database.GetBusOwnerByID(id)
}

func GetPendingBusOwners() ([]models.BusOwner, error) {
	return database.GetPendingBusOwners()
}

func CreateBusOwner(owner *models.BusOwner) error {
	owner.VerificationStatus = ensurePendingApprovalStatus(owner.VerificationStatus)

	if err := database.CreateBusOwner(owner); err != nil {
		return err
	}

	if isPendingApprovalStatus(owner.VerificationStatus) {
		notifyApprovalRequest(
			"bus owner",
			fmt.Sprintf("Company: %s", owner.CompanyName),
			fmt.Sprintf("Business email: %s", owner.BusinessEmail),
			fmt.Sprintf("Business phone: %s", owner.BusinessPhone),
			fmt.Sprintf("Identity/registration: %s", owner.IdentityOrIncorporationNo),
		)
	}

	return nil
}

func UpdateBusOwner(owner *models.BusOwner) error {
	return database.UpdateBusOwner(owner)
}

func DeleteBusOwner(id string) error {
	return database.DeleteBusOwner(id)
}

func VerifyBusOwner(id string, status string, documents interface{}) error {
	if err := database.VerifyBusOwner(id, status, documents); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		owner, err := database.GetBusOwnerByID(id)
		if err != nil {
			return err
		}
		if owner != nil {
			notifyApprovalDecision(
				owner.BusinessPhone,
				"bus owner",
				"approved",
				formatDecisionDetail("Company", owner.CompanyName),
				formatDecisionDetail("Email", owner.BusinessEmail),
			)
		}
	}

	return nil
}
