package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sts-backend/internal/config"
	"sync"
	"time"
)

type SMSService struct {
	config *config.Config
}

// SMSRequest represents the request payload for eSMS API
type SMSRequest struct {
	Recipient string `json:"recipient"`
	SenderID  string `json:"senderId"`
	Message   string `json:"message"`
}

// SMSResponse represents the response from eSMS API
type SMSResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Reference  string `json:"reference"`
}

// DialogSMSResponse represents the response from Dialog eSMS API
type DialogSMSResponse struct {
	Status  string      `json:"status"`
	Comment string      `json:"comment"`
	Data    interface{} `json:"data,omitempty"`
	ErrCode string      `json:"errCode"`
}

// DialogLoginResponse represents the login response from Dialog eSMS API
type DialogLoginResponse struct {
	Status        string `json:"status"`
	Comment       string `json:"comment"`
	AccessToken   string `json:"accessToken"`
	TokenValidity string `json:"tokenValidity"`
	ErrCode       string `json:"errCode"`
}

var (
	dialogAccessToken string
	tokenMutex        sync.Mutex
)

func NewSMSService(cfg *config.Config) *SMSService {
	return &SMSService{
		config: cfg,
	}
}

func (s *SMSService) dialogAPIv2BaseURL() string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.config.DialogSMSAPIURL), "/")
	if baseURL == "" {
		baseURL = "https://e-sms.dialog.lk/api/v2"
	}
	return baseURL
}

func normalizeDialogAPIv2Mobile(recipient string) (string, error) {
	mobile := strings.TrimSpace(recipient)
	mobile = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "").Replace(mobile)

	switch {
	case strings.HasPrefix(mobile, "94") && len(mobile) == 11:
		mobile = strings.TrimPrefix(mobile, "94")
	case strings.HasPrefix(mobile, "0") && len(mobile) == 10:
		mobile = strings.TrimPrefix(mobile, "0")
	}

	if len(mobile) != 9 {
		return "", fmt.Errorf("Dialog API v2 mobile number must be 9 digits after normalization, got %q", mobile)
	}

	return mobile, nil
}

// SendSMS sends an SMS using the configured method
func (s *SMSService) SendSMS(recipient, message string) error {
	hasDialogURLConfig := s.config.DialogSMSMethod == "url" && s.config.DialogSMSEsmsqk != ""
	hasDialogAPIv2Token := strings.TrimSpace(s.config.DialogSMSAccessToken) != ""
	hasDialogAPIv2Login := s.config.DialogSMSUsername != "" && s.config.DialogSMSPassword != ""
	hasDialogAPIv2Config := s.config.DialogSMSMethod == "api_v2" && (hasDialogAPIv2Token || hasDialogAPIv2Login)
	hasLegacyConfig := s.config.ESMSAPIKey != ""

	// Keep dev mode non-destructive only when no real gateway is configured.
	if s.config.SMSMode == "dev" && !hasDialogURLConfig && !hasDialogAPIv2Config && !hasLegacyConfig {
		log.Printf("📱 [DEV MODE] SMS not sent. Recipient: %s, Message: %s", recipient, message)
		return nil
	}

	// Use the configured Dialog method first, then legacy eSMS only if Dialog is not configured.
	if hasDialogURLConfig {
		return s.sendDialogSMSURL(recipient, message)
	}

	if hasDialogAPIv2Config {
		return s.sendDialogSMSAPIv2(recipient, message)
	}

	// Fallback to legacy eSMS API if Dialog is not configured.
	if hasLegacyConfig {
		return s.sendLegacyESMS(recipient, message)
	}

	log.Println("⚠️ SMS service not configured - skipping SMS notification")
	return nil
}

// sendDialogSMSURL sends SMS using Dialog eSMS URL method (GET request with esmsqk)
func (s *SMSService) sendDialogSMSURL(recipient, message string) error {
	if s.config.DialogSMSEsmsqk == "" {
		return fmt.Errorf("Dialog SMS esmsqk key not configured")
	}

	// Build URL with query parameters for the Dialog URL Message Key flow.
	baseURL := "https://e-sms.dialog.lk/api/v1/message-via-url/create/url-campaign"
	pushNotificationURL := s.config.FrontendURL
	if strings.TrimSpace(pushNotificationURL) == "" {
		pushNotificationURL = "https://xx/xx"
	}
	params := url.Values{}
	params.Add("esmsqk", s.config.DialogSMSEsmsqk)
	params.Add("list", recipient)
	params.Add("source_address", s.config.DialogSMSMask)
	params.Add("message", message)
	params.Add("push_notification_url", pushNotificationURL)

	fullURL := baseURL + "?" + params.Encode()

	// Send GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to send Dialog SMS: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Dialog SMS response: %w", err)
	}

	// Parse response
	var dialogResponse DialogSMSResponse
	if err := json.Unmarshal(body, &dialogResponse); err != nil {
		log.Printf("Failed to parse Dialog SMS response: %v. Raw response: %s", err, string(body))
		plainResponse := strings.TrimSpace(string(body))
		if plainResponse == "1" {
			log.Printf("✅ Dialog SMS sent successfully to %s (Status: %d)", recipient, resp.StatusCode)
			return nil
		}
		return fmt.Errorf("Dialog SMS API returned code %s", plainResponse)
	}

	// Check response status
	if dialogResponse.Status != "success" && resp.StatusCode != 200 {
		return fmt.Errorf("Dialog SMS API error: %s", dialogResponse.Comment)
	}

	log.Printf("✅ Dialog SMS sent successfully to %s", recipient)
	return nil
}

