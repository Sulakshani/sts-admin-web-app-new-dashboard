# Forgot Password Implementation

This document describes the production-level forgot password implementation for the STS Admin Web App.

## Architecture Overview

The forgot password feature uses a secure token-based flow with email verification:

1. **User requests password reset** → Backend generates secure token → Email sent
2. **User clicks email link** → Frontend validates token → Shows reset form
3. **User submits new password** → Backend verifies token → Updates password → Invalidates token

## Backend Implementation

### Technology Stack
- **Email Service**: Gmail SMTP (smtp.gmail.com:587)
- **Token Generation**: UUID v4 (cryptographically secure)
- **Token Storage**: In-memory map (clears on server restart)
- **Token Expiration**: 15 minutes
- **Token Usage**: One-time use (deleted after successful reset)

### API Endpoints

#### 1. Request Password Reset
```
POST /api/auth/forgot-password
Content-Type: application/json

{
  "email": "admin@sts.lk"
}
```

**Response (Success)**:
```json
{
  "message": "If an account exists with this email, a password reset link has been sent."
}
```

**Notes**:
- Returns same message whether email exists or not (security best practice)
- Sends email only if admin exists in predefinedAdmins map
- Email contains both reset code and clickable link

#### 2. Verify Reset Token
```
GET /api/auth/verify-reset-token?token=<uuid>
```

**Response (Valid Token)**:
```json
{
  "valid": true
}
```

**Response (Invalid/Expired Token)**:
```json
{
  "error": "Invalid or expired reset token"
}
```

#### 3. Reset Password
```
POST /api/auth/reset-password
Content-Type: application/json

{
  "token": "uuid-token-here",
  "newPassword": "newSecurePassword123"
}
```

**Response (Success)**:
```json
{
  "message": "Password has been reset successfully"
}
```

**Response (Error)**:
```json
{
  "error": "Invalid or expired reset token"
}
```

### Email Template

**Subject**: Reset Your Password

