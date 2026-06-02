# Admin Authentication Database Schema

## Overview
This document provides complete database schema definitions and SQL commands for implementing the admin authentication system with role-based access control (RBAC).

---

## Tables

### 1. **admin_users** (Main Admin Users Table)

#### Purpose
Stores all admin users with roles (super_admin, supervisor, admin), permissions, and access control information.

#### Create Table SQL

```sql
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
```

#### Column Definitions

| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique identifier for each admin user |
| `email` | TEXT | NOT NULL, UNIQUE | User's email address (login credential) |
| `password_hash` | TEXT | NOT NULL | Bcrypt-hashed password for authentication |
| `full_name` | TEXT | NOT NULL | Full name of the admin user |
| `role` | TEXT | NOT NULL, CHECK (super_admin\|supervisor\|admin) | User's role determining permissions |
| `app_scope` | TEXT | CHECK (bus\|driver\|lounges\|passenger) | Application scope the user manages (required for supervisor/admin) |
| `supervisor_id` | UUID | FK→admin_users(id) | Reference to parent supervisor (required only for admin role) |
| `function_permissions` | JSONB | NOT NULL, DEFAULT [] | Array of permission strings (e.g., ["buses.read", "buses.write", "\*"]) |
| `is_active` | BOOLEAN | NOT NULL, DEFAULT true | Whether the user account is active |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Account creation timestamp |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update timestamp |
| `last_login_at` | TIMESTAMP | NULLABLE | Last successful login timestamp (tracked for analytics) |
| `created_by` | UUID | FK→admin_users(id) | ID of the admin who created this user |

---

## Role Hierarchy and Rules

### Role Definitions

#### 1. **super_admin**
- **Permissions**: Full system access (`permissions = ["*"]`)
- **App Scope**: NULL (not required)
- **Supervisor ID**: NULL (not required)
- **Capabilities**:
  - Create/edit/delete any admin user
  - View all admin accounts
  - Manage all scopes (bus, driver, lounges, passenger)

#### 2. **supervisor**
- **Permissions**: Customizable (e.g., `["buses.read", "buses.verify"]`)
- **App Scope**: REQUIRED (one of: bus, driver, lounges, passenger)
- **Supervisor ID**: OPTIONAL (can exist standalone)
- **Capabilities**:
  - Oversee admin users within their app_scope
  - View and verify entities in their scope
  - Assign admin users to manage specific entities

#### 3. **admin**
- **Permissions**: Customizable (e.g., `["buses.write", "buses.verify"]`)
- **App Scope**: REQUIRED (inherited from supervisor)
- **Supervisor ID**: REQUIRED (must be a supervisor or super_admin in same scope)
- **Capabilities**:
  - Manage entities within their scope
  - Execute operations assigned by supervisor
  - Limited to their supervisor's scope

### Hierarchy Validation Rules

```sql
-- Validation Rules (implemented in backend):
-- 1. If role = 'super_admin':
--    - app_scope MUST be NULL
--    - supervisor_id MUST be NULL
--    - permissions MUST contain "*"

-- 2. If role = 'supervisor':
--    - app_scope MUST be non-empty and valid
--    - supervisor_id IS OPTIONAL
--    - permissions MUST be a valid array

-- 3. If role = 'admin':
--    - app_scope MUST be non-empty and valid
--    - supervisor_id MUST reference an active supervisor or super_admin
--    - Supervisor's app_scope MUST match admin's app_scope
--    - permissions MUST be a valid array
```

---

## SQL Commands

### Add Audit Columns (if updating existing table)

```sql
-- Add last_login_at column
ALTER TABLE admin_users 
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Add created_by column  
ALTER TABLE admin_users 
ADD COLUMN IF NOT EXISTS created_by UUID 
REFERENCES admin_users(id) ON DELETE SET NULL;

-- Add function_permissions column (if not already present)
ALTER TABLE admin_users 
ADD COLUMN IF NOT EXISTS function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb;
```

### Create Indexes for Performance

```sql
-- Index on email for login queries
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);

-- Index on role for filtering by role
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);

-- Index on app_scope for scope-based queries
CREATE INDEX IF NOT EXISTS idx_admin_users_app_scope ON admin_users(app_scope);

-- Index on is_active for status filtering
CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);

-- Index on supervisor_id for hierarchy queries
CREATE INDEX IF NOT EXISTS idx_admin_users_supervisor_id ON admin_users(supervisor_id);
```

### Create or Replace Function: Validate Supervisor Hierarchy

```sql
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
    IF NEW.role IN ('supervisor', 'admin') AND NEW.app_scope IS NULL THEN
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

-- Attach trigger to admin_users table
DROP TRIGGER IF EXISTS admin_hierarchy_check ON admin_users;
CREATE TRIGGER admin_hierarchy_check
BEFORE INSERT OR UPDATE ON admin_users
FOR EACH ROW
EXECUTE FUNCTION validate_admin_hierarchy();
```

### Create or Replace Function: Update Last Login

```sql
CREATE OR REPLACE FUNCTION update_admin_last_login(admin_id UUID)
RETURNS void AS $$
BEGIN
    UPDATE admin_users 
    SET last_login_at = TIMEZONE('utc'::text, NOW())
    WHERE id = admin_id;
END;
$$ LANGUAGE plpgsql;
```

---

## Sample Data

### Insert Super Admin (Bootstrap)

