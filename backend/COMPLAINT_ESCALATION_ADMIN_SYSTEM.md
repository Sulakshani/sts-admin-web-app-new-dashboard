# Complaint Escalation System with Admin Role-Based Assignment

## Overview
Complete implementation of complaint escalation system where:
- Super Admin creates Admins for each app scope (bus, driver, lounges, passenger)
- Each Admin is assigned to a Supervisor
- Complaints are automatically assigned to the relevant app admin
- Auto-escalation to supervisor if admin doesn't resolve within SLA period
- Access removal from unresolved admin after escalation to supervisor
- Super Admin has visibility of all complaints and escalation statuses

---

## Database Tables

### 1. **admin_users** (Updated)

#### Purpose
Stores admin users with role hierarchy and app-scope assignment for complaint management.

#### Updated Table SQL

```sql
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    
    -- Role Hierarchy
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('super_admin', 'supervisor', 'admin')),
    
    -- App Scope for assignment
    app_scope TEXT CHECK (app_scope IN ('bus', 'driver', 'lounges', 'passenger') OR app_scope IS NULL),
    
    -- Supervisor relationship
    supervisor_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    
    -- Permissions
    function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    
    -- Status tracking
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit columns
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL
);
```

#### Key Additions for Complaint Escalation
- `app_scope`: Determines which category of complaints this admin handles
- `supervisor_id`: Used to escalate unresolved complaints
- `is_active`: Allows deactivating admins (access revocation after escalation)

#### Column Definitions

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID | Unique admin identifier |
| `email` | TEXT | Login email |
| `password_hash` | TEXT | Bcrypt hash |
| `full_name` | TEXT | Admin name |
| `role` | TEXT | super_admin \| supervisor \| admin |
| `app_scope` | TEXT | bus \| driver \| lounges \| passenger (null for super_admin) |
| `supervisor_id` | UUID FK | Parent supervisor (required for admin, optional for supervisor) |
| `function_permissions` | JSONB | ["complaints.read", "complaints.assign", "\*"] |
| `is_active` | BOOLEAN | Active/Inactive status for access control |
| `created_by` | UUID FK | Super admin who created this admin |

---

### 2. **complaint_assignments** (New)

#### Purpose
Tracks current assignment of complaints to admins, replacement admin, and time windows for resolution.

#### Create Table SQL

```sql
CREATE TABLE IF NOT EXISTS complaint_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    complaint_id UUID NOT NULL UNIQUE REFERENCES report_issues(id) ON DELETE CASCADE,
    
    -- Current Assignment
    assigned_to_admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE RESTRICT,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    assignment_type TEXT NOT NULL DEFAULT 'initial' CHECK (assignment_type IN ('initial', 'reassigned', 'escalated')),
    
    -- SLA & Deadline Tracking
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    sla_hours INTEGER NOT NULL DEFAULT 24,
    resolution_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    escalation_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Escalation Tracking
    escalation_level INTEGER NOT NULL DEFAULT 1 CHECK (escalation_level IN (1, 2, 3)),
    is_escalated BOOLEAN DEFAULT false,
    escalation_reason TEXT,
    escalated_at TIMESTAMP WITH TIME ZONE,
    
    -- Previous Assignment (for tracking who failed to resolve)
    previous_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    previous_assignment_reason TEXT,
    
    -- Status
    status TEXT NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'in_progress', 'resolved', 'escalated', 'closed')),
    
    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    resolved_by_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_admin ON complaint_assignments(assigned_to_admin_id);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_deadline ON complaint_assignments(resolution_deadline);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_escalation ON complaint_assignments(escalation_deadline) WHERE is_escalated = false;
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_status ON complaint_assignments(status);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_source_app ON complaint_assignments(source_app);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_complaint ON complaint_assignments(complaint_id);
```

