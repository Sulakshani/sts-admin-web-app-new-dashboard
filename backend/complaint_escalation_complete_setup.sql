-- =====================================================================
-- Complaint Escalation System with Admin Role-Based Assignment
-- Complete SQL Setup Script
-- =====================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================================
-- 1. UPDATE ADMIN_USERS TABLE (if not already created)
-- =====================================================================

-- Create or update admin_users table if it doesn't exist
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('super_admin', 'supervisor', 'admin')),
    app_scope TEXT CHECK (app_scope IN ('bus', 'driver', 'lounges', 'passenger') OR app_scope IS NULL),
    supervisor_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Add missing columns to existing admin_users table
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL;

-- Create indexes on admin_users
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_admin_users_app_scope ON admin_users(app_scope);
CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);
CREATE INDEX IF NOT EXISTS idx_admin_users_supervisor_id ON admin_users(supervisor_id);
CREATE INDEX IF NOT EXISTS idx_admin_users_created_at ON admin_users(created_at DESC);

-- =====================================================================
-- 2. CREATE COMPLAINT_ASSIGNMENTS TABLE (New)
-- =====================================================================

CREATE TABLE IF NOT EXISTS complaint_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    complaint_id UUID NOT NULL UNIQUE REFERENCES report_issues(id) ON DELETE CASCADE,
    assigned_to_admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE RESTRICT,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    assignment_type TEXT NOT NULL DEFAULT 'initial' CHECK (assignment_type IN ('initial', 'reassigned', 'escalated')),
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    sla_hours INTEGER NOT NULL DEFAULT 24,
    resolution_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    escalation_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    escalation_level INTEGER NOT NULL DEFAULT 1 CHECK (escalation_level IN (1, 2, 3)),
    is_escalated BOOLEAN DEFAULT false,
    escalation_reason TEXT,
    escalated_at TIMESTAMP WITH TIME ZONE,
    previous_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    previous_assignment_reason TEXT,
    status TEXT NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'in_progress', 'resolved', 'escalated', 'closed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    resolved_by_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for complaint_assignments
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_admin ON complaint_assignments(assigned_to_admin_id);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_deadline ON complaint_assignments(resolution_deadline);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_escalation ON complaint_assignments(escalation_deadline) WHERE is_escalated = false;
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_status ON complaint_assignments(status);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_source_app ON complaint_assignments(source_app);
CREATE INDEX IF NOT EXISTS idx_complaint_assignments_complaint ON complaint_assignments(complaint_id);

-- =====================================================================
-- 3. UPDATE COMPLAINT_ESCALATIONS TABLE
-- =====================================================================

CREATE TABLE IF NOT EXISTS complaint_escalations (
    complaint_id UUID PRIMARY KEY REFERENCES report_issues(id) ON DELETE CASCADE,
    current_level INTEGER NOT NULL DEFAULT 1 CHECK (current_level IN (1, 2, 3)),
    current_admin_id UUID REFERENCES admin_users(id),
    current_team TEXT NOT NULL,
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    next_escalation_due TIMESTAMP WITH TIME ZONE,
    final_resolution_due TIMESTAMP WITH TIME ZONE,
    assigned_to_admin_id UUID REFERENCES admin_users(id),
    previous_assigned_admin_id UUID REFERENCES admin_users(id),
    is_active BOOLEAN DEFAULT true,
    last_escalated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now())
);

-- Indexes for complaint_escalations
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_next_due ON complaint_escalations(next_escalation_due) WHERE next_escalation_due IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_level ON complaint_escalations(current_level);
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_admin ON complaint_escalations(current_admin_id) WHERE current_admin_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_active ON complaint_escalations(is_active) WHERE is_active = true;

-- =====================================================================
-- 4. UPDATE COMPLAINT_ESCALATION_HISTORY TABLE
-- =====================================================================

CREATE TABLE IF NOT EXISTS complaint_escalation_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    complaint_id UUID NOT NULL REFERENCES report_issues(id) ON DELETE CASCADE,
    escalation_level INTEGER NOT NULL CHECK (escalation_level IN (1, 2, 3)),
    from_level INTEGER,
    to_level INTEGER,
    from_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    to_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    from_team TEXT,
    to_team TEXT,
    escalation_reason TEXT NOT NULL,
    escalation_type TEXT NOT NULL DEFAULT 'sla_timeout' CHECK (escalation_type IN (
        'sla_timeout',
        'manual_escalation',
        'priority_increase',
        'reassignment',
        'resolution'
    )),
    escalated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    escalated_by_admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now())
);

