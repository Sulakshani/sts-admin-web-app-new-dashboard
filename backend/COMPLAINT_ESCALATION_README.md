# Complaint Escalation System

## Overview

The Complaint Escalation System automatically manages and escalates complaints through defined teams and levels based on complaint categories. If a complaint is not resolved within the specified timeframe (5 days by default), it automatically escalates to the next level.

## Features

- ✅ **Automatic Escalation**: Complaints escalate automatically after 5 days if unresolved
- ✅ **Category-Based Teams**: Each complaint category has dedicated teams
- ✅ **Multi-Level Hierarchy**: 3-level escalation hierarchy (Level 1 → Level 2 → Level 3)
- ✅ **Assignment Tracking**: Track which admin is assigned to each complaint
- ✅ **Escalation History**: Complete audit trail of all escalation events
- ✅ **Manual Escalation**: Admins can manually escalate complaints when needed
- ✅ **Statistics Dashboard**: View escalation stats and overdue complaints

## Escalation Configuration

### Categories and Teams

| Category | Level 1 | Level 2 | Level 3 | Days per Level |
|----------|---------|---------|---------|----------------|
| **Bus Delay** | Operations Support Team | Operations Manager | Senior Management | 5 days |
| **Maintenance Issue** | Maintenance Team | Maintenance Supervisor | Technical Manager | 5 days |
| **Flat Wheel** | Maintenance Team | Fleet Supervisor | Technical Manager | 5 → 2 days |
| **Passenger Complaint** | Customer Service Team | Customer Service Manager | Head of Customer Experience | 5 days |
| **Safety Concern** | Safety & Compliance Team | Safety Officer | Head of Safety & Compliance | 5 → 2 days |
| **Equipment Malfunction** | Technical Support Team | Technical Supervisor | Technical Manager | 5 days |
| **Other** | General Support Team | Support Manager | Senior Management | 5 days |

## Database Setup

### 1. Run the Migration

Execute the SQL migration to create escalation tables:

```bash
cd backend
psql -U your_username -d your_database -f complaint_escalation_schema.sql
```

Or manually execute via your PostgreSQL client:

```sql
-- complaint_escalation_schema.sql contains:
-- - complaint_escalations table
-- - complaint_escalation_history table
-- - Indexes for performance
-- - Triggers for auto-updating timestamps
```

### 2. Initialize Existing Complaints

After running the migration, initialize escalation for existing complaints:

```sql
-- This would be done programmatically or via a one-time script
-- For each existing complaint in report_issues table
```

## API Endpoints

### Get Escalation Configuration

```http
GET /api/escalation/config
```

Returns escalation configuration for all categories.

**Response:**
```json
{
  "configs": {
    "bus_delay": {
      "category": "bus_delay",
      "levels": [
        {
          "level": 1,
          "team_name": "Operations Support Team",
          "escalation_days": 5,
          "notification_emails": ["operations@aasl.lk"]
        },
        ...
      ]
    },
    ...
  }
}
```

### Get Category-Specific Configuration

```http
GET /api/escalation/config/:category
```

Example: `/api/escalation/config/bus_delay`

### Get Complaint Escalation Status

```http
GET /api/escalation/complaint/:id
```

**Response:**
```json
{
  "current_level": 1,
  "current_team": "Operations Support Team",
  "assigned_to_admin_id": "uuid-here",
  "assigned_to_name": "John Doe",
  "last_escalated_at": "2026-03-01T10:00:00Z",
  "next_escalation_due": "2026-03-06T10:00:00Z",
  "escalation_history": [
    {
      "level": 1,
      "team_name": "Operations Support Team",
      "escalated_at": "2026-03-01T10:00:00Z",
      "escalated_by": "system",
      "reason": "Initial assignment"
    }
  ]
}
```

### Get Escalation History

```http
GET /api/escalation/complaint/:id/history
```

Returns complete escalation history for a complaint.

### Manual Escalation

```http
POST /api/escalation/complaint/:id/escalate
Content-Type: application/json

{
  "category": "bus_delay",
  "current_level": 1,
  "escalated_by": "admin-uuid"
}
```

Manually escalates a complaint to the next level.

### Assign Complaint to Admin

```http
POST /api/escalation/complaint/:id/assign
Content-Type: application/json

{
  "admin_id": "admin-uuid-here"
}
```