#### Column Definitions

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID | Unique assignment record |
| `complaint_id` | UUID FK | Reference to report_issues |
| `assigned_to_admin_id` | UUID FK | Current admin handling complaint |
| `assigned_at` | TIMESTAMP | When assignment was made |
| `assignment_type` | TEXT | initial \| reassigned \| escalated |
| `source_app` | TEXT | bus \| driver \| lounges \| passenger |
| `sla_hours` | INTEGER | Hours allowed to resolve (default 24) |
| `resolution_deadline` | TIMESTAMP | Final deadline |
| `escalation_deadline` | TIMESTAMP | When to escalate to supervisor |
| `escalation_level` | INTEGER | 1=admin, 2=supervisor, 3=super_admin |
| `is_escalated` | BOOLEAN | Has this been escalated? |
| `escalation_reason` | TEXT | Why it was escalated |
| `previous_admin_id` | UUID FK | Admin who failed to resolve |
| `status` | TEXT | assigned \| in_progress \| resolved \| escalated \| closed |
| `resolved_by_admin_id` | UUID FK | Admin who resolved it |

---

### 3. **complaint_escalations** (Updated)

#### Purpose
Tracks current escalation state with assignment history.

#### Updated Table SQL

```sql
CREATE TABLE IF NOT EXISTS complaint_escalations (
    complaint_id UUID PRIMARY KEY REFERENCES report_issues(id) ON DELETE CASCADE,
    
    -- Current State
    current_level INTEGER NOT NULL DEFAULT 1 CHECK (current_level IN (1, 2, 3)),
    current_admin_id UUID REFERENCES admin_users(id),
    current_team TEXT NOT NULL,
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    
    -- SLA Tracking
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    next_escalation_due TIMESTAMP WITH TIME ZONE,
    final_resolution_due TIMESTAMP WITH TIME ZONE,
    
    -- Assignment History
    assigned_to_admin_id UUID REFERENCES admin_users(id),
    previous_assigned_admin_id UUID REFERENCES admin_users(id),
    
    -- Status & Audit
    is_active BOOLEAN DEFAULT true,
    last_escalated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now())
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_next_due ON complaint_escalations(next_escalation_due) WHERE next_escalation_due IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_level ON complaint_escalations(current_level);
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_admin ON complaint_escalations(current_admin_id) WHERE current_admin_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_active ON complaint_escalations(is_active) WHERE is_active = true;
```

---

### 4. **complaint_escalation_history** (Updated)

#### Purpose
Audit trail of all escalation events with admin transitions.

#### Updated Table SQL

```sql
CREATE TABLE IF NOT EXISTS complaint_escalation_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    complaint_id UUID NOT NULL REFERENCES report_issues(id) ON DELETE CASCADE,
    
    -- Escalation Details
    escalation_level INTEGER NOT NULL CHECK (escalation_level IN (1, 2, 3)),
    from_level INTEGER,
    to_level INTEGER,
    
    -- Admin Transition
    from_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    to_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    from_team TEXT,
    to_team TEXT,
    
    -- Reason & Details
    escalation_reason TEXT NOT NULL,
    escalation_type TEXT NOT NULL DEFAULT 'sla_timeout' CHECK (escalation_type IN (
        'sla_timeout',      -- Admin didn't resolve within SLA
        'manual_escalation', -- Manually escalated by supervisor
        'priority_increase', -- Priority increased by admin
        'reassignment',      -- Reassigned due to admin inactivity
        'resolution'         -- Resolved and closed
    )),
    
    -- Timestamps
    escalated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    escalated_by_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now())
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_complaint ON complaint_escalation_history(complaint_id);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_admin ON complaint_escalation_history(to_admin_id);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_type ON complaint_escalation_history(escalation_type);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_level ON complaint_escalation_history(escalation_level);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_created ON complaint_escalation_history(created_at DESC);
```

---

### 5. **escalation_sla_config** (New)

#### Purpose
Configurable SLA times for each escalation level and app scope.

#### Create Table SQL

