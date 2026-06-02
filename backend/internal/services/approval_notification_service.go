package services

import (
	"fmt"
	"log"
	"strings"

	"sts-backend/internal/config"
)

const defaultApprovalNotificationPhone = "94715342627"

type approvalSMSNotifier interface {
	SendApprovalRequestNotification(recipient, requestType string, details ...string) error
	SendApprovalDecisionNotification(recipient, requestType, decision string, details ...string) error
}

var loadApprovalNotificationConfig = config.LoadConfig

var newApprovalSMSNotifier = func(cfg *config.Config) approvalSMSNotifier {
	return NewSMSService(cfg)
}

func normalizeApprovalNotificationPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "").Replace(phone)

	if phone == "" {
		return defaultApprovalNotificationPhone
	}

	if strings.HasPrefix(phone, "0") && len(phone) == 10 {
		return "94" + strings.TrimPrefix(phone, "0")
	}

	return phone
}

func ensurePendingApprovalStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "pending"
	}
	return strings.ToLower(strings.TrimSpace(status))
}

func isPendingApprovalStatus(status string) bool {
	return ensurePendingApprovalStatus(status) == "pending"
}

func isApprovedStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	return s == "approved" || s == "verified"
}

func notifyApprovalRequest(requestType string, details ...string) {
	cfg := loadApprovalNotificationConfig()
	phoneList := strings.Split(cfg.ApprovalNotificationPhone, ",")

	var recipients []string
	for _, phone := range phoneList {
		normalizedPhone := normalizeApprovalNotificationPhone(phone)
		if normalizedPhone != "" {
			recipients = append(recipients, normalizedPhone)
		}
	}

	if len(recipients) == 0 {
		log.Println("No valid recipients found for approval notification.")
		return
	}

	smsService := newApprovalSMSNotifier(cfg)
	// Use a new method for bulk sending if available, or loop
	for _, recipient := range recipients {
		if err := smsService.SendApprovalRequestNotification(recipient, requestType, details...); err != nil {
			log.Printf("failed to send %s approval notification to %s: %v", requestType, recipient, err)
		}
	}
}

func notifyApprovalDecision(recipient, requestType, decision string, details ...string) {
	cfg := loadApprovalNotificationConfig()
	if strings.TrimSpace(recipient) == "" {
		recipient = cfg.ApprovalNotificationPhone
	}
	recipient = normalizeApprovalNotificationPhone(recipient)

	smsService := newApprovalSMSNotifier(cfg)
	if err := smsService.SendApprovalDecisionNotification(recipient, requestType, decision, details...); err != nil {
		log.Printf("failed to send %s decision notification: %v", requestType, err)
	}
}

func formatDecisionDetail(label, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s", label, value)
}
