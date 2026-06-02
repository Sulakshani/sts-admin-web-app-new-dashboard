# Complaint Escalation System - Complete Implementation Guide

**Created**: March 30, 2026  
**Status**: Ready for Implementation  
**Database**: PostgreSQL  
**Framework**: Go + Gin (Backend), Angular (Frontend)

---

## 📁 Documentation Files Created

All files are located in `backend/` directory:

1. **ADMIN_DATABASE_SCHEMA.md**
   - Admin users table structure
   - Role hierarchy and rules
   - SQL commands for setup
   - API integration points
   - Bootstrap process

2. **COMPLAINT_ESCALATION_ADMIN_SYSTEM.md**
   - Complete escalation system design
   - All 5 database tables (admin_users, complaint_assignments, complaint_escalations, complaint_escalation_history, escalation_sla_config)
   - Role hierarchy and permissions
   - SQL commands with samples
   - Query examples for each role
   - View definitions
   - Triggers and functions

3. **COMPLAINT_ESCALATION_AUTH_INTEGRATION.md**
   - JWT token enhancement
   - Authentication middleware
   - API endpoints with examples
   - Database setup instructions
   - Backend implementation checklist
   - Scheduler code
   - Testing checklist

4. **complaint_escalation_complete_setup.sql**
   - Complete SQL setup script
   - All table definitions
   - Default SLA configurations (48 entries for all apps/levels/priorities)
   - Indexes for performance
   - Triggers for timestamp management
   - Views for dashboards
   - Helper functions

5. **admin_users_setup.sql** (Previously created)
   - Admin users table setup
   - Hierarchy validation triggers
   - Indexes and functions

---

## 🗄️ Database Tables Summary

### 1. admin_users (Updated)
**Purpose**: Authentication and role management

**Key Columns**:
```
id UUID PRIMARY KEY
email TEXT UNIQUE
password_hash TEXT (bcrypt)
full_name TEXT
role TEXT (super_admin | supervisor | admin)
app_scope TEXT (bus | driver | lounges | passenger)
supervisor_id UUID (FK → admin_users)
function_permissions JSONB (["complaints.read", ...])
is_active BOOLEAN (for access control)
```

**Role Hierarchy**:
```
super_admin
├── Can create/manage all admins
├── Can view ALL complaints globally
├── app_scope = NULL
└── supervisor_id = NULL

supervisor (app-scoped)
├── Can manage app-specific admins
├── Can handle escalated complaints
├── app_scope = REQUIRED (bus|driver|lounges|passenger)
└── supervisor_id = OPTIONAL

admin (app-scoped)
├── Handles assigned complaints
├── Cannot escalate directly
├── app_scope = REQUIRED
└── supervisor_id = REQUIRED (parent supervisor)
```

### 2. complaint_assignments (New)
**Purpose**: Track current complaint assignments and SLA

**Key Columns**:
```
id UUID PRIMARY KEY
complaint_id UUID UNIQUE (FK → report_issues)
assigned_to_admin_id UUID (FK → admin_users)
source_app TEXT (bus|driver|lounges|passenger)
escalation_level INTEGER (1|2|3)
resolution_deadline TIMESTAMP
escalation_deadline TIMESTAMP
is_escalated BOOLEAN
previous_admin_id UUID (for tracking failed assignments)
status TEXT (assigned|in_progress|resolved|escalated|closed)
```

### 3. complaint_escalations (Updated)
**Purpose**: Current escalation state

**Key Columns**:
```
complaint_id UUID PRIMARY KEY (FK → report_issues)
current_level INTEGER (1|2|3)
current_admin_id UUID (FK → admin_users)
source_app TEXT
next_escalation_due TIMESTAMP
assigned_to_admin_id UUID
previous_assigned_admin_id UUID
```

### 4. complaint_escalation_history (Updated)
**Purpose**: Audit trail of all escalations

**Key Columns**:
```
id UUID PRIMARY KEY
complaint_id UUID (FK → report_issues)
escalation_level INTEGER
from_admin_id UUID
to_admin_id UUID
escalation_reason TEXT
escalation_type TEXT (sla_timeout|manual_escalation|priority_increase|reassignment|resolution)
escalated_at TIMESTAMP
```

### 5. escalation_sla_config (New)
**Purpose**: Configurable SLA times

**Key Columns**:
```
id UUID PRIMARY KEY
source_app TEXT (bus|driver|lounges|passenger)
escalation_level INTEGER (1|2|3)
issue_priority TEXT (low|medium|high|emergency)
escalation_threshold_hours INTEGER
final_resolution_hours INTEGER
auto_escalate_on_timeout BOOLEAN
notify_supervisor BOOLEAN
revoke_admin_access_on_escalation BOOLEAN
```

**Default SLA Example (Bus App)**:
```
Level 1 → Low Priority: 48 hours escalation, 96 hours final
Level 1 → Emergency: 2 hours escalation, 6 hours final
Level 2 → Emergency: 1 hour escalation, 3 hours final
Level 3 (Super Admin): Immediate handling
```

---

## 🔄 Complaint Escalation Workflow