```sql
CREATE TABLE IF NOT EXISTS escalation_sla_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Configuration
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    escalation_level INTEGER NOT NULL CHECK (escalation_level IN (1, 2, 3)),
    issue_priority TEXT NOT NULL DEFAULT 'medium' CHECK (issue_priority IN ('low', 'medium', 'high', 'emergency')),
    
    -- SLA Settings (in hours)
    escalation_threshold_hours INTEGER NOT NULL,
    final_resolution_hours INTEGER NOT NULL,
    
    -- Assignment Rules
    auto_escalate_on_timeout BOOLEAN DEFAULT true,
    notify_supervisor BOOLEAN DEFAULT true,
    revoke_admin_access_on_escalation BOOLEAN DEFAULT true,
    
    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    
    UNIQUE(source_app, escalation_level, issue_priority)
);

-- Sample Data
INSERT INTO escalation_sla_config (source_app, escalation_level, issue_priority, escalation_threshold_hours, final_resolution_hours, auto_escalate_on_timeout, notify_supervisor, revoke_admin_access_on_escalation)
VALUES
    -- Bus app - Low priority
    ('bus', 1, 'low', 48, 96, true, false, false),
    ('bus', 2, 'low', 24, 48, true, true, true),
    ('bus', 3, 'low', 8, 24, true, true, false),
    -- Bus app - High priority
    ('bus', 1, 'high', 8, 24, true, true, true),
    ('bus', 2, 'high', 4, 12, true, true, true),
    ('bus', 3, 'high', 1, 4, true, true, false),
    -- Bus app - Emergency
    ('bus', 1, 'emergency', 2, 6, true, true, true),
    ('bus', 2, 'emergency', 1, 3, true, true, true),
    ('bus', 3, 'emergency', 0.5, 2, true, true, false)
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_escalation_sla_config_app_level ON escalation_sla_config(source_app, escalation_level);
```

---

## Role Hierarchy & Permissions

### Admin Creation Flow

```
super_admin creates admin
         ↓
  admin requires supervisor assignment (supervisor_id)
         ↓
  admin can only handle complaints for their app_scope
         ↓
  if admin doesn't resolve → escalate to supervisor
         ↓
  supervisor can view all complaints in their scope
         ↓
  if supervisor doesn't resolve → escalate to super_admin
         ↓
  super_admin can view ALL complaints globally
```

### Complaint Assignment Rules

```sql
-- When complaint is created (source_app from report_issues):
-- 1. Find admin WHERE role='admin' AND app_scope=complaint.app_scope AND is_active=true
-- 2. If no admin found, assign to supervisor of that app_scope
-- 3. If no supervisor, escalate to super_admin immediately
-- 4. Calculate deadline based on escalation_sla_config
-- 5. Send notification to assigned admin
```

---

## API Integration & SQL Commands

### 1. Create Admin with Supervisor Assignment

```sql
-- Super admin creates an admin for 'bus' app scope
INSERT INTO admin_users (
    id, email, password_hash, full_name, role, 
    app_scope, supervisor_id, function_permissions, 
    is_active, created_by
)
VALUES (
    uuid_generate_v4(),
    'bus-admin@example.com',
    '$2a$10$...',  -- bcrypt hash
    'Bus Admin',
    'admin',
    'bus',  -- app scope
    (SELECT id FROM admin_users WHERE email='bus-supervisor@example.com'),  -- supervisor_id
    '["complaints.read", "complaints.assign", "complaints.update"]'::jsonb,
    true,
    (SELECT id FROM admin_users WHERE role='super_admin' LIMIT 1)
);
```

### 2. Auto-Assign New Complaint to App Admin

```sql
-- When a new complaint is created, trigger assignment
WITH complaint_data AS (
    SELECT id, source_app, priority 
    FROM report_issues 
    WHERE id = $1
),
target_admin AS (
    SELECT id, email, full_name 
    FROM admin_users
    WHERE role = 'admin'
    AND app_scope = (SELECT source_app FROM complaint_data)
    AND is_active = true
    LIMIT 1
)
INSERT INTO complaint_assignments (
    complaint_id,
    assigned_to_admin_id,
    source_app,
    sla_hours,
    resolution_deadline,
    escalation_deadline,
    assignment_type,
    status
)
SELECT 
    c.id,
    t.id,
    c.source_app,
    24,  -- Default SLA
    timezone('utc'::text, now()) + interval '24 hours',
    timezone('utc'::text, now()) + interval '12 hours',
    'initial',
    'assigned'
FROM complaint_data c, target_admin t
WHERE FOUND  -- Only if admin was found
RETURNING *;

-- If no admin found, create event to notify supervisor/super_admin
```

