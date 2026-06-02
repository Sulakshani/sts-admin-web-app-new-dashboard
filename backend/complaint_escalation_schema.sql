-- Complaint Escalation System Schema
-- This schema adds escalation tracking for complaints

-- Admin hierarchy fields required for app-specific assignment and supervisor fallback
ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS role VARCHAR NOT NULL DEFAULT 'admin',
    ADD COLUMN IF NOT EXISTS app_scope VARCHAR,
    ADD COLUMN IF NOT EXISTS supervisor_id UUID REFERENCES admin_users(id),
    ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES admin_users(id),
    ADD COLUMN IF NOT EXISTS function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'admin_users_role_check'
          AND conrelid = 'admin_users'::regclass
    ) THEN
        ALTER TABLE admin_users
            ADD CONSTRAINT admin_users_role_check
            CHECK (role IN ('super_admin', 'supervisor', 'admin'));
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'admin_users_app_scope_check'
          AND conrelid = 'admin_users'::regclass
    ) THEN
        ALTER TABLE admin_users
            ADD CONSTRAINT admin_users_app_scope_check
            CHECK (app_scope IS NULL OR app_scope IN ('bus', 'driver', 'lounges', 'passenger'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_admin_users_role_scope_active
    ON admin_users(role, app_scope, is_active);

CREATE INDEX IF NOT EXISTS idx_admin_users_supervisor
    ON admin_users(supervisor_id)
    WHERE supervisor_id IS NOT NULL;

-- Table to track current escalation status of each complaint
CREATE TABLE IF NOT EXISTS complaint_escalations (
    complaint_id UUID PRIMARY KEY REFERENCES report_issues(id) ON DELETE CASCADE,
    current_level INTEGER NOT NULL DEFAULT 1,
    current_team VARCHAR NOT NULL,
    source_app VARCHAR NOT NULL,
    assigned_to_admin_id UUID REFERENCES admin_users(id),
    previous_assigned_admin_id UUID REFERENCES admin_users(id),
    last_escalated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    next_escalation_due TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'complaint_escalations_source_app_check'
          AND conrelid = 'complaint_escalations'::regclass
    ) THEN
        ALTER TABLE complaint_escalations
            ADD CONSTRAINT complaint_escalations_source_app_check
            CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger'));
    END IF;
END $$;

-- Table to track escalation history for audit trail
CREATE TABLE IF NOT EXISTS complaint_escalation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    complaint_id UUID NOT NULL REFERENCES report_issues(id) ON DELETE CASCADE,
    level INTEGER NOT NULL,
    team_name VARCHAR NOT NULL,
    escalated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    escalated_by VARCHAR NOT NULL, -- 'auto' or admin user ID
    from_admin_id UUID REFERENCES admin_users(id),
    to_admin_id UUID REFERENCES admin_users(id),
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Backward-compatible alias for integrations expecting plural table name.
CREATE OR REPLACE VIEW complaint_escalations_history AS
SELECT *
FROM complaint_escalation_history;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_complaint_escalations_next_due 
    ON complaint_escalations(next_escalation_due) 
    WHERE next_escalation_due IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_complaint_escalations_level 
    ON complaint_escalations(current_level);

CREATE INDEX IF NOT EXISTS idx_complaint_escalations_assigned_to 
    ON complaint_escalations(assigned_to_admin_id) 
    WHERE assigned_to_admin_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_complaint_escalations_source_app
    ON complaint_escalations(source_app);

CREATE INDEX IF NOT EXISTS idx_escalation_history_complaint 
    ON complaint_escalation_history(complaint_id, escalated_at);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_complaint_escalation_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_complaint_escalation_timestamp
    BEFORE UPDATE ON complaint_escalations
    FOR EACH ROW
    EXECUTE FUNCTION update_complaint_escalation_updated_at();

-- Comments for documentation
COMMENT ON TABLE complaint_escalations IS 'Tracks the current escalation level and team assignment for each complaint';
COMMENT ON TABLE complaint_escalation_history IS 'Maintains a complete audit trail of all escalation events';
COMMENT ON COLUMN complaint_escalations.current_level IS 'Current escalation level (1, 2, 3, etc.)';
COMMENT ON COLUMN complaint_escalations.current_team IS 'Name of the team currently responsible for the complaint';
COMMENT ON COLUMN complaint_escalations.assigned_to_admin_id IS 'Specific admin user assigned to handle this complaint';
COMMENT ON COLUMN complaint_escalations.next_escalation_due IS 'When the complaint should be escalated to the next level if not resolved';
