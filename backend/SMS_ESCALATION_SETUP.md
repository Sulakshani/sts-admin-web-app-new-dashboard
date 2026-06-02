# SMS Escalation System - Setup Guide

## Overview

The complaint escalation system automatically escalates unresolved complaints through multiple levels and sends SMS notifications to assigned teams at each level.

## Features

- ✅ Automatic escalation after 5 days (configurable per level)
- ✅ Manual escalation via API endpoint
- ✅ SMS notifications via eSMS API
- ✅ Configurable team assignments with phone numbers
- ✅ Maximum 3 escalation levels per category
- ✅ Escalation history tracking
- ✅ Statistics and monitoring

## Environment Configuration

Add the following to your `.env` file:

```env
# eSMS API Configuration
ESMS_API_URL=https://api.esms.lk/v1/sms/send
ESMS_API_KEY=your_esms_api_key_here
ESMS_SENDER_ID=BusLounge
```

### Getting eSMS API Credentials

1. Sign up at your eSMS provider
2. Navigate to API settings
3. Generate API key
4. Register your Sender ID (e.g., "BusLounge")
5. Add the credentials to your `.env` file

## Team Assignment Configuration

Edit `backend/internal/config/assignment.go` to customize team assignments:

```go
var CategoryAssignment = map[string]map[int]Assignment{
    "Service Issue": {
        1: {"Customer Service", "Support Agent", "94771234567"},
        2: {"Customer Service", "Team Lead", "94771234568"},
        3: {"Customer Service", "Manager", "94771234569"},
    },
    "Operations & Scheduling": {
        1: {"Operations Team", "Operations Officer", "94771234570"},
        2: {"Operations Team", "Operations Manager", "94771234571"},
        3: {"Operations Team", "Operations Director", "94771234572"},
    },
    // Add more categories...
}
```

### Assignment Structure

Each assignment contains:
- **TeamName**: The team responsible for this level
- **RoleName**: The role/position of the assigned person
- **PhoneNumber**: Mobile number for SMS notifications (include country code)

## Escalation Categories

The system maps complaint `issue_type` to escalation categories:

| Issue Type | Escalation Category |
|-----------|-------------------|
| `bus_delay` | Operations & Scheduling |
| `maintenance_issue`, `flat_wheel`, `equipment_malfunction` | Vehicle & Facility |
| `passenger_complaint` | Service Issue |
| `safety_concern` | Safety & Security |
| Other | Other |

You can customize this mapping in `backend/internal/services/complaint_service.go`.

## Escalation Schedule

- **Check Interval**: Every 1 hour (configurable in `main.go`)
- **Escalation Period**: 5 days per level (configurable in `escalation.go`)
- **Maximum Levels**: 3 levels per category

## API Endpoints

### Manual Escalation

```http
POST /api/escalation/complaint/:id/escalate
Content-Type: application/json

{
  "category": "Service Issue",
  "current_level": 1,
  "escalated_by": "admin_user_123"
}
```

### Get Escalation Info

```http
GET /api/escalation/complaint/:id
```

### Get Escalation History

```http
GET /api/escalation/complaint/:id/history
```

### Initialize Escalation

```http
POST /api/escalation/complaint/:id/initialize
Content-Type: application/json

{
  "category": "Service Issue"
}
```

### Get Escalation Statistics

```http
GET /api/escalation/stats
```

## SMS Templates

### New Complaint Assignment

```
Hello {RoleName},

You have been assigned a new complaint:

Complaint ID: {ID}
Category: {Category}

Please review and respond within 5 days.

Thank you!
```

### Complaint Escalated

```
Hello {RoleName},

A complaint has been escalated to you (Level {Level}):

Complaint ID: {ID}
Category: {Category}

This requires urgent attention. Please review immediately.

Thank you!
```

## How It Works

### Automatic Escalation Flow

1. **Day 0**: Complaint created → Assigned to Level 1 → Escalation due in 5 days → SMS sent to Level 1 team
2. **Day 5**: If unresolved → Auto-escalate to Level 2 → New due date in 5 days → SMS sent to Level 2 team
3. **Day 10**: If unresolved → Auto-escalate to Level 3 (Manager) → SMS sent to Level 3 team
4. **Day 15+**: Stays at Level 3 (no further escalation)

### Manual Escalation

Admins can manually escalate any complaint before the automatic escalation using the API endpoint.

## Monitoring & Logs

The system provides detailed logging:

```
✅ Complaint escalation scheduler started (runs every 1 hour)
🔍 Running scheduled escalation check...
📊 Found 3 complaint(s) that need escalation
⬆️  Escalated complaint abc-123 from level 1 to level 2
📱 Sending escalation notification (Level 2) to 94771234568
✅ SMS sent successfully to 94771234568 (Reference: SMS-REF-123)
✅ Escalation process completed. Processed 3 complaint(s)
```

## Troubleshooting

### SMS Not Sending

1. **Check configuration**:
   ```bash
   # Verify environment variables are set
   echo $ESMS_API_KEY
   echo $ESMS_API_URL
   ```

2. **Check logs** for errors:
   ```
   ❌ Failed to send SMS to 94771234567: SMS API error
   ⚠️  SMS service not configured - skipping SMS notification
   ```

3. **Verify API credentials** with eSMS provider

4. **Test manually**:
   ```bash
   curl -X POST https://api.esms.lk/v1/sms/send \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_API_KEY" \
     -d '{
       "recipient": "94771234567",
       "senderId": "BusLounge",
       "message": "Test message"
     }'
   ```

### Escalation Not Working

1. **Check escalation scheduler** is running:
   ```
   ✅ Complaint escalation scheduler started
   ```

2. **Verify complaint status** is not `resolved` or `closed`

3. **Check next_escalation_due** date in database:
   ```sql
   SELECT complaint_id, current_level, next_escalation_due 
   FROM complaint_escalations;
   ```

4. **Manually trigger escalation check** (for testing):
   ```go
   escalationScheduler.RunEscalationCheckNow()
   ```

## Database Schema

The system uses two tables:

### complaint_escalations

Tracks current escalation status:
```sql
CREATE TABLE complaint_escalations (
    complaint_id VARCHAR PRIMARY KEY,
    current_level INTEGER,
    current_team VARCHAR,
    assigned_to_admin_id UUID,
    last_escalated_at TIMESTAMP,
    next_escalation_due TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### complaint_escalation_history

Tracks escalation history:
```sql
CREATE TABLE complaint_escalation_history (
    id SERIAL PRIMARY KEY,
    complaint_id VARCHAR,
    level INTEGER,
    team_name VARCHAR,
    escalated_at TIMESTAMP DEFAULT NOW(),
    escalated_by VARCHAR,
    reason TEXT
);
```

## Testing

### Test Manual Escalation

```bash
# Escalate complaint to next level
curl -X POST http://localhost:8080/api/escalation/complaint/your-complaint-id/escalate \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Service Issue",
    "current_level": 1,
    "escalated_by": "admin_test"
  }'
```

### Test Escalation Stats

```bash
curl http://localhost:8080/api/escalation/stats
```

## Production Deployment

1. **Set environment variables** in your production environment
2. **Configure team assignments** with real phone numbers
3. **Test SMS delivery** before going live
4. **Monitor logs** for any errors
5. **Set up alerts** for failed SMS notifications

## Cost Considerations

- Each SMS notification costs money (depends on your eSMS provider)
- Monitor SMS usage to avoid unexpected costs
- Consider implementing daily SMS limits if needed
- The system gracefully handles SMS failures without blocking escalation

## Support

For issues or questions:
- Check the logs first
- Verify environment configuration
- Test API credentials
- Review team assignments
