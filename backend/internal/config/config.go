package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	MaxConnections     int
	MaxIdleConnections int
	ConnMaxLifetime    int
	Port               string

	// Email configuration
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string
	FrontendURL   string

	// SMS configuration
	SMSMode                   string // "dev" or "production"
	DialogSMSMethod           string // "url" or "api_v2"
	ApprovalNotificationPhone string

	// Dialog eSMS URL Method (Recommended)
	DialogSMSEsmsqk string
	DialogSMSMask   string

	// Dialog eSMS API v2 Method (Alternative)
	DialogSMSAPIURL      string
	DialogSMSAccessToken string
	DialogSMSUsername    string
	DialogSMSPassword    string

	// Legacy eSMS API support
	ESMSAPIURL   string
	ESMSAPIKey   string
	ESMSSenderID string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	return &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		MaxConnections:     getEnvAsInt("DATABASE_MAX_CONNECTIONS", 10),
		MaxIdleConnections: getEnvAsInt("DATABASE_MAX_IDLE_CONNECTIONS", 5),
		ConnMaxLifetime:    getEnvAsInt("DATABASE_CONN_MAX_LIFETIME", 300),
		Port:               getEnv("PORT", "8080"),

		// Email configuration
		SMTPHost:      getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:      getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail: getEnv("SMTP_FROM_EMAIL", "noreply@sts.lk"),
		SMTPFromName:  getEnv("SMTP_FROM_NAME", "STS Admin System"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:4200"),

		// SMS configuration
		SMSMode:                   getEnv("SMS_MODE", "dev"),
		DialogSMSMethod:           getEnv("DIALOG_SMS_METHOD", "url"),
		ApprovalNotificationPhone: getEnv("APPROVAL_NOTIFICATION_PHONE", "94715342627,94772945875"),

		// Dialog eSMS URL Method
		DialogSMSEsmsqk: getEnv("DIALOG_SMS_ESMSQK", ""),
		DialogSMSMask:   getEnv("DIALOG_SMS_MASK", "KanchTest"),

		// Dialog eSMS API v2 Method
		DialogSMSAPIURL:      getEnv("DIALOG_SMS_API_URL", "https://e-sms.dialog.lk/api/v2"),
		DialogSMSAccessToken: getEnv("DIALOG_SMS_ACCESS_TOKEN", ""),
		DialogSMSUsername:    getEnv("DIALOG_SMS_USERNAME", ""),
		DialogSMSPassword:    getEnv("DIALOG_SMS_PASSWORD", ""),

		// Legacy eSMS API support
		ESMSAPIURL:   getEnv("ESMS_API_URL", "https://api.esms.lk/v1/sms/send"),
		ESMSAPIKey:   getEnv("ESMS_API_KEY", ""),
		ESMSSenderID: getEnv("ESMS_SENDER_ID", "BusLounge"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}
	return value
}