-- Indexes for complaint_escalation_history
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_complaint ON complaint_escalation_history(complaint_id);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_admin ON complaint_escalation_history(to_admin_id);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_type ON complaint_escalation_history(escalation_type);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_level ON complaint_escalation_history(escalation_level);
CREATE INDEX IF NOT EXISTS idx_complaint_escalation_history_created ON complaint_escalation_history(created_at DESC);

-- =====================================================================
-- 5. CREATE ESCALATION_SLA_CONFIG TABLE (New)
-- =====================================================================

CREATE TABLE IF NOT EXISTS escalation_sla_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_app TEXT NOT NULL CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger')),
    escalation_level INTEGER NOT NULL CHECK (escalation_level IN (1, 2, 3)),
    issue_priority TEXT NOT NULL DEFAULT 'medium' CHECK (issue_priority IN ('low', 'medium', 'high', 'emergency')),
    escalation_threshold_hours INTEGER NOT NULL,
    final_resolution_hours INTEGER NOT NULL,
    auto_escalate_on_timeout BOOLEAN DEFAULT true,
    notify_supervisor BOOLEAN DEFAULT true,
    revoke_admin_access_on_escalation BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    UNIQUE(source_app, escalation_level, issue_priority)
);

CREATE INDEX IF NOT EXISTS idx_escalation_sla_config_app_level ON escalation_sla_config(source_app, escalation_level);

-- =====================================================================
-- 6. INSERT DEFAULT SLA CONFIGURATIONS
-- =====================================================================

-- Delete existing default configurations if any
DELETE FROM escalation_sla_config WHERE created_by IS NULL;

-- Insert default SLA configurations for all apps and priorities
INSERT INTO escalation_sla_config (
    source_app, escalation_level, issue_priority, 
    escalation_threshold_hours, final_resolution_hours,
    auto_escalate_on_timeout, notify_supervisor, revoke_admin_access_on_escalation
)
VALUES
    -- BUS app configurations
    ('bus', 1, 'low', 48, 96, true, false, false),
    ('bus', 1, 'medium', 24, 48, true, true, false),
    ('bus', 1, 'high', 8, 24, true, true, true),
    ('bus', 1, 'emergency', 2, 6, true, true, true),
    
    ('bus', 2, 'low', 24, 48, true, true, true),
    ('bus', 2, 'medium', 12, 24, true, true, true),
    ('bus', 2, 'high', 4, 12, true, true, true),
    ('bus', 2, 'emergency', 1, 3, true, true, true),
    
    ('bus', 3, 'low', 8, 24, true, true, false),
    ('bus', 3, 'medium', 4, 12, true, true, false),
    ('bus', 3, 'high', 1, 4, true, true, false),
    ('bus', 3, 'emergency', 0.5, 2, true, true, false),
    
    -- DRIVER app configurations
    ('driver', 1, 'low', 48, 96, true, false, false),
    ('driver', 1, 'medium', 24, 48, true, true, false),
    ('driver', 1, 'high', 8, 24, true, true, true),
    ('driver', 1, 'emergency', 2, 6, true, true, true),
    
    ('driver', 2, 'low', 24, 48, true, true, true),
    ('driver', 2, 'medium', 12, 24, true, true, true),
    ('driver', 2, 'high', 4, 12, true, true, true),
    ('driver', 2, 'emergency', 1, 3, true, true, true),
    
    ('driver', 3, 'low', 8, 24, true, true, false),
    ('driver', 3, 'medium', 4, 12, true, true, false),
    ('driver', 3, 'high', 1, 4, true, true, false),
    ('driver', 3, 'emergency', 0.5, 2, true, true, false),
    
    -- LOUNGES app configurations
    ('lounges', 1, 'low', 72, 144, true, false, false),
    ('lounges', 1, 'medium', 48, 96, true, true, false),
    ('lounges', 1, 'high', 24, 48, true, true, true),
    ('lounges', 1, 'emergency', 4, 12, true, true, true),
    
    ('lounges', 2, 'low', 36, 72, true, true, true),
    ('lounges', 2, 'medium', 24, 48, true, true, true),
    ('lounges', 2, 'high', 12, 24, true, true, true),
    ('lounges', 2, 'emergency', 2, 6, true, true, true),
    
    ('lounges', 3, 'low', 12, 24, true, true, false),
    ('lounges', 3, 'medium', 6, 12, true, true, false),
    ('lounges', 3, 'high', 2, 4, true, true, false),
    ('lounges', 3, 'emergency', 0.5, 2, true, true, false),
    
    -- PASSENGER app configurations
    ('passenger', 1, 'low', 48, 96, true, false, false),
    ('passenger', 1, 'medium', 24, 48, true, true, false),
    ('passenger', 1, 'high', 8, 24, true, true, true),
    ('passenger', 1, 'emergency', 2, 6, true, true, true),
    
    ('passenger', 2, 'low', 24, 48, true, true, true),
    ('passenger', 2, 'medium', 12, 24, true, true, true),
    ('passenger', 2, 'high', 4, 12, true, true, true),
    ('passenger', 2, 'emergency', 1, 3, true, true, true),
    
    ('passenger', 3, 'low', 8, 24, true, true, false),
    ('passenger', 3, 'medium', 4, 12, true, true, false),
    ('passenger', 3, 'high', 1, 4, true, true, false),
    ('passenger', 3, 'emergency', 0.5, 2, true, true, false)
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 7. CREATE TRIGGERS FOR AUTO-UPDATE TIMESTAMPS
-- =====================================================================