Assigns a complaint to a specific admin user.

### Initialize Escalation for New Complaint

```http
POST /api/escalation/complaint/:id/initialize
Content-Type: application/json

{
  "category": "bus_delay"
}
```

Initializes escalation tracking for a new complaint.

### Get Escalation Statistics

```http
GET /api/escalation/stats
```

**Response:**
```json
{
  "by_level": {
    "1": 45,
    "2": 12,
    "3": 3
  },
  "overdue_count": 8
}
```

## Auto-Escalation Scheduler

The escalation scheduler runs automatically every hour to check for complaints that need escalation.

### Configuration

In `main.go`:

```go
// Initialize escalation scheduler (runs every hour)
escalationScheduler := services.NewEscalationScheduler(database.DB, 1*time.Hour)
escalationScheduler.Start()
```

### Customizing Check Interval

Change the interval by modifying the duration:

```go
// Check every 30 minutes
escalationScheduler := services.NewEscalationScheduler(database.DB, 30*time.Minute)

// Check every 6 hours
escalationScheduler := services.NewEscalationScheduler(database.DB, 6*time.Hour)

// Check once per day
escalationScheduler := services.NewEscalationScheduler(database.DB, 24*time.Hour)
```

### Scheduler Logs

The scheduler logs activity to help monitor escalations:

```
[GIN-debug] Starting complaint escalation scheduler (interval: 1h0m0s)
[GIN-debug] Running scheduled escalation check...
[GIN-debug] Successfully escalated 3 complaint(s)
[GIN-debug] Escalation stats: map[by_level:map[1:45 2:12 3:3] overdue_count:0]
```

## How It Works

### 1. New Complaint Created

When a new complaint is reported:

1. Complaint is saved to `report_issues` table
2. Escalation is initialized with Level 1 team
3. Timer starts: 5 days until next escalation
4. Record created in `complaint_escalations` table
5. History entry logged in `complaint_escalation_history`

### 2. Automatic Escalation

Every hour, the scheduler:

1. Queries `complaint_escalations` for overdue complaints
2. Filters out resolved/closed complaints
3. For each overdue complaint:
   - Moves to next escalation level
   - Updates team assignment
   - Sets new escalation due date
   - Logs history entry
   - (TODO: Sends notifications)

### 3. Manual Resolution

When admin resolves a complaint:

1. Update complaint status to "resolved" in `report_issues`
2. Set `resolved_at` timestamp
3. Set `resolved_by_id` to admin user ID
4. Escalation stops automatically (scheduler skips resolved complaints)

## Frontend Integration

### Display Escalation Info on Complaint Card

Update `complaint-management.component.ts`:

```typescript
interface ComplaintWithEscalation extends Complaint {
  escalation?: {
    current_level: number;
    current_team: string;
    assigned_to_admin_id?: string;
    assigned_to_name?: string;
    last_escalated_at?: string;
    next_escalation_due?: string;
    days_until_escalation?: number;
    escalation_history?: Array<{
      level: number;
      team_name: string;
      escalated_at: string;
      escalated_by: string;
      reason: string;
    }>;
  };
}
```

### Fetch Escalation Data

```typescript
loadComplaintWithEscalation(complaintId: string) {
  forkJoin({
    complaint: this.complaintService.getComplaintById(complaintId),
    escalation: this.http.get(`${environment.apiUrl}/escalation/complaint/${complaintId}`)
  }).subscribe(result => {
    const complaint = {
      ...result.complaint,
      escalation: result.escalation
    };
    // Use complaint with escalation data
  });
}
```

### Display Escalation Badge

```html
<div class="escalation-badge" [class.level-1]="complaint.escalation?.current_level === 1"
                               [class.level-2]="complaint.escalation?.current_level === 2"
                               [class.level-3]="complaint.escalation?.current_level === 3">
  Level {{ complaint.escalation?.current_level }} - {{ complaint.escalation?.current_team }}
</div>
```

### Show Days Until Escalation

```html
<div class="escalation-timer" *ngIf="complaint.escalation?.next_escalation_due">
  <i class="pi pi-clock"></i>
  Escalates in {{ getDaysUntilEscalation(complaint.escalation.next_escalation_due) }} days
</div>
```

### Manual Escalation Button

```html
<button pButton label="Escalate Now" 
        icon="pi pi-arrow-up"
        (click)="manualEscalate(complaint.id)"></button>
```

