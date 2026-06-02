package services

import (
	"log"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetPendingLoungeOwners() ([]models.LoungeOwner, error) {
	return database.GetPendingLoungeOwners()
}

func GetLoungeOwnerByID(id string) (*models.LoungeOwner, error) {
	return database.GetLoungeOwnerByID(id)
}

func VerifyLoungeOwner(id string, status string, notes string) error {
	if err := database.VerifyLoungeOwner(id, status, notes); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		owner, err := database.GetLoungeOwnerByID(id)
		if err != nil {
			log.Printf("failed to load lounge owner %s after approval: %v", id, err)
			return nil
		}
		if owner != nil {
			notifyApprovalDecision(
				owner.ContactNumber,
				"lounge owner",
				"approved",
				formatDecisionDetail("Name", owner.ManagerFullName),
				formatDecisionDetail("Email", owner.Email),
			)
		}
	}

	return nil
}