CREATE OR REPLACE FUNCTION update_timestamp_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = timezone('utc'::text, now());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for complaint_assignments
DROP TRIGGER IF EXISTS trigger_complaint_assignments_timestamp ON complaint_assignments;
CREATE TRIGGER trigger_complaint_assignments_timestamp
BEFORE UPDATE ON complaint_assignments
FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

-- Trigger for complaint_escalations
DROP TRIGGER IF EXISTS trigger_complaint_escalations_timestamp ON complaint_escalations;
CREATE TRIGGER trigger_complaint_escalations_timestamp
BEFORE UPDATE ON complaint_escalations
FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

-- Trigger for admin_users
DROP TRIGGER IF EXISTS trigger_admin_users_timestamp ON admin_users;
CREATE TRIGGER trigger_admin_users_timestamp
BEFORE UPDATE ON admin_users
FOR EACH ROW
EXECUTE FUNCTION update_timestamp_column();

-- =====================================================================
-- 8. CREATE VIEWS FOR EASY QUERYING
-- =====================================================================

CREATE OR REPLACE VIEW v_complaint_escalation_dashboard AS
SELECT 
    ca.complaint_id,
    ca.assigned_to_admin_id,
    admin_current.full_name as assigned_admin_name,
    admin_current.email as assigned_admin_email,
    admin_current.app_scope,
    supervisor.full_name as supervisor_name,
    supervisor.email as supervisor_email,
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
    (SELECT COUNT(*) FROM complaint_escalation_history WHERE complaint_id = ca.complaint_id) as escalation_count,
    (SELECT MAX(escalated_at) FROM complaint_escalation_history WHERE complaint_id = ca.complaint_id) as last_escalation_at
FROM complaint_assignments ca
LEFT JOIN report_issues ri ON ca.complaint_id = ri.id
LEFT JOIN admin_users admin_current ON ca.assigned_to_admin_id = admin_current.id
LEFT JOIN admin_users supervisor ON admin_current.supervisor_id = supervisor.id;

-- View for admin workload
CREATE OR REPLACE VIEW v_admin_complaint_workload AS
SELECT 
    au.id,
    au.full_name,
    au.email,
    au.role,
    au.app_scope,
    COUNT(ca.id) FILTER (WHERE ca.status IN ('assigned', 'in_progress')) as active_complaints,
    COUNT(ca.id) FILTER (WHERE ca.status = 'resolved') as resolved_complaints,
    COUNT(ca.id) FILTER (WHERE ca.escalation_deadline < timezone('utc'::text, now()) AND ca.status IN ('assigned', 'in_progress')) as overdue_complaints,
    MIN(ca.escalation_deadline) FILTER (WHERE ca.status IN ('assigned', 'in_progress')) as next_deadline