// sendDialogSMSAPIv2 sends SMS using Dialog eSMS API v2 method (POST with bearer token)
func (s *SMSService) sendDialogSMSAPIv2(recipient, message string) error {
	if strings.TrimSpace(s.config.DialogSMSAccessToken) == "" && (s.config.DialogSMSUsername == "" || s.config.DialogSMSPassword == "") {
		return fmt.Errorf("Dialog SMS API v2 token or login credentials not configured")
	}

	token, err := s.getDialogToken()
	if err != nil {
		return fmt.Errorf("failed to get Dialog API token: %w", err)
	}
	mobile, err := normalizeDialogAPIv2Mobile(recipient)
	if err != nil {
		return err
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"msisdn":         []map[string]string{{"mobile": mobile}},
		"message":        message,
		"sourceAddress":  s.config.DialogSMSMask,
		"transaction_id": time.Now().UnixMilli(),
		"payment_method": 0,
	}
	if strings.TrimSpace(s.config.FrontendURL) != "" {
		payload["push_notification_url"] = s.config.FrontendURL
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Dialog SMS request: %w", err)
	}

	log.Printf("Dialog APIv2 Request Payload: %s", string(jsonData))

	// Create HTTP request
	apiURL := s.dialogAPIv2BaseURL() + "/sms"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Dialog SMS request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Dialog SMS request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Dialog SMS response: %w", err)
	}

	// Parse response
	var dialogResponse DialogSMSResponse
	if err := json.Unmarshal(body, &dialogResponse); err != nil {
		log.Printf("Failed to parse Dialog SMS response: %v. Raw response: %s", err, string(body))
		return fmt.Errorf("Dialog SMS API returned an unparsable response: %s", string(body))
	}

	// Check response status
	if dialogResponse.Status != "success" {
		return fmt.Errorf("Dialog SMS API error: %s (Code: %s)", dialogResponse.Comment, dialogResponse.ErrCode)
	}

	log.Printf("✅ Dialog SMS sent successfully to %s", recipient)
	return nil
}

func (s *SMSService) getDialogToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if token := strings.TrimSpace(s.config.DialogSMSAccessToken); token != "" {
		return token, nil
	}

	// Fallback for accounts where Dialog enables an API login endpoint.

	loginPayload := map[string]string{
		"username": s.config.DialogSMSUsername,
		"password": s.config.DialogSMSPassword,
	}
	jsonData, err := json.Marshal(loginPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	log.Printf("Dialog Login Request Payload: %s", string(jsonData))

	loginURL := s.dialogAPIv2BaseURL() + "/login"
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}
	log.Printf("Dialog Login Response: %s", string(body))

	var loginResponse DialogLoginResponse
	if err := json.Unmarshal(body, &loginResponse); err != nil {
		return "", fmt.Errorf("failed to parse login response: %s", string(body))
	}

	if loginResponse.Status != "success" {
		return "", fmt.Errorf("Dialog API login failed: %s (Code: %s)", loginResponse.Comment, loginResponse.ErrCode)
	}

	dialogAccessToken = loginResponse.AccessToken
	return dialogAccessToken, nil
}

// sendLegacyESMS sends SMS using legacy eSMS API (for backward compatibility)
func (s *SMSService) sendLegacyESMS(recipient, message string) error {
	// Prepare request payload
	smsRequest := SMSRequest{
		Recipient: recipient,
		SenderID:  s.config.ESMSSenderID,
		Message:   message,
	}

	jsonData, err := json.Marshal(smsRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.config.ESMSAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.ESMSAPIKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SMS response: %w", err)
	}

	// Parse response
	var smsResponse SMSResponse
	if err := json.Unmarshal(body, &smsResponse); err != nil {
		log.Printf("Failed to parse SMS response: %v. Raw response: %s", err, string(body))
		// If status is 200, consider it success even if parsing fails
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			log.Printf("✅ SMS sent successfully to %s (Status: %d)", recipient, resp.StatusCode)
			return nil
		}
		return fmt.Errorf("SMS API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check response status
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("SMS API error: %s (Status: %d)", smsResponse.Message, resp.StatusCode)
	}

	log.Printf("✅ SMS sent successfully to %s (Reference: %s)", recipient, smsResponse.Reference)
	return nil
}