### 3. Check for Auto-Escalation (Run by Scheduler Every Hour)

```sql
-- Find complaints past escalation deadline
SELECT ca.id, ca.complaint_id, ca.assigned_to_admin_id, 
       ca.escalation_deadline, ca.escalation_level
FROM complaint_assignments ca
WHERE ca.is_escalated = false
AND ca.escalation_deadline < timezone('utc'::text, now())
AND ca.status IN ('assigned', 'in_progress')
ORDER BY ca.escalation_deadline ASC;

-- For each complaint, escalate:
UPDATE complaint_assignments
SET 
    is_escalated = true,
    escalation_reason = 'SLA timeout - auto escalated',
    escalated_at = timezone('utc'::text, now()),
    previous_admin_id = assigned_to_admin_id,
    escalation_level = escalation_level + 1,
    assignment_type = 'escalated'
WHERE complaint_id = $1;

-- Find supervisor for escalation
SELECT supervisor_id FROM admin_users 
WHERE id = (SELECT assigned_to_admin_id FROM complaint_assignments WHERE complaint_id = $1);

-- Re-assign to supervisor
UPDATE complaint_assignments
SET assigned_to_admin_id = $2  -- supervisor_id
WHERE complaint_id = $1;

-- Log escalation event
INSERT INTO complaint_escalation_history (
    complaint_id, escalation_level, from_level, to_level,
    from_admin_id, to_admin_id, escalation_reason, 
    escalation_type, escalated_by_admin_id
)
VALUES ($1, 2, 1, 2, $3, $4, 'SLA timeout', 'sla_timeout', NULL);

-- If supervisor_id is required, remove access from previous admin
UPDATE admin_users
SET is_active = false
WHERE id = (SELECT previous_admin_id FROM complaint_assignments WHERE complaint_id = $1)
AND role = 'admin'  -- Only deactivate if they're just an admin
AND (SELECT COUNT(*) FROM complaint_assignments 
     WHERE assigned_to_admin_id = admin_users.id 
     AND status != 'resolved') = 1;  -- If this is their only complaint
```

### 4. Admin Views Assigned Complaints

```sql
-- Get all complaints assigned to logged-in admin
SELECT 
    ca.complaint_id,
    ca.assigned_to_admin_id,
    ca.resolution_deadline,
    ca.escalation_deadline,
    ca.escalation_level,
    ca.source_app,
    ca.status,
    ri.issue_type,
    ri.priority,
    ri.description,
    ri.status as complaint_status,
    EXTRACT(EPOCH FROM (ca.resolution_deadline - timezone('utc'::text, now()))) / 3600 as hours_remaining
FROM complaint_assignments ca
JOIN report_issues ri ON ca.complaint_id = ri.id
WHERE ca.assigned_to_admin_id = $1  -- Logged-in admin
AND ca.status IN ('assigned', 'in_progress')
ORDER BY ca.escalation_deadline ASC;
```

### 5. Supervisor Views Escalated Complaints

```sql
-- Get all complaints escalated to supervisor
SELECT 
    ca.complaint_id,
    ca.assigned_to_admin_id,
    ca.previous_admin_id,
    ca.resolution_deadline,
    ca.escalation_level,
    ca.source_app,
    ca.status,
    ri.issue_type,
    ri.priority,
    ca.escalation_reason,
    ca.escalated_at
FROM complaint_assignments ca
JOIN report_issues ri ON ca.complaint_id = ri.id
JOIN admin_users supervisor ON ca.assigned_to_admin_id = supervisor.id
WHERE supervisor.id = $1  -- Logged-in supervisor
AND ca.escalation_level = 2
AND ca.status IN ('assigned', 'in_progress', 'escalated')
ORDER BY ca.resolution_deadline ASC;
```

### 6. Super Admin Views All Complaints with Status

