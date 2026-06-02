package handlers

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
	"sts-backend/internal/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Global email service
var emailService *services.EmailService
var appConfig *config.Config

type bootstrapStatusInfo struct {
	State      string    `json:"state"`
	Message    string    `json:"message"`
	CheckedAt  time.Time `json:"checked_at"`
	Configured bool      `json:"configured"`
}

var bootstrapStatus = bootstrapStatusInfo{
	State:     "not_checked",
	Message:   "Bootstrap check not executed yet",
	CheckedAt: time.Time{},
}

// InitAuthHandlers initializes the auth handlers with required services
func InitAuthHandlers(cfg *config.Config) {
	appConfig = cfg
	emailService = services.NewEmailService(cfg)
	if envSecret := strings.TrimSpace(os.Getenv("JWT_SECRET")); envSecret != "" {
		jwtSecret = []byte(envSecret)
	}
	ensureBootstrapSuperAdmin()
}

func setBootstrapStatus(state, message string, configured bool) {
	bootstrapStatus = bootstrapStatusInfo{
		State:      state,
		Message:    message,
		CheckedAt:  time.Now(),
		Configured: configured,
	}
}

func ensureBootstrapSuperAdmin() {
	if database.DB == nil {
		setBootstrapStatus("db_unavailable", "Database connection is not available during bootstrap check", false)
		log.Printf("Super admin bootstrap check: database unavailable")
		return
	}

	email := strings.TrimSpace(os.Getenv("SUPERADMIN_EMAIL"))
	password := os.Getenv("SUPERADMIN_PASSWORD")
	fullName := strings.TrimSpace(os.Getenv("SUPERADMIN_NAME"))
	if fullName == "" {
		fullName = "Super Administrator"
	}

	if email == "" || password == "" {
		setBootstrapStatus("skipped_missing_env", "SUPERADMIN_EMAIL and/or SUPERADMIN_PASSWORD not set; bootstrap skipped", false)
		log.Printf("Super admin bootstrap check: skipped because SUPERADMIN_EMAIL or SUPERADMIN_PASSWORD is not set")
		return
	}

	configured := true

	var exists bool
	if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM admin_users WHERE email = $1)`, email).Scan(&exists); err != nil {
		setBootstrapStatus("error", "Failed to query admin_users during bootstrap check", configured)
		log.Printf("Super admin bootstrap check failed: %v", err)
		return
	}
	if exists {
		setBootstrapStatus("already_exists", "Super admin already exists in admin_users", configured)
		log.Printf("Super admin bootstrap check: account already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		setBootstrapStatus("error", "Failed to hash super admin password during bootstrap", configured)
		log.Printf("Super admin bootstrap failed while hashing password: %v", err)
		return
	}

	permissionsJSON, _ := json.Marshal([]string{"*"})
	if _, err := database.DB.Exec(`
		INSERT INTO admin_users (
			id, email, password_hash, full_name, role, app_scope,
			supervisor_id, function_permissions, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, 'super_admin', NULL, NULL, $5::jsonb, true, NOW(), NOW())
	`, uuid.New().String(), email, string(hash), fullName, string(permissionsJSON)); err != nil {
		setBootstrapStatus("error", "Failed to insert bootstrap super admin into admin_users", configured)
		log.Printf("Super admin bootstrap insert failed: %v", err)
		return
	}

	setBootstrapStatus("created", "Bootstrap super admin created successfully", configured)
	log.Printf("Bootstrapped super admin account for %s", email)
}

// GetBootstrapStatus returns the latest super admin bootstrap check result.
func GetBootstrapStatus(c *gin.Context) {
	c.JSON(http.StatusOK, bootstrapStatus)
}

type Admin struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never send password in response
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"` // super_admin, supervisor, admin
	AppScope  string    `json:"app_scope,omitempty"`
	SupervisorID *string `json:"supervisor_id,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	AdminUser    AdminUser `json:"admin_user"`
}

type AdminUser struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	Role        string    `json:"role"`
	AppScope    string    `json:"app_scope,omitempty"`
	SupervisorID *string   `json:"supervisor_id,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	IsActive    bool      `json:"is_active"`
	LastLoginAt time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JWT Secret - In production, use environment variable
var jwtSecret = []byte("your-secret-key-change-in-production")

// Password reset token storage (in-memory - in production use Redis or database)
type PasswordResetToken struct {
	Email     string
	Token     string
	ExpiresAt time.Time
	Used      bool
}

