# Quick Start: In-Memory Complaint Escalation

## What Was Implemented

✅ **In-memory escalation system** - No database tables required
✅ **Automatic escalation** - Complaints escalate after 5 days if unresolved
✅ **SMS notifications** - Teams receive SMS when assigned/escalated
✅ **3-level hierarchy** - Operations, Maintenance, Customer Service, Safety teams
✅ **Hourly scheduler** - Checks and escalates complaints every hour

## Setup Steps

### 1. Configure SMS Service

Add to your `.env` file:

```env
# eSMS API Configuration
ESMS_API_URL=https://api.esms.lk/v1/sms/send
ESMS_API_KEY=your_api_key_here
ESMS_SENDER_ID=BusLounge
```

### 2. Configure Team Phone Numbers

Edit `backend/internal/config/assignment.go` and update phone numbers:

```go
var CategoryAssignment = map[string]map[int]Assignment{
    "Service Issue": {
        1: {"Customer Service", "Support Agent", "94715342627"},  // ← Update
        2: {"Customer Service", "Team Lead", "94715342627"},      // ← Update
        3: {"Customer Service", "Manager", "94715342627"},        // ← Update
    },
    // ... update all phone numbers
}
```

### 3. Start the Server

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
✅ Complaint escalation scheduler started (runs every 1 hour)
```

## How It Works

### Automatic Flow

1. **Complaint created** → System initializes escalation at Level 1
2. **SMS sent** → Assigned team member receives notification
3. **After 5 days** (if unresolved) → Auto-escalate to Level 2
4. **SMS sent** → Level 2 team receives notification
5. **After 5 more days** (if still unresolved) → Escalate to Level 3
6. **SMS sent** → Level 3 team receives notification

### Category to Team Mapping

| Complaint Category | Team |
|-------------------|------|
| bus_delay | Operations & Scheduling Team |
| maintenance_issue, flat_wheel, equipment_malfunction | Vehicle & Facility Team |
| passenger_complaint | Service Issue Team |
| safety_concern | Safety & Security Team |
| Other | General Support Team |

## Testing

### Test Automatic Escalation

1. Create a test complaint in the database
2. Check server logs - should see:
   ```
   📋 Initialized escalation for complaint XXX at Level 1 (Customer Service)
   📱 Sent escalation SMS to Support Agent (94715342627)
   ```

### Test Manual Escalation

```bash
curl -X POST http://localhost:8080/api/complaints/COMPLAINT_ID/escalate \
  -H "Content-Type: application/json" \
  -d '{"escalated_by": "admin123"}'
```

### Check Escalation Stats

```bash
curl http://localhost:8080/api/escalation/stats
```

Response:
```json
{
  "by_level": {
    "1": 5,
    "2": 2,
    "3": 1
  },
  "overdue_count": 3,
  "total_tracked": 3
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/complaints/:id/escalation` | GET | Get escalation status |
| `/api/complaints/:id/escalate` | POST | Manual escalate |
| `/api/escalation/stats` | GET | Get statistics |
| `/api/escalation/config` | GET | Get escalation configuration |

## Important Notes

⚠️ **State is in-memory** - Escalation state resets when server restarts
⚠️ **No persistence** - For production, consider adding file-based or Redis storage
✅ **Works without DB tables** - Uses only the existing `report_issues` table

## Customization

### Change Escalation Period

Edit `backend/internal/models/escalation.go`:

```go
EscalationDays: 5,  // Change from 5 to your desired number of days
```

### Change Check Interval

Edit `backend/cmd/server/main.go`:

```go
escalationScheduler := services.NewEscalationScheduler(database.DB, cfg, 1*time.Hour)
//                                                                       ↑ Change this
```

### Add Custom Categories

Edit `backend/internal/config/assignment.go` and add new category mappings.

## Troubleshooting

### SMS Not Sending

1. Verify `.env` has correct eSMS credentials
2. Check phone numbers include country code (e.g., 94715342627)
3. Look for error logs: `Failed to send SMS notification`

### Escalation Not Working

1. Check server logs for escalation scheduler messages
2. Verify complaints exist with status NOT 'resolved' or 'closed'
3. Check escalation timer: `next_escalation_due` in memory state

### No Logs Appearing

1. Ensure scheduler started: Look for `✅ Complaint escalation scheduler started`
2. Wait for hourly check or restart server to trigger immediate check
3. Create a test complaint to initialize escalation

## Monitoring

Watch server logs for these indicators:

- `📋` Escalation initialized
- `⬆️` Complaint escalated
- `📱` SMS notification sent
- `Running scheduled escalation check...` - Hourly scheduler running
- `Successfully escalated X complaint(s)` - Escalations completed

## Next Steps

If you need persistent storage without database tables:

1. Add file-based storage (JSON export/import)
2. Use Redis for distributed in-memory storage
3. Implement event logging to files
4. Add database tables (see `backend/complaint_escalation_schema.sql`)
