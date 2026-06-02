# In-Memory Complaint Escalation System

## Overview

The complaint escalation system has been implemented using **in-memory storage** instead of database tables. This allows the escalation system to work without creating additional database schema.

## How It Works

### 1. In-Memory State Storage

The escalation service stores escalation states in memory using Go's `sync.Map`:

```go
type EscalationState struct {
    ComplaintID       string
    CurrentLevel      int
    CurrentTeam       string
    LastEscalatedAt   time.Time
    NextEscalationDue time.Time
    LastNotifiedLevel int
    Category          string
}
```

### 2. Automatic Initialization

When a complaint is fetched from the `report_issues` table:
- If no escalation state exists in memory, it's automatically initialized
- Starts at Level 1 with the appropriate team based on category
- Sets the escalation timer (5 days by default)

### 3. Escalation Levels

Each complaint category has 3 escalation levels:

| Category | Level 1 | Level 2 | Level 3 |
|----------|---------|---------|---------|
| **Service Issue** | Customer Service - Support Agent | Customer Service - Team Lead | Customer Service - Manager |
| **Operations & Scheduling** | Operations Team - Operations Officer | Operations Team - Operations Manager | Operations Team - Operations Director |
| **Vehicle & Facility** | Maintenance Team - Maintenance Technician | Maintenance Team - Maintenance Supervisor | Maintenance Team - Maintenance Manager |
| **Safety & Security** | Safety & Security - Security Officer | Safety & Security - Security Manager | Safety & Security - Security Director |
| **Other** | General Support - Support Agent | General Support - Support Manager | General Support - General Manager |

### 4. Escalation Schedule

- **Check Interval**: Every 1 hour (configured in main.go)
- **Escalation Period**: 5 days per level (configurable in models/escalation.go)
- **Maximum Levels**: 3 levels per category

### 5. SMS Notifications

When a complaint is assigned or escalated:
1. The system retrieves the phone number from `config/assignment.go`
2. Sends an SMS via the eSMS service
3. Logs the notification status

## Configuration

### Team Assignment Configuration

Edit `backend/internal/config/assignment.go` to customize team assignments:

```go
var CategoryAssignment = map[string]map[int]Assignment{
    "Service Issue": {
        1: {"Customer Service", "Support Agent", "94715342627"},
        2: {"Customer Service", "Team Lead", "94715342627"},
        3: {"Customer Service", "Manager", "94715342627"},
    },
    // ... more categories
}
```

### Environment Variables

Add the following to your `.env` file:

```env
# eSMS API Configuration
ESMS_API_URL=https://api.esms.lk/v1/sms/send
ESMS_API_KEY=your_esms_api_key_here
ESMS_SENDER_ID=BusLounge
```

## API Endpoints

### Get Escalation Status

```http
GET /api/complaints/:id/escalation
```

Returns the current escalation level and team for a complaint.

### Manual Escalation

```http
POST /api/complaints/:id/escalate
Content-Type: application/json

{
  "escalated_by": "admin_user_123"
}
```

Manually escalates a complaint to the next level.

### Get Escalation Stats

```http
GET /api/escalation/stats
```

Returns statistics about complaints by escalation level and overdue count.

## How Escalation Works

### Automatic Escalation

1. **Scheduler runs every hour** (configured in main.go)
2. **Queries all unresolved complaints** from `report_issues` table
3. **Checks escalation state** for each complaint:
   - If no state exists → initialize at Level 1
   - If escalation is due → escalate to next level
4. **Sends SMS notification** to the assigned team

### Manual Escalation

1. Admin calls the `/api/complaints/:id/escalate` endpoint
2. System gets current escalation level from memory
3. Escalates to the next level (if available)
4. Sends SMS notification

## Key Features

✅ **No Database Tables Required** - All escalation state is stored in memory
✅ **Automatic Initialization** - Complaints are auto-assigned when first accessed
✅ **SMS Notifications** - Teams are notified via SMS at each escalation
✅ **Configurable Teams** - Easy to customize teams and phone numbers
✅ **Multiple Levels** - 3-level escalation hierarchy
✅ **Hourly Checks** - Automatic escalation checker runs every hour

## Limitations

⚠️ **State is not persistent** - Escalation state is lost when the server restarts
⚠️ **No history tracking** - Escalation history is not saved (only current state)
⚠️ **Memory-based** - All escalation states are stored in RAM

## Workarounds for Persistence

If you need to persist escalation state across server restarts without creating database tables, consider:

1. **File-based storage**: Save escalation states to a JSON file periodically
2. **Redis/Memcached**: Use external in-memory databases
3. **Logging**: Write escalation events to log files for audit trail

## Testing

### Manual Test

1. Start the backend server:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. Create a test complaint in your database

3. Wait for the escalation scheduler to run (or trigger it manually)

4. Check the logs for escalation messages:
   ```
   📋 Initialized escalation for complaint XXX at Level 1
   📱 Sent escalation SMS to Support Agent
   ```

### Trigger Manual Escalation

```bash
curl -X POST http://localhost:8080/api/complaints/COMPLAINT_ID/escalate \
  -H "Content-Type: application/json" \
  -d '{"escalated_by": "admin123"}'
```

## Logs

The system produces detailed logs:

- `📋` Escalation initialization
- `⬆️` Escalation to next level
- `📱` SMS notifications sent
- `✅` Scheduler status

Monitor logs to track escalation activity.

## Support

If SMS notifications are not working:
1. Check your `.env` file has the correct eSMS API credentials
2. Verify phone numbers in `config/assignment.go` include country code
3. Check logs for SMS error messages