var passwordResetTokens = make(map[string]*PasswordResetToken)

// Helper function to get admin from database
func getAdminByEmail(email string) (Admin, bool) {
	if database.DB == nil {
		return Admin{}, false
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	admin, err := getAdminByEmailFromDB(normalizedEmail)
	if err == nil {
		return admin, true
	}

	return Admin{}, false
}

func getAdminByEmailFromDB(email string) (Admin, error) {
	query := `
		SELECT id::text, email, password_hash, full_name,
			COALESCE(role, 'admin') AS role,
			COALESCE(app_scope, '') AS app_scope,
			supervisor_id::text,
			COALESCE(function_permissions, '[]'::jsonb) AS function_permissions,
			COALESCE(is_active, true),
			COALESCE(created_at, NOW()),
			COALESCE(updated_at, NOW())
		FROM admin_users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`

	var admin Admin
	var supervisorID sql.NullString
	var functionPermissions []byte
	err := database.DB.QueryRow(query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Password,
		&admin.FullName,
		&admin.Role,
		&admin.AppScope,
		&supervisorID,
		&functionPermissions,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return Admin{}, err
	}
	if supervisorID.Valid {
		admin.SupervisorID = &supervisorID.String
	}
	if err := json.Unmarshal(functionPermissions, &admin.Permissions); err != nil {
		admin.Permissions = []string{}
	}

	return admin, nil
}

func getAdminByIDFromDB(id string) (Admin, error) {
	query := `
		SELECT id::text, email, password_hash, full_name,
			COALESCE(role, 'admin') AS role,
			COALESCE(app_scope, '') AS app_scope,
			supervisor_id::text,
			COALESCE(function_permissions, '[]'::jsonb) AS function_permissions,
			COALESCE(is_active, true),
			COALESCE(created_at, NOW()),
			COALESCE(updated_at, NOW())
		FROM admin_users
		WHERE id::text = $1
		LIMIT 1
	`

	var admin Admin
	var supervisorID sql.NullString
	var functionPermissions []byte
	err := database.DB.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.Email,
		&admin.Password,
		&admin.FullName,
		&admin.Role,
		&admin.AppScope,
		&supervisorID,
		&functionPermissions,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return Admin{}, err
	}
	if supervisorID.Valid {
		admin.SupervisorID = &supervisorID.String
	}
	if err := json.Unmarshal(functionPermissions, &admin.Permissions); err != nil {
		admin.Permissions = []string{}
	}

	return admin, nil
}

func verifyAdminPassword(admin *Admin, password string) bool {
	stored := strings.TrimSpace(admin.Password)
	if stored == "" {
		return false
	}

	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}

	// Backward compatibility for legacy rows that may still store plain text.
	if subtle.ConstantTimeCompare([]byte(stored), []byte(password)) != 1 {
		return false
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return true
	}

	admin.Password = string(newHash)
	if database.DB != nil {
		_, _ = database.DB.Exec(`UPDATE admin_users SET password_hash = $2, updated_at = NOW() WHERE id::text = $1`, admin.ID, admin.Password)
	}

	return true
}

