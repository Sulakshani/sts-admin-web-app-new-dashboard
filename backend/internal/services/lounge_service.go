package services

import (
	"fmt"
	"sts-backend/internal/database"
	"sts-backend/internal/models"
)

func GetAllLounges() ([]models.Lounge, error) {
	repo := database.NewLoungeRepository(database.DB)
	return repo.GetLounges()
}

func GetPendingLounges() ([]models.Lounge, error) {
	repo := database.NewLoungeRepository(database.DB)
	return repo.GetPendingLounges()
}

func GetLoungeByID(id string) (*models.Lounge, error) {
	repo := database.NewLoungeRepository(database.DB)
	return repo.GetLoungeByID(id)
}

func CreateLounge(l models.Lounge) error {
	l.Verification = ensurePendingApprovalStatus(l.Verification)

	repo := database.NewLoungeRepository(database.DB)
	if err := repo.CreateLounge(l); err != nil {
		return err
	}

	if isPendingApprovalStatus(l.Verification) {
		// Lounge creation also creates a lounge owner record, so notify for both.
		notifyApprovalRequest(
			"lounge owner",
			fmt.Sprintf("Owner: %s", l.LoungeOwner),
			fmt.Sprintf("Email: %s", l.OwnerEmail),
			fmt.Sprintf("Contact: %s", l.OwnerContact),
			fmt.Sprintf("NIC: %s", l.OwnerNIC),
		)
		notifyApprovalRequest(
			"lounge",
			fmt.Sprintf("Lounge: %s", l.LoungeName),
			fmt.Sprintf("Owner: %s", l.LoungeOwner),
			fmt.Sprintf("Contact: %s", l.LoungeContact),
			fmt.Sprintf("Address: %s", l.Address),
		)
	}

	return nil
}

func UpdateLounge(l models.Lounge) error {
	repo := database.NewLoungeRepository(database.DB)
	return repo.UpdateLounge(l)
}

func DeleteLounge(id string) error {
	repo := database.NewLoungeRepository(database.DB)
	return repo.DeleteLounge(id)
}

func UpdateLoungeVerification(id string, status string, documents string) error {
	repo := database.NewLoungeRepository(database.DB)
	if err := repo.UpdateLoungeVerification(id, status, documents); err != nil {
		return err
	}

	if isApprovedStatus(status) {
		lounge, err := repo.GetLoungeByID(id)
		if err != nil {
			return err
		}
		if lounge != nil {
			notifyApprovalDecision(
				lounge.OwnerContact,
				"lounge",
				"approved",
				formatDecisionDetail("Lounge", lounge.LoungeName),
				formatDecisionDetail("Owner", lounge.LoungeOwner),
			)
		}
	}

	return nil
}