```
┌─ Complaint Created ─┐
│                     │
v                     
┌─────────────────────────────────┐
│ 1. Auto-Assignment              │
│    Find admin WHERE             │
│    role='admin'                 │
│    app_scope=complaint.app_scope│
│    is_active=true               │
└─────────────────────────────────┘
           │
           ├─ Found → Assign to admin (Level 1)
           └─ Not Found → Assign to supervisor (Level 2)
                    │
                    └─ Not Found → Assign to super_admin (Level 3)
           
┌─────────────────────────────────┐
│ 2. SLA Calculation              │
│    resolution_deadline = now +  │
│    final_resolution_hours       │
│    escalation_deadline = now +  │
│    escalation_threshold_hours   │
└─────────────────────────────────┘
           │
           v
┌─────────────────────────────────┐
│ 3. Admin Views Complaint        │
│    Login → See dashboard        │
│    View assigned complaints     │
│    Time remaining shown         │
└─────────────────────────────────┘
           │
   ┌───────┴────────┐
   │ Work on it    │ Ignore it
   │ (Update to    │ (No changes)
   │ in_progress)  │
   │               │
   v               v
┌─────────────────────────────────┐
│ 4. Check Auto-Escalation (Hourly)
│    Is escalation_deadline < now?
│    Is status = assigned/in_progress?
└─────────────────────────────────┘
           │
   ┌───────┴────────────────────┐
   │ YES (Escalate)        │ NO (Continue)
   │                           │ Keep assigned
   │                           │ to admin
   v
┌─────────────────────────────────┐
│ 5. Escalation Actions           │
│    1. Update escalation_level++ │
│    2. Reassign to supervisor    │
│    3. Set is_escalated=true     │
│    4. Log escalation_history    │
│    5. Notify supervisor         │
│    6. Deactivate unresolved     │
│       admin (optional)          │
└─────────────────────────────────┘
           │
           v
┌─────────────────────────────────┐
│ 6. Supervisor Handles Complaint │
│    Can reassign or resolve      │
│    Can further escalate to      │
│    super admin if needed        │
└─────────────────────────────────┘
           │
           v
┌─────────────────────────────────┐
│ 7. Resolution                   │
│    Mark status = resolved       │
│    Log resolution in history    │
│    Close complaint              │
└─────────────────────────────────┘
```

---

## 🔐 Access Control Rules

### Admin Can:
- View assigned complaints
- Update complaint status
- Resolve complaint
- See time remaining

### Supervisor Can:
- View ALL escalated complaints in their app scope
- View issues from their team admins
- Reassign complaints
- Manually escalate to super admin
- Resolve complaints

### Super Admin Can:
- View ALL complaints globally
- View system-wide escalation status
- See admin workload
- Create/manage admins
- Modify SLA configurations
- Force escalation

### After Escalation:
- **Previous Admin**: Loses access to that specific complaint
- **New Admin**: Can now view and manage
- **History**: All transitions logged in `complaint_escalation_history`

---

## 📊 Key Queries

### 1. Get Admin's Assigned Complaints
```sql
SELECT ca.*, ri.issue_type, ri.priority
FROM complaint_assignments ca
JOIN report_issues ri ON ca.complaint_id = ri.id
WHERE ca.assigned_to_admin_id = $1
AND ca.status IN ('assigned', 'in_progress')
ORDER BY ca.escalation_deadline ASC;
```

### 2. Get Supervisor's Escalated Complaints
```sql
SELECT ca.*, ri.issue_type, ri.priority
FROM complaint_assignments ca
JOIN report_issues ri ON ca.complaint_id = ri.id
WHERE ca.assigned_to_admin_id = $1
AND ca.escalation_level = 2
AND ca.status IN ('assigned', 'in_progress', 'escalated')
ORDER BY ca.resolution_deadline ASC;
```

### 3. Find Complaints Ready for Auto-Escalation
```sql
SELECT ca.id, ca.assigned_to_admin_id
FROM complaint_assignments ca
WHERE ca.is_escalated = false
AND ca.escalation_deadline < now()
AND ca.status IN ('assigned', 'in_progress')
ORDER BY ca.escalation_deadline ASC;
```

### 4. Get Admin Workload Overview
```sql
SELECT 
    au.full_name,
    COUNT(*) FILTER (WHERE status='assigned') as assigned,
    COUNT(*) FILTER (WHERE status='in_progress') as in_progress,
    COUNT(*) FILTER (WHERE status='resolved') as resolved,
    MIN(escalation_deadline) as next_deadline
FROM admin_users au
LEFT JOIN complaint_assignments ca ON au.id = ca.assigned_to_admin_id
WHERE au.is_active = true
GROUP BY au.id, au.full_name;
```

### 5. Get Escalation History for Complaint
```sql
SELECT 
    ceh.*,
    from_admin.full_name as from_admin_name,
    to_admin.full_name as to_admin_name
FROM complaint_escalation_history ceh
LEFT JOIN admin_users from_admin ON ceh.from_admin_id = from_admin.id
LEFT JOIN admin_users to_admin ON ceh.to_admin_id = to_admin.id
WHERE ceh.complaint_id = $1
ORDER BY ceh.escalated_at DESC;
```