// AdminLogin handles admin login using admin_users table credentials
func AdminLogin(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find admin by email
	admin, exists := getAdminByEmail(req.Email)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Check if admin is active
	if !admin.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Account is inactive",
		})
		return
	}

	// Verify password
	if !verifyAdminPassword(&admin, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	if database.DB != nil {
		_, _ = database.DB.Exec(`UPDATE admin_users SET last_login_at = NOW(), updated_at = NOW() WHERE id::text = $1`, admin.ID)
		if latestAdmin, ok := getAdminByEmail(admin.Email); ok {
			admin = latestAdmin
		}
	}

	// Generate tokens
	accessToken, err := generateAccessToken(admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	refreshToken, err := generateRefreshToken(admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}

	// Prepare response
	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		AdminUser: AdminUser{
			ID:          admin.ID,
			Email:       admin.Email,
			FullName:    admin.FullName,
			Role:        admin.Role,
			AppScope:    admin.AppScope,
			SupervisorID: admin.SupervisorID,
			Permissions: admin.Permissions,
			IsActive:    admin.IsActive,
			LastLoginAt: time.Now(),
			CreatedAt:   admin.CreatedAt,
			UpdatedAt:   time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// generateAccessToken creates a JWT access token
func generateAccessToken(admin Admin) (string, error) {
	var supervisorID interface{}
	if admin.SupervisorID != nil && *admin.SupervisorID != "" {
		supervisorID = *admin.SupervisorID
	}

	claims := jwt.MapClaims{
		"admin_id":  admin.ID,
		"email":     admin.Email,
		"full_name": admin.FullName,
		"role":      admin.Role,
		"app_scope": admin.AppScope,
		"supervisor_id": supervisorID,
		"permissions": admin.Permissions,
		"exp":       time.Now().Add(time.Hour * 1).Unix(), // 1 hour expiration
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// generateRefreshToken creates a JWT refresh token
func generateRefreshToken(admin Admin) (string, error) {
	var supervisorID interface{}
	if admin.SupervisorID != nil && *admin.SupervisorID != "" {
		supervisorID = *admin.SupervisorID
	}

	claims := jwt.MapClaims{
		"admin_id":      admin.ID,
		"email":         admin.Email,
		"role":          admin.Role,
		"app_scope":     admin.AppScope,
		"supervisor_id": supervisorID,
		"exp":           time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days expiration
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AdminLogout handles admin logout
func AdminLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// AdminProfile returns the current admin's profile
func AdminProfile(c *gin.Context) {
	claims, err := getJWTClaimsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	email, _ := claims["email"].(string)
	admin, exists := getAdminByEmail(email)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Admin not found",
		})
		return
	}

	c.JSON(http.StatusOK, AdminUser{
		ID:          admin.ID,
		Email:       admin.Email,
		FullName:    admin.FullName,
		Role:        admin.Role,
		AppScope:    admin.AppScope,
		SupervisorID: admin.SupervisorID,
		Permissions: admin.Permissions,
		IsActive:    admin.IsActive,
		LastLoginAt: time.Now(),
		CreatedAt:   admin.CreatedAt,
		UpdatedAt:   admin.UpdatedAt,
	})
}

// RefreshAccessToken refreshes the access token using refresh token
func RefreshAccessToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Verify refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token claims",
		})
		return
	}

	email, _ := claims["email"].(string)
	admin, exists := getAdminByEmail(email)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Admin not found",
		})
		return
	}

	// Generate new access token
	accessToken, err := generateAccessToken(admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		ExpiresIn:    3600,
		AdminUser: AdminUser{
			ID:          admin.ID,
			Email:       admin.Email,
			FullName:    admin.FullName,
			Role:        admin.Role,
			AppScope:    admin.AppScope,
			SupervisorID: admin.SupervisorID,
			Permissions: admin.Permissions,
			IsActive:    admin.IsActive,
			LastLoginAt: time.Now(),
			CreatedAt:   admin.CreatedAt,
			UpdatedAt:   time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// RequestPasswordReset initiates password reset process
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// Check if admin exists
	admin, exists := getAdminByEmail(req.Email)
	if !exists {
		// Don't reveal if email exists or not (security best practice)
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Check if admin is active
	if !admin.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Generate reset token (6-digit code for simplicity)
	resetToken := generateResetToken()

	// Store reset token (expires in 15 minutes)
	passwordResetTokens[resetToken] = &PasswordResetToken{
		Email:     req.Email,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// Send password reset email
	if emailService != nil {
		err := emailService.SendPasswordResetEmail(req.Email, resetToken)
		if err != nil {
			// Log error but don't reveal to user
			println("Error sending email:", err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset code has been sent to your email. Please check your inbox.",
	})
}

// VerifyResetToken verifies if a reset token is valid
func VerifyResetToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	resetToken, exists := passwordResetTokens[req.Token]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired reset token",
		})
		return
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		delete(passwordResetTokens, req.Token)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reset token has expired",
		})
		return
	}

	// Check if token was already used
	if resetToken.Used {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reset token has already been used",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"email": resetToken.Email,
	})
}

