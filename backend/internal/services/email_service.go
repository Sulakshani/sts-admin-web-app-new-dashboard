package services

import (
	"bytes"
	"fmt"
	"net/smtp"
	"sts-backend/internal/config"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// SendPasswordResetEmail sends a password reset email with a token
func (s *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	subject := "Password Reset Request - STS Admin"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.FrontendURL, resetToken)
	
	// HTML email template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .code-box { background: white; border: 2px dashed #667eea; padding: 20px; margin: 20px 0; text-align: center; border-radius: 5px; }
        .code { font-size: 24px; font-weight: bold; color: #667eea; letter-spacing: 3px; }
        .footer { color: #999; font-size: 12px; text-align: center; margin-top: 20px; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>We received a request to reset your password for your STS Admin Dashboard account.</p>
            
            <div class="code-box">
                <p style="margin: 0; color: #666;">Your Reset Code:</p>
                <div class="code">%s</div>
                <p style="margin: 10px 0 0 0; font-size: 12px; color: #999;">Expires in 15 minutes</p>
            </div>

            <p style="text-align: center;">Or click the button below to reset your password:</p>
            
            <div style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </div>

            <div class="warning">
                <strong>⚠️ Security Notice:</strong>
                <ul style="margin: 10px 0 0 0;">
                    <li>This link expires in 15 minutes</li>
                    <li>If you didn't request this reset, please ignore this email</li>
                    <li>Never share this code with anyone</li>
                </ul>
            </div>

            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;"><a href="%s">%s</a></p>
        </div>
        <div class="footer">
            <p>© 2026 Bus & Lounge Reservation System. All rights reserved.</p>
            <p>This is an automated email. Please do not reply to this message.</p>
        </div>
    </div>
</body>
</html>
`, resetToken, resetURL, resetURL, resetURL)

	// Plain text version (fallback)
	plainBody := fmt.Sprintf(`
Password Reset Request

Hello,

We received a request to reset your password for your STS Admin Dashboard account.

Your Reset Code: %s
This code expires in 15 minutes.

Or visit this link to reset your password:
%s

Security Notice:
- This link expires in 15 minutes
- If you didn't request this reset, please ignore this email
- Never share this code with anyone

© 2026 Bus & Lounge Reservation System. All rights reserved.
This is an automated email. Please do not reply to this message.
`, resetToken, resetURL)

	return s.sendEmail(toEmail, subject, plainBody, htmlBody)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, plainBody, htmlBody string) error {
	// Check if SMTP is configured
	if s.config.SMTPUsername == "" || s.config.SMTPPassword == "" {
		return fmt.Errorf("SMTP not configured. Please set SMTP_USERNAME and SMTP_PASSWORD in .env file")
	}

	from := s.config.SMTPFromEmail
	
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.SMTPFromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/alternative; boundary=boundary123"

	// Build message
	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	
	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(plainBody)
	message.WriteString("\r\n\r\n")
	
	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")
	message.WriteString("--boundary123--")

	// SMTP authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email (optional - can be used later)
func (s *EmailService) SendWelcomeEmail(toEmail, adminName string) error {
	subject := "Welcome to STS Admin Dashboard"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif;">
    <h2>Welcome, %s!</h2>
    <p>Your admin account has been created successfully.</p>
    <p>You can now log in to the STS Admin Dashboard.</p>
</body>
</html>
`, adminName)
	
	plainBody := fmt.Sprintf("Welcome, %s! Your admin account has been created successfully.", adminName)
	
	return s.sendEmail(toEmail, subject, plainBody, htmlBody)
}
