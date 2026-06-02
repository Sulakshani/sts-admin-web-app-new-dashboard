# Dialog SMS Configuration Guide

## Overview

The complaint escalation system now supports **Dialog eSMS** for sending SMS notifications with two modes: **dev** (testing) and **production** (real SMS).

## Configuration Options

### 1. SMS Mode

Set in `.env` file:

```env
SMS_MODE=dev          # For development (logs only, no real SMS)
SMS_MODE=production   # For production (sends real SMS via Dialog)
```

### 2. Dialog SMS Methods

#### Method 1: URL Method (Recommended) ✅

Uses GET request with esmsqk key. More reliable and doesn't require login endpoint.

```env
DIALOG_SMS_METHOD=url
DIALOG_SMS_ESMSQK=your_esmsqk_key_here
DIALOG_SMS_MASK=KanchTest
```

**How to get your esmsqk key:**
1. Login to https://e-sms.dialog.lk
2. Navigate to "URL Message Key" section
3. Copy your esmsqk token

#### Method 2: API v2 Method (Alternative)

Uses POST request with username/password. Requires access to `/api/v2/login` endpoint.

```env
DIALOG_SMS_METHOD=api_v2
DIALOG_SMS_API_URL=https://e-sms.dialog.lk/api/v2
DIALOG_SMS_USERNAME=your_username
DIALOG_SMS_PASSWORD=your_password
DIALOG_SMS_MASK=KanchTest
```

## Current Configuration

Your `.env` file is currently configured with:

```env
# SMS Mode: "dev" (no actual SMS, log only) or "production" (real SMS via Dialog)
SMS_MODE=dev

# Dialog SMS Method: "url" (recommended)
DIALOG_SMS_METHOD=url

# Dialog eSMS URL Method Configuration
DIALOG_SMS_ESMSQK=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NzI2OSwiY3VzdG9tZXJfcm9sZSI6MCwiaWF0IjoxNzYxMjkzMTcxLCJleHAiOjQ4ODU0OTU1NzF9.wGg7hN7CczGQMcuHJ9hLovTccgJK79injwhg7laBqro
DIALOG_SMS_MASK=KanchTest

# Alternative API v2 credentials (if needed)
DIALOG_SMS_USERNAME=kanchanadesilva
DIALOG_SMS_PASSWORD=Dialog@123
```

## How It Works

### Dev Mode (`SMS_MODE=dev`)

When in dev mode:
- ✅ No actual SMS is sent
- ✅ SMS content is logged to console
- ✅ Perfect for testing escalation logic
- ✅ No SMS credits consumed

**Log output example:**
```
📱 [DEV MODE] SMS not sent. Recipient: 94715342627, Message: Hello Support Agent...
```

### Production Mode (`SMS_MODE=production`)

When in production mode:
- ✅ Real SMS is sent via Dialog eSMS
- ✅ Uses configured method (URL or API v2)
- ✅ Escalation notifications sent to team members
- ✅ SMS credits consumed

**Log output example:**
```
✅ Dialog SMS sent successfully to 94715342627
```

## Testing

### Test in Dev Mode

1. Set `SMS_MODE=dev` in `.env`
2. Restart the server
3. Create a test complaint or trigger escalation
4. Check logs for `[DEV MODE]` messages

### Test in Production Mode

1. Set `SMS_MODE=production` in `.env`
2. Ensure Dialog SMS credentials are correct
3. Restart the server
4. Create a test complaint or trigger manual escalation
5. Check logs for successful SMS delivery

## Switching Between Modes

### For Local Development
```env
SMS_MODE=dev
```

### For Testing with Real SMS
```env
SMS_MODE=production
DIALOG_SMS_METHOD=url
```

### For Production Deployment
```env
SMS_MODE=production
DIALOG_SMS_METHOD=url
DIALOG_SMS_ESMSQK=your_production_esmsqk_key
DIALOG_SMS_MASK=YourProductionMask
```

## API Endpoints

The Dialog eSMS API is called automatically when:
- New complaint is assigned → SMS to Level 1 team
- Complaint escalates → SMS to next level team
- Manual escalation triggered → SMS to escalated team

### URL Method Endpoint
```
GET https://e-sms.dialog.lk/api/sms/send?esmsqk={key}&message={msg}&target={phone}&mask={mask}
```

### API v2 Method Endpoint
```
POST https://e-sms.dialog.lk/api/v2/sms/send
Content-Type: application/json

{
  "username": "...",
  "password": "...",
  "message": "...",
  "msisdn": "94715342627",
  "alias": "KanchTest"
}
```

## Phone Number Format

Phone numbers in [assignment.go](internal/config/assignment.go) must include country code:

✅ Correct: `94715342627`  
❌ Wrong: `0715342627`  
❌ Wrong: `+94715342627`

## Troubleshooting

### SMS Not Sending in Production Mode

1. **Check SMS Mode**
   ```bash
   # In .env file
   SMS_MODE=production  # Must be "production", not "dev"
   ```

2. **Verify Credentials**
   - For URL method: Check `DIALOG_SMS_ESMSQK` is valid
   - For API v2: Check `DIALOG_SMS_USERNAME` and `DIALOG_SMS_PASSWORD`

3. **Check Phone Numbers**
   - Must include country code (94 for Sri Lanka)
   - No spaces or special characters
   - Example: `94715342627`

4. **Check Server Logs**
   ```
   ✅ Dialog SMS sent successfully to ...  # Success
   ❌ Failed to send Dialog SMS: ...       # Error with details
   ```

### Dev Mode Not Working

1. Ensure `SMS_MODE=dev` in `.env`
2. Restart the server
3. Look for `[DEV MODE]` in logs

### SMS Sent But Not Received

1. Verify phone number is correct in [assignment.go](internal/config/assignment.go)
2. Check Dialog eSMS account has sufficient credits
3. Verify mask/sender ID is approved by Dialog
4. Check Dialog eSMS dashboard for delivery status

## Cost Optimization

### Development
- Use `SMS_MODE=dev` to avoid consuming SMS credits
- Test escalation logic without sending real SMS

### Production
- Monitor SMS usage in Dialog eSMS dashboard
- Use appropriate escalation intervals to reduce SMS volume
- Consider email notifications as backup

## Security Notes

🔒 **Important Security Practices:**

1. **Never commit `.env` file** to version control
2. **Use different credentials** for dev/staging/production
3. **Rotate esmsqk key** periodically
4. **Restrict access** to Dialog eSMS dashboard
5. **Monitor usage** for unusual activity

## Support

### Dialog eSMS Support
- Website: https://e-sms.dialog.lk
- Documentation: Check Dialog eSMS portal
- Support: Contact Dialog customer support

### System Configuration
- See [ESCALATION_QUICK_START.md](ESCALATION_QUICK_START.md) for escalation setup
- See [IN_MEMORY_ESCALATION_GUIDE.md](IN_MEMORY_ESCALATION_GUIDE.md) for escalation details
- Update team phone numbers in [assignment.go](internal/config/assignment.go)

## Quick Reference

| Environment Variable | Purpose | Example |
|---------------------|---------|---------|
| `SMS_MODE` | dev or production | `dev` |
| `DIALOG_SMS_METHOD` | url or api_v2 | `url` |
| `DIALOG_SMS_ESMSQK` | esmsqk API key | `eyJhbGci...` |
| `DIALOG_SMS_MASK` | Sender ID/Mask | `KanchTest` |
| `DIALOG_SMS_USERNAME` | API v2 username | `kanchanadesilva` |
| `DIALOG_SMS_PASSWORD` | API v2 password | `Dialog@123` |
