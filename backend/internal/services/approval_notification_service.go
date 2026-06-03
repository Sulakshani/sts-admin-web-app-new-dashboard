package services

import (
	"fmt"
	"log"
	"strings"

	"sts-backend/internal/config"
)

const defaultApprovalNotificationPhone = "94715342627"

var defaultApprovalNotificationPhones = []string{"94715342627", "94772945875"}

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

func normalizeApprovalNotificationPhones(phones string) []string {
	phoneList := strings.FieldsFunc(phones, func(r rune) bool {
		return r == ',' || r == ';'
	})

	if len(phoneList) == 0 {
		return append([]string(nil), defaultApprovalNotificationPhones...)
	}

	seen := make(map[string]struct{}, len(phoneList))
	normalizedPhones := make([]string, 0, len(phoneList))
	for _, phone := range phoneList {
		normalizedPhone := normalizeApprovalNotificationPhone(phone)
		if normalizedPhone == "" {
			continue
		}
		if _, exists := seen[normalizedPhone]; exists {
			continue
		}
		seen[normalizedPhone] = struct{}{}
		normalizedPhones = append(normalizedPhones, normalizedPhone)
	}

	if len(normalizedPhones) == 0 {
		return append([]string(nil), defaultApprovalNotificationPhones...)
	}

	return normalizedPhones
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
	recipients := normalizeApprovalNotificationPhones(cfg.ApprovalNotificationPhone)

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