// ResetPassword resets the admin password using a valid token
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Password must be at least 8 characters",
		})
		return
	}

	resetToken, exists := passwordResetTokens[req.Token]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired reset token",
		})
		return
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		delete(passwordResetTokens, req.Token)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reset token has expired",
		})
		return
	}

	// Check if token was already used
	if resetToken.Used {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Reset token has already been used",
		})
		return
	}

	// Ensure admin exists
	if _, exists := getAdminByEmail(resetToken.Email); !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Admin not found",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process new password",
		})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database connection is required",
		})
		return
	}

	if _, err := database.DB.Exec(`
		UPDATE admin_users
		SET password_hash = $2,
			updated_at = NOW()
		WHERE email = $1
	`, resetToken.Email, string(hashedPassword)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update password",
		})
		return
	}

	// Mark token as used
	resetToken.Used = true

	// Clean up used token after a delay
	go func() {
		time.Sleep(5 * time.Minute)
		delete(passwordResetTokens, req.Token)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// generateResetToken generates a 6-digit reset code
func generateResetToken() string {
	// Generate random 6-digit code
	return uuid.New().String()[:8] // Using first 8 chars of UUID for simplicity
}

// ==================== USER MANAGEMENT ENDPOINTS ====================

// Middleware to check if user is super admin
func RequireSuperAdmin(c *gin.Context) {
	claims, err := getJWTClaimsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	role, _ := claims["role"].(string)
	if normalizeRoleValue(role) != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Super admin access required",
		})
		c.Abort()
		return
	}

	c.Set("auth_claims", claims)
	c.Next()
}

func extractPermissionsFromClaims(claims jwt.MapClaims) []string {
	permissionsRaw, exists := claims["permissions"]
	if !exists || permissionsRaw == nil {
		return []string{}
	}

	if arr, ok := permissionsRaw.([]string); ok {
		return arr
	}

	arr, ok := permissionsRaw.([]interface{})
	if !ok {
		return []string{}
	}

	permissions := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			permissions = append(permissions, s)
		}
	}

	return permissions
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == "*" || p == required {
			return true
		}
		if strings.HasSuffix(p, ".*") {
			prefix := strings.TrimSuffix(p, ".*")
			if strings.HasPrefix(required, prefix+".") {
				return true
			}
		}
	}
	return false
}

func normalizeRoleValue(role string) string {
	r := strings.ToLower(strings.TrimSpace(role))
	r = strings.ReplaceAll(r, "-", "_")
	r = strings.ReplaceAll(r, " ", "_")
	return r
}

func ValidatePermissions(permissions []string) error {
	for _, p := range permissions {
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("permissions must not contain empty values")
		}
	}
	return nil
}

func RequireAdminPermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getJWTClaimsFromRequest(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		role, _ := claims["role"].(string)
		if normalizeRoleValue(role) == "super_admin" {
			c.Next()
			return
		}

		permissions := extractPermissionsFromClaims(claims)
		// Backward compatibility: allow legacy accounts that have no permission claims yet.
		if len(permissions) == 0 {
			c.Next()
			return
		}
		if !hasPermission(permissions, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: " + permission})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getJWTClaimsFromRequest(c *gin.Context) (jwt.MapClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("authorization bearer token required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func isValidAppScope(appScope string) bool {
	return appScope == "bus" || appScope == "driver" || appScope == "lounges" || appScope == "passenger"
}

func validateUserHierarchy(role, appScope string, supervisorID *string, excludeUserID string) error {
	if role != "super_admin" && role != "supervisor" && role != "admin" {
		return fmt.Errorf("role must be one of: super_admin, supervisor, admin")
	}

	if role != "super_admin" {
		if !isValidAppScope(appScope) {
			return fmt.Errorf("app_scope must be one of: bus, driver, lounges, passenger")
		}
		if role == "admin" && (supervisorID == nil || *supervisorID == "") {
			return fmt.Errorf("supervisor_id is required for admin users")
		}
	}

	if supervisorID == nil || *supervisorID == "" {
		return nil
	}

	if excludeUserID != "" && *supervisorID == excludeUserID {
		return fmt.Errorf("supervisor_id cannot be the same as user id")
	}

	if database.DB == nil {
		return nil
	}

	var supervisorRole, supervisorScope string
	err := database.DB.QueryRow(`
		SELECT role, COALESCE(app_scope, '')
		FROM admin_users
		WHERE id::text = $1 AND is_active = true
	`, *supervisorID).Scan(&supervisorRole, &supervisorScope)
	if err != nil {
		return fmt.Errorf("invalid supervisor_id")
	}

	if supervisorRole != "supervisor" && supervisorRole != "super_admin" {
		return fmt.Errorf("supervisor_id must belong to supervisor or super_admin")
	}

	if supervisorRole == "supervisor" && supervisorScope != appScope {
		return fmt.Errorf("supervisor and user must have the same app_scope")
	}

	return nil
}

// CreateUser creates a new admin user (super admin only)
func CreateUser(c *gin.Context) {
	var req struct {
		Email        string  `json:"email" binding:"required,email"`
		FullName     string  `json:"full_name" binding:"required"`
		Password     string  `json:"password" binding:"required,min=8"`
		Role         string  `json:"role" binding:"required"`
		AppScope     string  `json:"app_scope"`
		SupervisorID *string `json:"supervisor_id"`
		Permissions  []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Password must be at least 8 characters",
		})
		return
	}

	// Check if user already exists
	if _, exists := getAdminByEmail(req.Email); exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User with this email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	if req.Role != "super_admin" && req.Role != "supervisor" && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be one of: super_admin, supervisor, admin"})
		return
	}

	if err := validateUserHierarchy(req.Role, req.AppScope, req.SupervisorID, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != "super_admin" && len(req.Permissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permissions are required for non-super-admin users"})
		return
	}
	if err := ValidatePermissions(req.Permissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID := ""
	claims, err := getJWTClaimsFromRequest(c)
	if err == nil {
		creatorID, _ = claims["admin_id"].(string)
	}

	// Create new admin user
	newAdmin := Admin{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FullName:  req.FullName,
		Role:      req.Role,
		AppScope:  req.AppScope,
		SupervisorID: req.SupervisorID,
		Permissions: req.Permissions,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if newAdmin.Role == "super_admin" {
		newAdmin.Permissions = []string{"*"}
	}

	permissionsJSON, _ := json.Marshal(newAdmin.Permissions)

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection is required"})
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO admin_users (
			id, email, password_hash, full_name, role, app_scope,
			supervisor_id, function_permissions, is_active, created_at, updated_at, created_by
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7::uuid, $8::jsonb, true, NOW(), NOW(), NULLIF($9, '')::uuid)
	`, newAdmin.ID, newAdmin.Email, newAdmin.Password, newAdmin.FullName, newAdmin.Role, newAdmin.AppScope, newAdmin.SupervisorID, string(permissionsJSON), creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin user"})
		return
	}

	c.JSON(http.StatusCreated, AdminUser{
		ID:        newAdmin.ID,
		Email:     newAdmin.Email,
		FullName:  newAdmin.FullName,
		Role:      newAdmin.Role,
		AppScope:  newAdmin.AppScope,
		SupervisorID: newAdmin.SupervisorID,
		Permissions: newAdmin.Permissions,
		IsActive:  newAdmin.IsActive,
		CreatedAt: newAdmin.CreatedAt,
		UpdatedAt: newAdmin.UpdatedAt,
	})
}

// GetAllUsers lists all users (super admin only)
func GetAllUsers(c *gin.Context) {
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection is required"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id::text, email, full_name, COALESCE(role, 'admin'),
			COALESCE(app_scope, ''), supervisor_id::text,
			COALESCE(function_permissions, '[]'::jsonb),
			COALESCE(is_active, true), COALESCE(created_at, NOW()), COALESCE(updated_at, NOW()),
			last_login_at
		FROM admin_users
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	defer rows.Close()

	users := []AdminUser{}
	for rows.Next() {
		var u AdminUser
		var supervisorID sql.NullString
		var functionPermissions []byte
		var lastLoginAt sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.AppScope, &supervisorID, &functionPermissions, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &lastLoginAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read users"})
			return
		}
		if supervisorID.Valid {
			u.SupervisorID = &supervisorID.String
		}
		if lastLoginAt.Valid {
			u.LastLoginAt = lastLoginAt.Time
		}
		if err := json.Unmarshal(functionPermissions, &u.Permissions); err != nil {
			u.Permissions = []string{}
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// UpdateUser updates a user's information (super admin only)
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		FullName     string  `json:"full_name"`
		IsActive     *bool   `json:"is_active"`
		Role         string  `json:"role"`
		AppScope     string  `json:"app_scope"`
		SupervisorID *string `json:"supervisor_id"`
		Permissions  *[]string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection is required"})
		return
	}

	existingAdmin, err := getAdminByIDFromDB(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	targetRole := existingAdmin.Role
	if req.Role != "" {
		targetRole = req.Role
	}

	targetAppScope := existingAdmin.AppScope
	if req.AppScope != "" {
		targetAppScope = req.AppScope
	}

	targetSupervisor := existingAdmin.SupervisorID
	if req.SupervisorID != nil {
		if *req.SupervisorID == "" {
			targetSupervisor = nil
		} else {
			targetSupervisor = req.SupervisorID
		}
	}

	if err := validateUserHierarchy(targetRole, targetAppScope, targetSupervisor, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if targetRole == "super_admin" {
		targetAppScope = ""
		targetSupervisor = nil
	}

	targetPermissions := existingAdmin.Permissions
	if req.Permissions != nil {
		targetPermissions = *req.Permissions
	}
	if targetRole == "super_admin" {
		targetPermissions = []string{"*"}
	}
	if targetRole != "super_admin" && len(targetPermissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permissions are required for non-super-admin users"})
		return
	}
	if err := ValidatePermissions(targetPermissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	permissionsJSON, _ := json.Marshal(targetPermissions)

	_, err = database.DB.Exec(`
		UPDATE admin_users
		SET full_name = COALESCE(NULLIF($2, ''), full_name),
			is_active = COALESCE($3, is_active),
			role = $4,
			app_scope = NULLIF($5, ''),
			supervisor_id = $6::uuid,
			function_permissions = $7::jsonb,
			updated_at = NOW()
		WHERE id::text = $1
	`, userID, req.FullName, req.IsActive, targetRole, targetAppScope, targetSupervisor, string(permissionsJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	admin, err := getAdminByIDFromDB(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, AdminUser{
		ID:           admin.ID,
		Email:        admin.Email,
		FullName:     admin.FullName,
		Role:         admin.Role,
		AppScope:     admin.AppScope,
		SupervisorID: admin.SupervisorID,
		Permissions:  admin.Permissions,
		IsActive:     admin.IsActive,
		CreatedAt:    admin.CreatedAt,
		UpdatedAt:    admin.UpdatedAt,
	})
	return
}

// GetAdminUserPermissions returns function permissions of an admin user (super admin only)
func GetAdminUserPermissions(c *gin.Context) {
	userID := c.Param("id")

	var permissions []byte
	err := database.DB.QueryRow(`
		SELECT COALESCE(function_permissions, '[]'::jsonb)
		FROM admin_users
		WHERE id::text = $1
	`, userID).Scan(&permissions)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch permissions"})
		return
	}

	var list []string
	if err := json.Unmarshal(permissions, &list); err != nil {
		list = []string{}
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "permissions": list})
}

// UpdateAdminUserPermissions updates function permissions of an admin user (super admin only)
func UpdateAdminUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permissions are required"})
		return
	}
	if err := ValidatePermissions(req.Permissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role string
	err := database.DB.QueryRow(`SELECT role FROM admin_users WHERE id::text = $1`, userID).Scan(&role)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user"})
		return
	}

	if role == "super_admin" {
		req.Permissions = []string{"*"}
	}
	permissionsJSON, _ := json.Marshal(req.Permissions)

	_, err = database.DB.Exec(`
		UPDATE admin_users
		SET function_permissions = $2::jsonb,
			updated_at = NOW()
		WHERE id::text = $1
	`, userID, string(permissionsJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permissions updated", "user_id": userID, "permissions": req.Permissions})
}

// GetAppUserPermissions returns function permissions of an app user (super admin only)
func GetAppUserPermissions(c *gin.Context) {
	userID := c.Param("id")

	var permissions []byte
	err := database.DB.QueryRow(`
		SELECT COALESCE(function_permissions, '[]'::jsonb)
		FROM users
		WHERE id::text = $1
	`, userID).Scan(&permissions)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "App user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app user permissions"})
		return
	}

	var list []string
	if err := json.Unmarshal(permissions, &list); err != nil {
		list = []string{}
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "permissions": list})
}

// UpdateAppUserPermissions updates function permissions of an app user (super admin only)
func UpdateAppUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permissions are required"})
		return
	}
	if err := ValidatePermissions(req.Permissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permissionsJSON, _ := json.Marshal(req.Permissions)
	result, err := database.DB.Exec(`
		UPDATE users
		SET function_permissions = $2::jsonb,
			updated_at = NOW()
		WHERE id::text = $1
	`, userID, string(permissionsJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update app user permissions"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "App user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app user permissions updated", "user_id": userID, "permissions": req.Permissions})
}

// DeleteUser deletes an admin user (super admin only, cannot delete super admin accounts)
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection is required"})
		return
	}

	var role string
	err := database.DB.QueryRow(`SELECT role FROM admin_users WHERE id::text = $1`, userID).Scan(&role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if role == "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete super admin accounts"})
		return
	}

	if _, err := database.DB.Exec(`DELETE FROM admin_users WHERE id::text = $1`, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