---

## 🔧 Implementation Order

### Phase 1: Database Setup (Week 1)
1. Run `complaint_escalation_complete_setup.sql`
2. Verify all tables created
3. Check indexes present
4. Test sample queries

### Phase 2: Backend API (Week 1-2)
1. Update `AdminLogin` to include role/app_scope in JWT
2. Create middleware: `CheckComplaintAccess`
3. Update handlers:
   - `GetAdminComplaints` (role-based filtering)
   - `GetEscalatedComplaints` (supervisors)
   - `GetAllComplaints` (super admin)
   - `UpdateComplaintStatus`
   - `ResolveComplaint`
   - `ManuallyEscalateComplaint`
4. Add scheduler: `auto_escalate_overdue_complaints()`

### Phase 3: Frontend Updates (Week 2)
1. Update admin dashboard
2. Add complaint assignment view
3. Add escalation notification
4. Add resolution form
5. Add super admin monitoring dashboard

### Phase 4: Testing (Week 3)
1. Unit tests for escalation logic
2. Integration tests for workflows
3. Load tests for scheduler
4. UAT with stakeholders

### Phase 5: Deployment (Week 4)
1. Database migration on production
2. Backend deployment
3. Frontend deployment
4. Monitoring setup

---

## 📈 Performance Considerations

### Indexes Created
```
admin_users(email, role, app_scope, is_active, supervisor_id, created_at)
complaint_assignments(admin, deadline, escalation_deadline, status, source_app, complaint)
complaint_escalations(next_due, level, admin, active, source_app)
complaint_escalation_history(complaint, admin, type, level, created)
```

### Query Optimization
- Use indexes on foreign keys
- Filter by status to reduce dataset
- Limit results with pagination
- Cache dashboard views

### Scheduler Optimization
- Run every 1 hour (configurable)
- Batch update escalations
- Use database functions for efficiency
- Log only critical events

---

## 🚨 Monitoring & Alerts

### Metrics to Monitor
```
1. Average complaint resolution time per admin
2. Escalation rate per app scope
3. Admin response time
4. SLA compliance rate
5. Supervisor intervention rate
```

### Alerts to Setup
```
1. Complaint approaching escalation deadline (2 hours warning)
2. Escalation due but not processed (error)
3. Admin offline for escalated complaint
4. Multiple escalations for same complaint
5. SLA breach for critical complaints
```

---

## 📝 Notes for Implementation

### Important
- **Bcrypt Password Hashing**: Use cost=10 for balance of security/performance
- **JWT Expiry**: Set to 24 hours, use refresh token for longer sessions
- **Timezone**: All timestamps use UTC for consistency
- **UUID Generation**: PostgreSQL handles uuid_generate_v4()

### Optional But Recommended
- Add SMS/Email notifications on escalation
- Add audit logging to database for compliance
- Add Slack integration for notifications
- Add analytics dashboard for management

### Testing Tips
- Create test admin for each role
- Create test complaints with different priorities
- Manually trigger escalation (don't wait 1 hour)
- Test with various time zones
- Load test scheduler with 1000+ complaints

---

## 📞 Support & References

### PostgreSQL Docs
- UUID Functions: https://www.postgresql.org/docs/current/uuid-ossp.html
- JSON Functions: https://www.postgresql.org/docs/current/functions-json.html
- Triggers: https://www.postgresql.org/docs/current/sql-createtrigger.html

### Go Packages
- JWT-Go: https://github.com/golang-jwt/jwt
- Gin: https://github.com/gin-gonic/gin
- PostgreSQL Driver: https://github.com/lib/pq

---

## ✅ Verification Checklist

After implementation, verify:

- [ ] All 5 tables created and accessible
- [ ] All indexes present and active
- [ ] Default SLA configurations loaded (48 entries)
- [ ] Triggers firing on updates
- [ ] Views returning correct data
- [ ] JWT token includes role, app_scope, supervisor_id
- [ ] Admin can view only assigned complaints
- [ ] Supervisor can view escalated complaints
- [ ] Super admin can view all complaints
- [ ] Auto-escalation works hourly
- [ ] Escalation history logged correctly
- [ ] Access removed from previous admin after escalation
- [ ] Notifications sending to admins
- [ ] Dashboard displaying urgency levels correctly
- [ ] All API endpoints secured with middleware

---

## 🎯 Success Criteria

The implementation is successful when:

1. ✅ Admins receive complaints automatically based on app scope
2. ✅ Complaints escalate to supervisor after SLA timeout (no manual intervention)
3. ✅ Unresolved admins lose access to escalated complaints
4. ✅ Supervisors can see and manage escalated complaints
5. ✅ Super admin has global visibility of all complaints
6. ✅ Complete audit trail of all escalations maintained
7. ✅ All roles see only authorized complaints
8. ✅ System performs well with 1000+ concurrent complaints
9. ✅ No data loss or corruption during escalations
10. ✅ Stakeholders satisfied with workflow

---