```sql
-- Bootstrap super admin (requires bcrypt hash of password)
INSERT INTO admin_users (
    id, email, password_hash, full_name, role, 
    app_scope, supervisor_id, function_permissions, is_active
)
VALUES (
    uuid_generate_v4(),
    'admin@example.com',
    '$2a$10$...',  -- Bcrypt hash of password
    'System Administrator',
    'super_admin',
    NULL,
    NULL,
    '["*"]'::jsonb,
    true
);
```

### Insert Supervisor User

```sql
-- Supervisor for 'bus' scope
INSERT INTO admin_users (
    id, email, password_hash, full_name, role,
    app_scope, supervisor_id, function_permissions, is_active, created_by
)
VALUES (
    uuid_generate_v4(),
    'bus-supervisor@example.com',
    '$2a$10$...',  -- Bcrypt hash
    'Bus Operations Supervisor',
    'supervisor',
    'bus',
    NULL,  -- Supervisors don't require a parent supervisor
    '["buses.read", "buses.verify", "buses.write"]'::jsonb,
    true,
    (SELECT id FROM admin_users WHERE role = 'super_admin' LIMIT 1)
);
```

### Insert Admin User

```sql
-- Admin user under supervisor
INSERT INTO admin_users (
    id, email, password_hash, full_name, role,
    app_scope, supervisor_id, function_permissions, is_active, created_by
)
VALUES (
    uuid_generate_v4(),
    'bus-admin@example.com',
    '$2a$10$...',  -- Bcrypt hash
    'Bus Administrator',
    'admin',
    'bus',  -- Must match supervisor's scope
    (SELECT id FROM admin_users WHERE email = 'bus-supervisor@example.com'),
    '["buses.read", "buses.write"]'::jsonb,
    true,
    (SELECT id FROM admin_users WHERE role = 'super_admin' LIMIT 1)
);
```

---

## Utility Queries

### List All Active Admins

```sql
SELECT 
    id::text,
    email,
    full_name,
    role,
    app_scope,
    is_active,
    created_at,
    last_login_at
FROM admin_users
WHERE is_active = true
ORDER BY created_at DESC;
```

### Find All Users Under a Supervisor

```sql
SELECT 
    id::text,
    email,
    full_name,
    role,
    app_scope,
    is_active
FROM admin_users
WHERE supervisor_id = $1  -- Pass supervisor UUID
AND is_active = true;
```

### Get Users by Scope

```sql
SELECT 
    id::text,
    email,
    full_name,
    role,
    is_active
FROM admin_users
WHERE app_scope = $1  -- Pass scope: 'bus', 'driver', 'lounges', 'passenger'
AND is_active = true
ORDER BY role DESC;
```

### Update Last Login

```sql
UPDATE admin_users
SET last_login_at = TIMEZONE('utc'::text, NOW())
WHERE id = $1;
```

### Reset User Password

```sql
UPDATE admin_users
SET password_hash = $1,
    updated_at = TIMEZONE('utc'::text, NOW())
WHERE email = $2;
```

### Deactivate User

```sql
UPDATE admin_users
SET is_active = false,
    updated_at = TIMEZONE('utc'::text, NOW())
WHERE id = $1;
```

### Delete User (with safety check)

```sql
-- Only delete if not super_admin and no one depends on them
DELETE FROM admin_users
WHERE id = $1
AND role != 'super_admin'
AND NOT EXISTS (
    SELECT 1 FROM admin_users 
    WHERE supervisor_id = admin_users.id
);
```

---

## Environment Variables Required

```bash
# .env file configuration
JWT_SECRET=your-jwt-secret-key-here
SUPERADMIN_EMAIL=admin@example.com
SUPERADMIN_PASSWORD=SecurePassword123
SUPERADMIN_NAME=System Administrator
```

---

## Bootstrap Process

The backend automatically creates the first super admin on startup if:
1. `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD`, and `SUPERADMIN_NAME` are set in `.env`
2. No admin user exists with that email
3. Database connection is available

Status endpoint: `GET /api/admin/auth/bootstrap-status`

Returns:
```json
{
    "status": "created|already_exists|skipped_missing_env|db_unavailable|error",
    "message": "Human-readable status message",
    "configured": true/false
}
```

---

## API Integration Points

### Create User (Super Admin Only)

**Endpoint**: `POST /api/admin/users`

**Request**:
```json
{
    "email": "admin@example.com",
    "full_name": "Admin Name",
    "password": "SecurePassword123",
    "role": "admin",
    "app_scope": "bus",
    "supervisor_id": "uuid-of-supervisor",
    "permissions": ["buses.read", "buses.write"]
}
```

**Validation Rules**:
- `supervisor_id` is required only if `role == 'admin'`
- `supervisor_id` is ignored if `role == 'supervisor'` or `role == 'super_admin'`
- `app_scope` is required for supervisor and admin roles
- `permissions` must be valid permission strings

### Login

**Endpoint**: `POST /api/admin/auth/login`

**Request**:
```json
{
    "email": "admin@example.com",
    "password": "SecurePassword123"
}
```

**Response**:
```json
{
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "user": {
        "id": "user-uuid",
        "email": "admin@example.com",
        "full_name": "Admin Name",
        "role": "admin",
        "app_scope": "bus"
    }
}
```

Updates `last_login_at` timestamp on successful login.

---

## Migration Checklist

When implementing this schema:

- [ ] Create `admin_users` table with all columns
- [ ] Create indexes on frequently queried columns
- [ ] Create validation trigger for role hierarchy
- [ ] Bootstrap super admin account on first run
- [ ] Set environment variables in `.env`
- [ ] Run backend to auto-initialize tables
- [ ] Test login with bootstrap account
- [ ] Create supervisor account via API
- [ ] Create admin account under supervisor via API
- [ ] Verify role hierarchy enforcement