FROM admin_users au
LEFT JOIN complaint_assignments ca ON au.id = ca.assigned_to_admin_id
GROUP BY au.id, au.full_name, au.email, au.role, au.app_scope;

-- =====================================================================
-- 9. CREATE HELPER FUNCTIONS
-- =====================================================================

-- Function to get SLA for a complaint
CREATE OR REPLACE FUNCTION get_complaint_sla(p_source_app TEXT, p_escalation_level INTEGER, p_priority TEXT)
RETURNS TABLE(escalation_threshold_hours INTEGER, final_resolution_hours INTEGER) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        esc.escalation_threshold_hours,
        esc.final_resolution_hours
    FROM escalation_sla_config esc
    WHERE esc.source_app = p_source_app
    AND esc.escalation_level = p_escalation_level
    AND esc.issue_priority = p_priority
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to auto-escalate overdue complaints
CREATE OR REPLACE FUNCTION auto_escalate_overdue_complaints()
RETURNS TABLE(complaint_id UUID, from_admin_id UUID, to_admin_id UUID, escalation_level INTEGER) AS $$
DECLARE
    v_complaint RECORD;
    v_supervisor_id UUID;
    v_current_level INTEGER;
BEGIN
    -- Find all complaints past escalation deadline
    FOR v_complaint IN
        SELECT ca.id, ca.complaint_id, ca.assigned_to_admin_id, ca.escalation_level, ca.escalation_deadline
        FROM complaint_assignments ca
        WHERE ca.is_escalated = false
        AND ca.escalation_deadline < timezone('utc'::text, now())
        AND ca.status IN ('assigned', 'in_progress')
        ORDER BY ca.escalation_deadline ASC
    LOOP
        -- Get supervisor for escalation
        SELECT supervisor_id INTO v_supervisor_id
        FROM admin_users
        WHERE id = v_complaint.assigned_to_admin_id;
        
        v_current_level := v_complaint.escalation_level + 1;
        
        -- Update assignment for escalation
        UPDATE complaint_assignments
        SET 
            is_escalated = true,
            escalation_reason = 'Auto-escalated due to SLA timeout',
            escalated_at = timezone('utc'::text, now()),
            previous_admin_id = assigned_to_admin_id,
            assigned_to_admin_id = COALESCE(v_supervisor_id, (SELECT id FROM admin_users WHERE role = 'super_admin' LIMIT 1)),
            escalation_level = v_current_level,
            assignment_type = 'escalated',
            status = 'escalated'
        WHERE complaint_id = v_complaint.complaint_id;
        
        -- Log escalation event
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
            v_complaint.complaint_id,
            v_current_level,
            v_complaint.escalation_level,
            v_current_level,
            v_complaint.assigned_to_admin_id,
            COALESCE(v_supervisor_id, (SELECT id FROM admin_users WHERE role = 'super_admin' LIMIT 1)),
            'SLA timeout - auto escalated',
            'sla_timeout'
        );
        
        RETURN QUERY SELECT v_complaint.complaint_id, v_complaint.assigned_to_admin_id, 
                           COALESCE(v_supervisor_id, (SELECT id FROM admin_users WHERE role = 'super_admin' LIMIT 1)), 
                           v_current_level;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Summary
-- =====================================================================

-- Tables created/updated:
-- 1. admin_users - User roles and hierarchy
-- 2. complaint_assignments - Current assignment of complaints to admins
-- 3. complaint_escalations - Current escalation state
-- 4. complaint_escalation_history - Audit trail of escalations
-- 5. escalation_sla_config - SLA configurations per app/level/priority

-- Views created:
-- 1. v_complaint_escalation_dashboard - Dashboard view for complaint status
-- 2. v_admin_complaint_workload - Admin workload overview

-- Functions created:
-- 1. get_complaint_sla() - Retrieve SLA for a complaint
-- 2. auto_escalate_overdue_complaints() - Auto-escalate past SLA complaints

-- Triggers created:
-- 1. Update timestamps on table modifications
-- All indexes created for performance optimization

COMMIT;
