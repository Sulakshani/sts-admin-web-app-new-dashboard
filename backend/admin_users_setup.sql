-- Admin Users Table Setup Script
-- This script creates and configures the admin_users table with proper constraints and indexes

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create admin_users table
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW()),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Add columns if table already exists without them
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_admin_users_app_scope ON admin_users(app_scope);
CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);
CREATE INDEX IF NOT EXISTS idx_admin_users_supervisor_id ON admin_users(supervisor_id);
CREATE INDEX IF NOT EXISTS idx_admin_users_created_at ON admin_users(created_at DESC);

-- Create validation function for admin hierarchy
CREATE OR REPLACE FUNCTION validate_admin_hierarchy()
RETURNS TRIGGER AS $$
BEGIN
    -- If role is admin, supervisor_id must be provided
    IF NEW.role = 'admin' AND NEW.supervisor_id IS NULL THEN
        RAISE EXCEPTION 'admin role requires supervisor_id';
    END IF;
    
    -- If role is admin, supervisor must be supervisor or super_admin
    IF NEW.role = 'admin' AND NEW.supervisor_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM admin_users 
            WHERE id = NEW.supervisor_id 
            AND role IN ('supervisor', 'super_admin')
            AND is_active = true
        ) THEN
            RAISE EXCEPTION 'supervisor_id must reference an active supervisor or super_admin';
        END IF;
    END IF;
    
    -- If role is supervisor or admin, app_scope must be set
    IF NEW.role IN ('supervisor', 'admin') AND (NEW.app_scope IS NULL OR NEW.app_scope = '') THEN
        RAISE EXCEPTION 'supervisor and admin roles require app_scope';
    END IF;
    
    -- If role is super_admin, app_scope and supervisor_id must be NULL
    IF NEW.role = 'super_admin' THEN
        NEW.app_scope := NULL;
        NEW.supervisor_id := NULL;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to admin_users table (drop and recreate to ensure it's active)
DROP TRIGGER IF EXISTS admin_hierarchy_check ON admin_users;
CREATE TRIGGER admin_hierarchy_check
BEFORE INSERT OR UPDATE ON admin_users
FOR EACH ROW
EXECUTE FUNCTION validate_admin_hierarchy();

-- Create function to update last login timestamp
CREATE OR REPLACE FUNCTION update_admin_last_login(admin_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE admin_users 
    SET last_login_at = TIMEZONE('utc'::text, NOW()),
        updated_at = TIMEZONE('utc'::text, NOW())
    WHERE id = admin_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to log admin actions (optional - for audit trail)
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    target_user_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin_id ON admin_audit_log(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created_at ON admin_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_action ON admin_audit_log(action);

-- Create function to log admin actions
CREATE OR REPLACE FUNCTION log_admin_action(
    p_admin_id UUID,
    p_action TEXT,
    p_target_user_id UUID DEFAULT NULL,
    p_details JSONB DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_audit_id UUID;
BEGIN
    INSERT INTO admin_audit_log (admin_id, action, target_user_id, details)
    VALUES (p_admin_id, p_action, p_target_user_id, p_details)
    RETURNING id INTO v_audit_id;
    
    RETURN v_audit_id;
END;
$$ LANGUAGE plpgsql;

-- Commit all changes
COMMIT;

-- Verify table was created
SELECT 'Admin users table created successfully' as status;
SELECT COUNT(*) as admin_user_count FROM admin_users;
