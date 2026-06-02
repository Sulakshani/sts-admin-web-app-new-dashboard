# Gmail SMTP Setup for Password Reset Emails

This guide will help you configure Gmail to send password reset emails for your application.

## Prerequisites

- A Gmail account (personal or Google Workspace)
- Access to backend `.env` file

## Step 1: Enable 2-Factor Authentication (2FA)

1. Go to [Google Account Security](https://myaccount.google.com/security)
2. Click on **2-Step Verification** (under "How you sign in to Google")
3. Follow the prompts to set up 2FA with your phone number
4. Once enabled, verify it's working by signing out and signing back in

## Step 2: Generate App Password

1. Go to [Google App Passwords](https://myaccount.google.com/apppasswords)
   - If you don't see this option, make sure 2FA is enabled first
2. Under "Select app", choose **Mail**
3. Under "Select device", choose **Other (Custom name)**
4. Enter a name like `STS Admin Password Reset`
5. Click **Generate**
6. **Copy the 16-character app password** (it will be shown in a yellow box)
   - Format: `xxxx xxxx xxxx xxxx` (remove spaces when copying)
   - This is shown only once, so copy it immediately

## Step 3: Update Backend `.env` File

Open your `backend/.env` file and update the following values:

```env
# Email Configuration (Gmail SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com          # ← Replace with your Gmail address
SMTP_PASSWORD=xxxx xxxx xxxx xxxx           # ← Replace with the app password from Step 2
SMTP_FROM_EMAIL=your-email@gmail.com        # ← Replace with your Gmail address
SMTP_FROM_NAME=STS Admin Team
FRONTEND_URL=http://localhost:4200
```

### Important Notes:

- **SMTP_USERNAME**: Your full Gmail address (e.g., `admin@gmail.com`)
- **SMTP_PASSWORD**: The 16-character app password (you can include or remove spaces)
- **SMTP_FROM_EMAIL**: Same as SMTP_USERNAME
- **SMTP_FROM_NAME**: Display name that appears in emails (can be customized)
- **FRONTEND_URL**: Your frontend URL (change for production deployment)

## Step 4: Restart Backend Server

After updating the `.env` file:

1. Stop the backend server (Ctrl+C in the terminal)
2. Start it again:
   ```powershell
   cd backend
   go run cmd/server/main.go
   ```

## Step 5: Test Password Reset Flow

1. Open frontend at http://localhost:4200
2. Click "Login"
3. Click "Forgot Password?"
4. Enter an admin email (e.g., `admin@sts.lk` or `superadmin@sts.lk`)
5. Click "Send Reset Link"
6. Check the Gmail inbox associated with that email for the reset email

### Expected Email Content:

- **Subject**: Reset Your Password
- **From**: STS Admin Team <your-email@gmail.com>
- **Contains**: 
  - Reset code (for manual entry)
  - Reset link button (for one-click)
  - Expiration time (15 minutes)

## Troubleshooting

### Email Not Sending

1. **Check Gmail Credentials**:
   - Verify SMTP_USERNAME is your correct Gmail address
   - Verify SMTP_PASSWORD is the app password (not your Gmail password)
   - Make sure there are no extra spaces in the .env file

2. **Check 2FA Status**:
   - Visit [Google Account Security](https://myaccount.google.com/security)
   - Confirm 2-Step Verification is ON

3. **Check Backend Logs**:
   - Look for error messages in the backend terminal
   - Common errors:
     - `535 Authentication failed`: Wrong username/password
     - `Connection refused`: Firewall blocking port 587

4. **Test SMTP Credentials**:
   You can test your credentials manually:
   ```bash
   telnet smtp.gmail.com 587
   ```

### Email Goes to Spam

1. Add your sending Gmail account to your contacts
2. Mark the email as "Not Spam" in Gmail
3. For production, set up proper SPF/DKIM records

### App Password Not Available

- Ensure 2FA is enabled on your Google account
- You may need to wait a few minutes after enabling 2FA
- Try using Google Chrome in incognito mode to access the App Passwords page

## Security Best Practices

1. **Never commit `.env` to Git** (already in `.gitignore`)
2. **Use different email accounts** for development and production
3. **Rotate app passwords** periodically (recommended every 6 months)
4. **Monitor sent emails** in Gmail's Sent folder for suspicious activity
5. **For production**:
   - Consider using a dedicated email service (SendGrid, AWS SES, Mailgun)
   - Use environment variables from hosting platform (not .env file)
   - Set up proper domain authentication (SPF, DKIM, DMARC)

## Gmail Sending Limits

- **Free Gmail**: ~500 emails per day
- **Google Workspace**: ~2,000 emails per day

For higher volumes, consider switching to a dedicated email service.

## Production Deployment Notes

When deploying to production:

1. Update `FRONTEND_URL` in `.env` to your production domain
2. Use environment variables from your hosting platform
3. Consider using a service like:
   - **SendGrid** (free tier: 100 emails/day)
   - **AWS SES** (pay-as-you-go, very cheap)
   - **Mailgun** (free tier: limited)
4. Set up custom domain email for better deliverability

## Testing Checklist

- [ ] 2FA enabled on Gmail account
- [ ] App password generated and copied
- [ ] `.env` file updated with correct credentials
- [ ] Backend server restarted
- [ ] Password reset email received in inbox
- [ ] Reset link works and redirects to reset page
- [ ] New password successfully updates
- [ ] Can login with new password

---

**Need Help?** Check the backend terminal for error messages or contact your system administrator.