// SendNewComplaintNotification sends SMS notification for new complaint assignment
func (s *SMSService) SendNewComplaintNotification(recipient, assigneeName, complaintID, category string) error {
	message := fmt.Sprintf(
		"Hello %s,\n\nYou have been assigned a new complaint:\n\nComplaint ID: %s\nCategory: %s\n\nPlease review and respond within 5 days.\n\nThank you!",
		assigneeName,
		complaintID,
		category,
	)

	log.Printf("📱 Sending new complaint notification to %s", recipient)
	return s.SendSMS(recipient, message)
}

// SendEscalationNotification sends SMS notification for complaint escalation
func (s *SMSService) SendEscalationNotification(recipient, assigneeName, complaintID, category string, level int) error {
	message := fmt.Sprintf(
		"Hello %s,\n\nA complaint has been escalated to you (Level %d):\n\nComplaint ID: %s\nCategory: %s\n\nThis requires urgent attention. Please review immediately.\n\nThank you!",
		assigneeName,
		level,
		complaintID,
		category,
	)

	log.Printf("📱 Sending escalation notification (Level %d) to %s", level, recipient)
	return s.SendSMS(recipient, message)
}

// SendResolutionNotification sends SMS notification when complaint is resolved
func (s *SMSService) SendResolutionNotification(recipient, customerName, complaintID string) error {
	message := fmt.Sprintf(
		"Dear %s,\n\nYour complaint (ID: %s) has been resolved.\n\nThank you for your patience!\n\nBusLounge Team",
		customerName,
		complaintID,
	)

	log.Printf("📱 Sending resolution notification to %s", recipient)
	return s.SendSMS(recipient, message)
}

// SendApprovalRequestNotification sends a generic approval request SMS.
func (s *SMSService) SendApprovalRequestNotification(recipient, requestType string, details ...string) error {
	var builder strings.Builder
	builder.WriteString("Hello,\n\n")
	builder.WriteString(fmt.Sprintf("A new %s approval request has been submitted.\n\n", requestType))
	for _, detail := range details {
		if detail == "" {
			continue
		}
		builder.WriteString(detail)
		if !strings.HasSuffix(detail, "\n") {
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\nPlease review it in the admin panel.\n\nThank you!")

	log.Printf("📱 Sending %s approval notification to %s", requestType, recipient)
	return s.SendSMS(recipient, builder.String())
}

// SendApprovalDecisionNotification sends an approval or rejection SMS to the requester.
func (s *SMSService) SendApprovalDecisionNotification(recipient, requestType, decision string, details ...string) error {
	decision = strings.ToLower(strings.TrimSpace(decision))
	if decision == "" {
		decision = "updated"
	}

	var builder strings.Builder
	builder.WriteString("Hello,\n\n")
	builder.WriteString(fmt.Sprintf("Your %s request has been %s.\n\n", requestType, decision))
	for _, detail := range details {
		if detail == "" {
			continue
		}
		builder.WriteString(detail)
		if !strings.HasSuffix(detail, "\n") {
			builder.WriteString("\n")
		}
	}
	if decision == "approved" || decision == "verified" {
		builder.WriteString("\nYou can now log in using your registered account.\n")
	}
	builder.WriteString("\nThank you!\n\n- STS Team")

	log.Printf("📱 Sending %s decision notification to %s", requestType, recipient)
	return s.SendSMS(recipient, builder.String())
}

// SendBulkSMS sends SMS to multiple recipients (for team notifications)
func (s *SMSService) SendBulkSMS(recipients []string, message string) error {
	if s.config.ESMSAPIKey == "" {
		log.Println("⚠️ SMS service not configured - skipping bulk SMS notification")
		return nil
	}

	var lastError error
	successCount := 0

	for _, recipient := range recipients {
		err := s.SendSMS(recipient, message)
		if err != nil {
			log.Printf("❌ Failed to send SMS to %s: %v", recipient, err)
			lastError = err
		} else {
			successCount++
		}
	}

	if successCount > 0 {
		log.Printf("✅ Successfully sent %d/%d SMS notifications", successCount, len(recipients))
	}

	return lastError
}

// SendAdminApprovalRequestNotification sends an approval request SMS to a list of admin phones.
func (s *SMSService) SendAdminApprovalRequestNotification(adminPhones []string, requestType string, details ...string) {
	var builder strings.Builder
	builder.WriteString("Hello Admin,\n\n")
	builder.WriteString(fmt.Sprintf("A new %s approval request has been submitted.\n\n", requestType))
	for _, detail := range details {
		if detail == "" {
			continue
		}
		builder.WriteString(detail)
		if !strings.HasSuffix(detail, "\n") {
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\nPlease review it in the admin panel.\n\nThank you!")

	message := builder.String()
	log.Printf("📱 Sending %s approval notification to %d admins", requestType, len(adminPhones))

	// Use a wait group to send SMS messages concurrently
	var wg sync.WaitGroup
	for _, phone := range adminPhones {
		if strings.TrimSpace(phone) == "" {
			continue
		}
		wg.Add(1)
		go func(recipient string) {
			defer wg.Done()
			err := s.SendSMS(recipient, message)
			if err != nil {
				log.Printf("❌ Failed to send admin approval SMS to %s: %v", recipient, err)
			}
		}(phone)
	}
	wg.Wait()
}