```sql
-- Dashboard showing all complaints with current assignment
SELECT 
    ca.complaint_id,
    to_json(admin_current.*) as assigned_admin,
    to_json(admin_previous.*) as previous_admin,
    ca.escalation_level,
    ca.source_app,
    ca.status,
    ca.is_escalated,
    ca.escalation_reason,
    ri.issue_type,
    ri.priority,
    ri.status as complaint_status,
    ca.resolution_deadline,
    CASE 
        WHEN ca.resolution_deadline < timezone('utc'::text, now()) THEN 'overdue'
        WHEN ca.escalation_deadline < timezone('utc'::text, now()) THEN 'escalation_due'
        ELSE 'on_track'
    END as urgency_status
FROM complaint_assignments ca
LEFT JOIN report_issues ri ON ca.complaint_id = ri.id
LEFT JOIN admin_users admin_current ON ca.assigned_to_admin_id = admin_current.id
LEFT JOIN admin_users admin_previous ON ca.previous_admin_id = admin_previous.id
WHERE (SELECT role FROM admin_users WHERE id = $1) = 'super_admin'
ORDER BY ca.resolution_deadline ASC;
```

### 7. Mark Complaint as Resolved

```sql
-- Admin marks complaint as resolved
UPDATE complaint_assignments
SET 
    status = 'resolved',
    resolved_by_admin_id = $1,  -- Admin ID
    resolved_at = timezone('utc'::text, now()),
    updated_at = timezone('utc'::text, now())
WHERE complaint_id = $2
AND assigned_to_admin_id = $1;

-- Log resolution
INSERT INTO complaint_escalation_history (
    complaint_id,
    escalation_level,
    from_level,
    to_level,
    from_admin_id,
    escalation_reason,
    escalation_type,
    escalated_by_admin_id
)
VALUES (
    $2,
    (SELECT escalation_level FROM complaint_assignments WHERE complaint_id = $2),
    (SELECT escalation_level FROM complaint_assignments WHERE complaint_id = $2),
    (SELECT escalation_level FROM complaint_assignments WHERE complaint_id = $2),
    $1,
    'Complaint resolved by admin',
    'resolution',
    $1
);

-- Update main table
UPDATE report_issues
SET status = 'resolved'
WHERE id = $2;
```

---

## Middleware & Authentication Integration

### Check Complaint Access Permission

```sql
-- In backend handler before returning complaint details:
-- 1. Get admin role and app_scope from JWT token
-- 2. Query current assignment for complaint

SELECT ca.complaint_id
FROM complaint_assignments ca
WHERE ca.complaint_id = $1
AND (
    -- Super admin can access all
    (SELECT role FROM admin_users WHERE id = $2) = 'super_admin'
    OR
    -- Supervisor can access complaints in their assignments
    (ca.assigned_to_admin_id = $2 AND ca.escalation_level = 2)
    OR
    -- Admin can access assigned complaints
    (ca.assigned_to_admin_id = $2 AND ca.escalation_level = 1)
    OR
    -- Supervisor can access complaints assigned to their team
    (ca.assigned_to_admin_id IN (
        SELECT id FROM admin_users 
        WHERE supervisor_id = $2 AND role = 'admin'
    ))
);

-- If no rows returned, return 403 Forbidden
```

---

## View for Complaint Dashboard