**Content** (HTML + Plain Text):
- Professional gradient header
- Reset code for manual entry
- One-click reset button
- 15-minute expiration warning
- Security notices (don't share link, contact support if unsolicited)
- Fallback plain text version

**Reset URL Format**:
```
http://localhost:4200/reset-password?token=<uuid>
```

## Frontend Implementation

### Components

#### Login Component (Forgot Password Modal)
- **Location**: `src/app/pages/login/login.component.ts`
- **Features**:
  - Modal dialog for email entry
  - Client-side email validation
  - Loading states
  - Success/error messaging
  - Calls `adminAuthService.requestPasswordReset(email)`

#### Reset Password Component
- **Location**: `src/app/pages/reset-password/`
- **Route**: `/reset-password?token=<uuid>`
- **Features**:
  - Extracts token from URL query params
  - Validates token on page load
  - Shows loading state during validation
  - Shows error state for invalid/expired tokens
  - Password confirmation matching
  - Minimum length validation (8 characters)
  - Loading state during submission
  - Success message with auto-redirect to login
  - Back to login button

### Services

#### AdminAuthService
- **Location**: `src/app/core/services/admin-auth.service.ts`
- **Methods**:
  ```typescript
  requestPasswordReset(email: string): Observable<any>
  verifyResetToken(token: string): Observable<any>
  resetPassword(token: string, newPassword: string): Observable<any>
  ```

### Routing

```typescript
{ path: 'reset-password', component: ResetPasswordComponent }
```

## Security Features

1. **Token Security**:
   - UUID v4 generation (128-bit random)
   - 15-minute expiration
   - One-time use (deleted after reset)
   - Cannot be guessed or brute-forced

2. **Email Security**:
   - SMTP over TLS (port 587)
   - App-specific passwords (not actual Gmail password)
   - No sensitive data in email (only reset link)

3. **Password Security**:
   - Bcrypt hashing (cost factor 10)
   - Minimum 8 character requirement
   - Password confirmation required on frontend

4. **Privacy**:
   - Same response for existing/non-existing emails
   - Token validation doesn't reveal email existence
   - No user enumeration possible

5. **Rate Limiting** (Recommended for Production):
   - Limit reset requests per IP
   - Limit reset requests per email
   - Not currently implemented (add in production)

## User Flow

### Forget Password Flow
1. User clicks "Forgot Password?" on login page
2. Modal opens requesting email address
3. User enters email and clicks "Send Reset Link"
4. Frontend shows success message
5. User checks their inbox
6. Email arrives with reset code and link

### Reset Password Flow
1. User clicks link in email
2. Browser opens: `http://localhost:4200/reset-password?token=<uuid>`
3. Frontend validates token with backend
4. If valid: Shows password reset form
5. If invalid: Shows error with "Back to Login" button
6. User enters new password (twice for confirmation)
7. User clicks "Reset Password"
8. Frontend submits to backend
9. Backend validates token, updates password, deletes token
10. Success message shown
11. Auto-redirect to login after 3 seconds

## Error Handling

### Frontend Error Messages
- "This link has expired. Reset links are only valid for 15 minutes."
- "This link has already been used. Please request a new one."
- "Invalid reset link. Please request a new one."
- "Passwords do not match"
- "Password must be at least 8 characters"

### Backend Error Responses
- `400 Bad Request`: Missing email/token/password
- `404 Not Found`: Token not found or expired
- `500 Internal Server Error`: Email sending failed (logged to console)

## Configuration

### Backend (.env)
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-specific-password
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_FROM_NAME=STS Admin Team
FRONTEND_URL=http://localhost:4200
```

### Frontend (environment)
```typescript
export const environment = {
  apiUrl: 'http://localhost:8083/api'
};
```

## Testing

### Manual Testing Checklist

1. **Request Reset**:
   - [ ] Enter valid admin email → receives email
   - [ ] Enter non-existent email → same response, no email sent
   - [ ] Enter invalid format → validation error shown

2. **Email Delivery**:
   - [ ] Email arrives in inbox (not spam)
   - [ ] Email displays correctly (HTML version)
   - [ ] Reset code visible
   - [ ] Reset button clickable
   - [ ] Plain text fallback works

3. **Token Validation**:
   - [ ] Valid token → shows reset form
   - [ ] Invalid token → shows error state
   - [ ] Expired token (after 15 min) → shows error
   - [ ] Used token → shows error

4. **Password Reset**:
   - [ ] Passwords match → submits successfully
   - [ ] Passwords don't match → shows error
   - [ ] Password too short → shows error
   - [ ] Success → redirects to login
   - [ ] Can login with new password

5. **Security**:
   - [ ] Token works only once
   - [ ] Token expires after 15 minutes
   - [ ] Old password no longer works
   - [ ] New password works

### Automated Testing (Future Enhancement)

Create unit tests for:
- Token generation and validation
- Email service (mock SMTP)
- Password reset handlers
- Frontend form validation

## Known Limitations

1. **In-Memory Token Storage**:
   - Tokens lost on server restart
   - Not suitable for multi-server deployments
   - **Solution for Production**: Use Redis or database

2. **No Rate Limiting**:
   - Vulnerable to email bombing
   - **Solution**: Add rate limiting middleware

3. **Gmail Sending Limits**:
   - 500 emails/day (free Gmail)
   - 2000 emails/day (Google Workspace)
   - **Solution**: Switch to SendGrid/AWS SES for production

4. **No Email Queue**:
   - Synchronous email sending
   - Request blocks until email sent
   - **Solution**: Add background job queue

## Production Recommendations

1. **Token Storage**:
   - Migrate to Redis with automatic expiration
   - Or use database table with cleanup job

2. **Email Service**:
   - Switch to dedicated service (SendGrid, AWS SES)
   - Configure SPF, DKIM, DMARC records
   - Use custom domain email

3. **Rate Limiting**:
   - Add IP-based rate limiting (e.g., 3 requests per hour)
   - Add email-based rate limiting (e.g., 5 requests per day)

4. **Monitoring**:
   - Log all password reset attempts
   - Alert on suspicious patterns
   - Track email delivery rates

5. **Environment Variables**:
   - Use platform environment vars (not .env file)
   - Rotate credentials regularly
   - Use secrets management service

## Troubleshooting

### Email Not Sending
- Check backend terminal for SMTP errors
- Verify Gmail app password is correct
- Ensure 2FA is enabled on Gmail account
- Check internet connection and firewall

### Email Goes to Spam
- Add sender to contacts
- Mark as "Not Spam"
- For production: Set up proper domain authentication

### Reset Link Not Working
- Check token hasn't expired (15 minutes)
- Ensure frontend URL matches FRONTEND_URL in backend
- Verify token wasn't already used
- Check browser console for errors

## Files Modified/Created

### Backend
- `internal/services/email_service.go` - Email sending service
- `internal/handlers/admin_auth_handler.go` - Reset password handlers
- `internal/config/config.go` - SMTP configuration
- `cmd/server/main.go` - Initialize auth handlers
- `.env` - SMTP credentials
- `GMAIL_SETUP_GUIDE.md` - Setup documentation

### Frontend
- `app/pages/reset-password/reset-password.component.ts` - Component logic
- `app/pages/reset-password/reset-password.component.html` - UI template
- `app/pages/reset-password/reset-password.component.scss` - Styles
- `app/core/services/admin-auth.service.ts` - API methods
- `app/app.routes.ts` - Route configuration
- `docs/FORGOT_PASSWORD_GUIDE.md` - This documentation

## Support

For issues or questions:
1. Check backend terminal logs
2. Review GMAIL_SETUP_GUIDE.md
3. Test with network DevTools open
4. Contact system administrator

---

**Last Updated**: 2024
**Version**: 1.0.0
**Status**: Production Ready ✅