```typescript
manualEscalate(complaintId: string) {
  this.http.post(`${environment.apiUrl}/escalation/complaint/${complaintId}/escalate`, {
    category: complaint.category,
    current_level: complaint.escalation.current_level,
    escalated_by: this.currentAdminId
  }).subscribe(() => {
    this.messageService.add({
      severity: 'success',
      summary: 'Escalated',
      detail: 'Complaint escalated to next level'
    });
    this.loadComplaints();
  });
}
```

## Testing

### 1. Test Escalation Initialization

```bash
# Create a test complaint
curl -X POST http://localhost:8083/api/complaints \
  -H "Content-Type: application/json" \
  -d '{
    "category": "bus_delay",
    "description": "Test complaint",
    "reported_by_id": "user-uuid"
  }'

# Initialize escalation
curl -X POST http://localhost:8083/api/escalation/complaint/{complaint-id}/initialize \
  -H "Content-Type: application/json" \
  -d '{"category": "bus_delay"}'

# Check escalation status
curl http://localhost:8083/api/escalation/complaint/{complaint-id}
```

### 2. Test Manual Escalation

```bash
curl -X POST http://localhost:8083/api/escalation/complaint/{complaint-id}/escalate \
  -H "Content-Type: application/json" \
  -d '{
    "category": "bus_delay",
    "current_level": 1,
    "escalated_by": "admin-uuid"
  }'
```

### 3. Test Automatic Escalation

```sql
-- Manually set next_escalation_due to past date to trigger escalation
UPDATE complaint_escalations
SET next_escalation_due = NOW() - INTERVAL '1 day'
WHERE complaint_id = 'your-complaint-id';

-- Wait for scheduler to run (or restart server to trigger immediate check)
-- Check logs for "Successfully escalated X complaint(s)"
```

### 4. View Statistics

```bash
curl http://localhost:8083/api/escalation/stats
```

## Customization

### Modify Escalation Days

Edit [escalation.go](backend/internal/models/escalation.go):

```go
{
    Level:          1,
    TeamName:       "Operations Support Team",
    AssignedTo:     []string{},
    EscalationDays: 3,  // Changed from 5 to 3 days
    NotificationEmails: []string{"operations@aasl.lk"},
},
```

### Add New Category

Add to `GetEscalationConfigs()` function in [escalation.go](backend/internal/models/escalation.go):

```go
"new_category": {
    Category: "new_category",
    Levels: []EscalationLevel{
        {
            Level:          1,
            TeamName:       "First Response Team",
            EscalationDays: 5,
            NotificationEmails: []string{"team@company.com"},
        },
        // Add more levels...
    },
},
```

### Change Scheduler Interval

In [main.go](backend/cmd/server/main.go):

```go
// Run every 30 minutes instead of every hour
escalationScheduler := services.NewEscalationScheduler(database.DB, 30*time.Minute)
```

## Troubleshooting

### Escalation Not Working

1. **Check database tables exist**:
   ```sql
   SELECT * FROM complaint_escalations LIMIT 1;
   SELECT * FROM complaint_escalation_history LIMIT 1;
   ```

2. **Check scheduler is running**:
   ```
   Look for "Starting complaint escalation scheduler" in server logs
   ```

3. **Check for overdue complaints**:
   ```sql
   SELECT * FROM complaint_escalations 
   WHERE next_escalation_due <= NOW();
   ```

4. **Manually trigger escalation check**:
   ```go
   // In code or via debug endpoint
   escalationScheduler.RunEscalationCheckNow()
   ```

### Complaints Not Escalating

- Verify complaint status is NOT "resolved" or "closed"
- Check `next_escalation_due` is in the past
- Review server logs for error messages
- Ensure database connection is active

## Future Enhancements

- [ ] Email notifications when complaints escalate
- [ ] SMS notifications for urgent escalations
- [ ] Slack/Teams integration for team notifications
- [ ] Dashboard widget showing escalation metrics
- [ ] Custom escalation rules per complaint
- [ ] SLA (Service Level Agreement) tracking
- [ ] Escalation pause/freeze functionality
- [ ] Bulk assignment of complaints
- [ ] Advanced filtering by escalation level
- [ ] Escalation calendar view

## License

Part of AASL STS Admin Web Application