```sql
-- Create comprehensive view for dashboards
CREATE OR REPLACE VIEW v_complaint_escalation_dashboard AS
SELECT 
    ca.complaint_id,
    ca.assigned_to_admin_id,
    admin_current.full_name as assigned_admin_name,
    admin_current.email as assigned_admin_email,
    admin_current.app_scope,
    supervisor.full_name as supervisor_name,
    ca.escalation_level,
    ca.source_app,
    ca.status,
    ca.is_escalated,
    ca.escalation_reason,
    ri.issue_type,
    ri.priority,
    ri.description,
    ri.status as complaint_status,
    ca.assigned_at,
    ca.resolution_deadline,
    ca.escalation_deadline,
    EXTRACT(EPOCH FROM (ca.resolution_deadline - timezone('utc'::text, now()))) / 3600 as hours_until_deadline,
    CASE 
        WHEN ca.resolution_deadline < timezone('utc'::text, now()) THEN 'OVERDUE'
        WHEN ca.escalation_deadline < timezone('utc'::text, now()) THEN 'ESCALATION_DUE'
        WHEN ca.escalation_deadline < timezone('utc'::text, now()) + interval '4 hours' THEN 'CRITICAL'
        ELSE 'ON_TRACK'
    END as urgency_level,
    ceh.escalation_count,
    ceh.last_escalation_at
FROM complaint_assignments ca
LEFT JOIN report_issues ri ON ca.complaint_id = ri.id
LEFT JOIN admin_users admin_current ON ca.assigned_to_admin_id = admin_current.id
LEFT JOIN admin_users supervisor ON admin_current.supervisor_id = supervisor.id
LEFT JOIN (
    SELECT complaint_id, COUNT(*) as escalation_count, MAX(escalated_at) as last_escalation_at
    FROM complaint_escalation_history
    GROUP BY complaint_id
) ceh ON ca.complaint_id = ceh.complaint_id;
```

---

## Database Triggers (Optional But Recommended)

### 1. Auto-Update updated_at on Complaint

```sql
-- Create trigger to update timestamps
CREATE OR REPLACE FUNCTION update_complaint_assignments_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = timezone('utc'::text, now());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_complaint_assignments_updated_at ON complaint_assignments;
CREATE TRIGGER trigger_complaint_assignments_updated_at
BEFORE UPDATE ON complaint_assignments
FOR EACH ROW
EXECUTE FUNCTION update_complaint_assignments_timestamp();
```

### 2. Log Escalation on Update

```sql
CREATE OR REPLACE FUNCTION log_escalation_on_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_escalated != OLD.is_escalated AND NEW.is_escalated = true THEN
        INSERT INTO complaint_escalation_history (
            complaint_id,
            escalation_level,
            from_level,
            to_level,
            from_admin_id,
            to_admin_id,
            escalation_reason,
            escalation_type
        ) VALUES (
            NEW.complaint_id,
            NEW.escalation_level,
            NEW.escalation_level - 1,
            NEW.escalation_level,
            OLD.assigned_to_admin_id,
            NEW.assigned_to_admin_id,
            NEW.escalation_reason,
            'sla_timeout'
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_log_escalation ON complaint_assignments;
CREATE TRIGGER trigger_log_escalation
AFTER UPDATE ON complaint_assignments
FOR EACH ROW
EXECUTE FUNCTION log_escalation_on_change();
```

---

## Implementation Checklist

- [ ] Create `escalation_sla_config` table with default values
- [ ] Create/Update `complaint_assignments` table
- [ ] Update `complaint_escalations` table structure
- [ ] Update `complaint_escalation_history` table
- [ ] Create indexes on all foreign keys and date fields
- [ ] Create triggers for auto-update timestamps
- [ ] Create backend scheduler to check escalations every hour
- [ ] Add permission checks in complaint endpoints
- [ ] Add complaint assignment logic on complaint creation
- [ ] Add escalation logic in scheduler
- [ ] Create dashboard query/endpoint for each role
- [ ] Update authentication middleware to include app_scope
- [ ] Add audit logging for all escalation events
- [ ] Test role hierarchy enforcement
- [ ] Test auto-escalation workflow

---

## Key Features Summary

✓ **Admin Creation**: Super admin creates app-specific admins with supervisor assignment
✓ **Auto Assignment**: New complaints assigned to relevant app admin automatically
✓ **SLA Tracking**: Configurable resolution deadlines per priority and app
✓ **Auto Escalation**: Complaints escalate to supervisor if admin misses deadline
✓ **Access Control**: Unresolved admins lose access post-escalation
✓ **Audit Trail**: Complete history of all escalations and transitions
✓ **Role-Based Views**: Each role sees only authorized complaints
✓ **Super Admin Oversight**: Global visibility of all complaint statuses
✓ **Supervisor Management**: Can reassign and handle escalated complaints

