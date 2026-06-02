# Complaint Escalation System - Authentication Integration & Implementation Guide

## Overview

This document outlines how the complaint escalation system integrates with the existing admin authentication system and provides step-by-step implementation guide.

---

## Authentication Integration

### JWT Token Enhancement

The existing JWT token from `AdminLogin` should be enhanced to include:

```json
{
    "admin_id": "uuid-of-admin",
    "email": "admin@example.com",
    "role": "admin",           // NEW: For complaint access control
    "app_scope": "bus",        // NEW: For filtering complaints
    "permissions": [
        "complaints.read",
        "complaints.assign",
        "complaints.update"
    ],
    "supervisor_id": "uuid-of-supervisor",  // NEW: For escalation
    "iat": 1234567890,
    "exp": 1234571490
}
```

### Updated AdminLogin Handler

```go
// In internal/handlers/admin_auth_handler.go

func AdminLogin(c *gin.Context) {
    var req struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
        return
    }

    admin, exists := getAdminByEmail(req.Email)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    // Update last_login_at
    _, _ = database.DB.Exec(
        "UPDATE admin_users SET last_login_at = TIMEZONE('utc'::text, now()) WHERE id::text = $1",
        admin.ID,
    )

    // Generate JWT with enhanced claims
    claims := jwt.MapClaims{
        "admin_id":     admin.ID,
        "email":        admin.Email,
        "full_name":    admin.FullName,
        "role":         admin.Role,           // NEW
        "app_scope":    admin.AppScope,       // NEW
        "permissions":  admin.Permissions,
        "supervisor_id": admin.SupervisorID,  // NEW
        "iat":          time.Now().Unix(),
        "exp":          time.Now().Add(time.Hour * 24).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(jwtSecret)

    c.JSON(http.StatusOK, gin.H{
        "access_token": tokenString,
        "user": gin.H{
            "id":        admin.ID,
            "email":     admin.Email,
            "full_name": admin.FullName,
            "role":      admin.Role,
            "app_scope": admin.AppScope,
        },
    })
}
```

---

## Middleware for Complaint Access Control

### Middleware: Check Complaint Access Permission

```go
// In internal/handlers/middleware.go

// CheckComplaintAccess verifies if admin has access to a specific complaint
func CheckComplaintAccess(c *gin.Context) {
    // Get complaint ID from URL parameter
    complaintID := c.Param("id")
    
    // Get admin info from JWT
    claims, err := getJWTClaimsFromRequest(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        c.Abort()
        return
    }
    
    adminID, _ := claims["admin_id"].(string)
    role, _ := claims["role"].(string)
    appScope, _ := claims["app_scope"].(string)
    
    // Super admin can access all complaints
    if role == "super_admin" {
        c.Set("admin_id", adminID)
        c.Next()
        return
    }
    
    // Check if admin is assigned to this complaint
    var assignedTo string
    err = database.DB.QueryRow(`
        SELECT assigned_to_admin_id::text FROM complaint_assignments
        WHERE complaint_id::text = $1
    `, complaintID).Scan(&assignedTo)
    
    if err != nil || assignedTo != adminID {
        // Check if admin is supervisor of assigned admin
        var supervisorID string
        err = database.DB.QueryRow(`
            SELECT supervisor_id::text FROM admin_users
            WHERE id::text = (
                SELECT assigned_to_admin_id::text FROM complaint_assignments
                WHERE complaint_id::text = $1
            )
        `, complaintID).Scan(&supervisorID)
        
        if err != nil || supervisorID != adminID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this complaint"})
            c.Abort()
            return
        }
    }
    
    c.Set("admin_id", adminID)
    c.Set("complaint_id", complaintID)
    c.Next()
}
```

### Middleware: Filter Complaints by Role

```go
// GetAdminComplaints returns complaints based on admin's role
func GetAdminComplaints(c *gin.Context) {
    claims, err := getJWTClaimsFromRequest(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    
    adminID, _ := claims["admin_id"].(string)
    role, _ := claims["role"].(string)
    appScope, _ := claims["app_scope"].(string)
    
    var query string
    var args []interface{}
    
    if role == "super_admin" {
        // Super admin sees all complaints
        query = `
            SELECT ca.*, ri.issue_type, ri.priority, ri.status as complaint_status
            FROM complaint_assignments ca
            JOIN report_issues ri ON ca.complaint_id = ri.id
            WHERE ca.status IN ('assigned', 'in_progress', 'escalated')
            ORDER BY ca.escalation_deadline ASC
        `
    } else if role == "supervisor" {
        // Supervisor sees assigned complaints and their team's complaints
        query = `
            SELECT ca.*, ri.issue_type, ri.priority, ri.status as complaint_status
            FROM complaint_assignments ca
            JOIN report_issues ri ON ca.complaint_id = ri.id
            LEFT JOIN admin_users admin ON ca.assigned_to_admin_id = admin.id
            WHERE (
                ca.assigned_to_admin_id::text = $1  -- Assigned to supervisor
                OR admin.supervisor_id::text = $1    -- Assigned to supervisor's team
            )
            AND ca.status IN ('assigned', 'in_progress', 'escalated')
            AND ca.escalation_level >= 2
            ORDER BY ca.escalation_deadline ASC
        `
        args = []interface{}{adminID}
    } else {
        // Admin sees only their assigned complaints
        query = `
            SELECT ca.*, ri.issue_type, ri.priority, ri.status as complaint_status
            FROM complaint_assignments ca
            JOIN report_issues ri ON ca.complaint_id = ri.id
            WHERE ca.assigned_to_admin_id::text = $1
            AND ca.status IN ('assigned', 'in_progress')
            AND ca.escalation_level = 1
            ORDER BY ca.escalation_deadline ASC
        `
        args = []interface{}{adminID}
    }
    
    rows, err := database.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch complaints"})
        return
    }
    defer rows.Close()
    
    // Parse and return complaints
    var complaints []map[string]interface{}
    for rows.Next() {
        // Scan complaint data
        // ... implementation
    }
    
    c.JSON(http.StatusOK, complaints)
}
```

---

## API Endpoints for Complaint Escalation

### 1. Get Assigned Complaints (Admin)

**Endpoint**: `GET /api/complaints/assigned`

**Authentication**: JWT Token (Admin role required)

**Response**:
```json
{
    "complaints": [
        {
            "complaint_id": "uuid",
            "issue_type": "bus_delay",
            "priority": "high",
            "description": "Bus delayed by 2 hours",
            "status": "assigned",
            "assigned_to": "Admin Name",
            "resolution_deadline": "2026-03-31T15:00:00Z",
            "escalation_deadline": "2026-03-31T03:00:00Z",
            "hours_remaining": 12.5,
            "urgency": "CRITICAL"
        }
    ]
}
```

### 2. Get Escalated Complaints (Supervisor)

**Endpoint**: `GET /api/complaints/escalated`

**Authentication**: JWT Token (Supervisor role required)

**Response**:
```json
{
    "complaints": [
        {
            "complaint_id": "uuid",
            "previous_admin": "Admin Name",
            "escalation_reason": "SLA timeout - auto escalated",
            "escalation_level": 2,
            "source_app": "bus",
            "status": "escalated",
            "escalated_at": "2026-03-30T15:00:00Z"
        }
    ]
}
```

### 3. Get All Complaints (Super Admin)

**Endpoint**: `GET /api/complaints/all`

**Authentication**: JWT Token (Super Admin role required)

**Query Parameters**:
- `status`: assigned | in_progress | escalated | resolved
- `source_app`: bus | driver | lounges | passenger
- `escalation_level`: 1 | 2 | 3
- `urgency`: ON_TRACK | CRITICAL | ESCALATION_DUE | OVERDUE

**Response**:
```json
{
    "total": 154,
    "complaints": [
        {
            "complaint_id": "uuid",
            "assigned_admin": "Admin Name",
            "supervisor": "Supervisor Name",
            "escalation_level": 2,
            "status": "escalated",
            "urgency": "OVERDUE",
            "hours_overdue": 3.5
        }
    ]
}
```

### 4. Update Complaint Status

**Endpoint**: `PUT /api/complaints/{id}/status`

**Authentication**: JWT Token (Admin/Supervisor)

**Request**:
```json
{
    "status": "in_progress",
    "notes": "Looking into this issue"
}
```

**Response**:
```json
{
    "success": true,
    "message": "Complaint status updated",
    "complaint_id": "uuid",
    "new_status": "in_progress",
    "updated_at": "2026-03-30T15:30:00Z"
}
```

### 5. Resolve Complaint

**Endpoint**: `POST /api/complaints/{id}/resolve`

**Authentication**: JWT Token (Admin role required)

**Request**:
```json
{
    "resolution_notes": "Issue resolved by restarting the service",
    "status": "resolved"
}
```

**Response**:
```json
{
    "success": true,
    "message": "Complaint marked as resolved",
    "complaint_id": "uuid",
    "resolved_by": "Admin Name",
    "resolved_at": "2026-03-30T15:45:00Z"
}
```

### 6. Manually Escalate Complaint

**Endpoint**: `POST /api/complaints/{id}/escalate`

**Authentication**: JWT Token (Supervisor role required)

**Request**:
```json
{
    "reason": "Needs super admin intervention",
    "escalation_type": "manual_escalation"
}
```

**Response**:
```json
{
    "success": true,
    "escalated_to": "Super Admin Name",
    "escalation_level": 3,
    "escalated_at": "2026-03-30T15:50:00Z"
}
```

---

## Database Setup Instructions

### Step 1: Connect to PostgreSQL

```bash
psql -U postgres -h localhost -d your_database_name
```

### Step 2: Run Setup Script

```bash
# Option A: Run entire script
psql -U postgres -h localhost -d your_database_name -f complaint_escalation_complete_setup.sql

# Option B: Run line by line in psql
\i complaint_escalation_complete_setup.sql
```

### Step 3: Verify Tables

```sql
-- Check all tables created
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE '%complaint%' 
OR table_name = 'escalation_sla_config';

-- Check admin_users table
\d admin_users

-- Check complaint_assignments table
\d complaint_assignments
```

### Step 4: Verify Indexes

```sql
SELECT indexname, tablename 
FROM pg_indexes 
WHERE tablename IN ('admin_users', 'complaint_assignments', 'complaint_escalations', 'complaint_escalation_history');
```

---

## Backend Implementation Checklist

- [ ] Update `AdminLogin` handler to include role, app_scope, supervisor_id in JWT
- [ ] Create `CheckComplaintAccess` middleware
- [ ] Create `GetAdminComplaints` handler with role-based filtering
- [ ] Create `GetEscalatedComplaints` handler for supervisors
- [ ] Create `GetAllComplaints` handler for super admins
- [ ] Create `UpdateComplaintStatus` handler
- [ ] Create `ResolveComplaint` handler
- [ ] Create `ManuallyEscalateComplaint` handler
- [ ] Create scheduler function `auto_escalate_overdue_complaints()`
- [ ] Add scheduler task to run hourly
- [ ] Implement complaint assignment logic on complaint creation
- [ ] Add permission checks in all complaint endpoints
- [ ] Add audit logging for escalation events
- [ ] Create dashboard endpoints returning views

---

## Scheduler Implementation

### Go Code for Hourly Escalation Check

```go
// In cmd/server/main.go

func startComplaintEscalationScheduler() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    go func() {
        for range ticker.C {
            err := escalateOverdueComplaints()
            if err != nil {
                log.Printf("Error escal: %v", err)
            }
        }
    }()
}

func escalateOverdueComplaints() error {
    // Call PostgreSQL function
    rows, err := database.DB.Query("SELECT * FROM auto_escalate_overdue_complaints()")
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var complaintID, fromAdminID, toAdminID string
        var escalationLevel int
        
        if err := rows.Scan(&complaintID, &fromAdminID, &toAdminID, &escalationLevel); err != nil {
            log.Printf("Error scanning escalation result: %v", err)
            continue
        }
        
        // Log escalation event
        log.Printf("Escalated complaint %s from admin %s to %s (level %d)", 
            complaintID, fromAdminID, toAdminID, escalationLevel)
        
        // Send notifications (if needed)
        // notifyAdminOfEscalation(toAdminID, complaintID)
    }
    
    return rows.Err()
}

// Call in main()
func main() {
    // ... existing code ...
    startComplaintEscalationScheduler()
    // ... rest of code ...
}
```

---

## Testing Checklist

- [ ] Create admin for 'bus' app scope
- [ ] Create complaint with source_app = 'bus'
- [ ] Verify auto-assignment to bus admin
- [ ] Wait for escalation deadline (or manually trigger)
- [ ] Verify auto-escalation to supervisor
- [ ] Verify bus admin loses access to complaint
- [ ] Login as supervisor and verify see escalated complaint
- [ ] Resolve as supervisor
- [ ] Verify super admin sees all complaints
- [ ] Test manual escalation
- [ ] Test role-based filtering
- [ ] Test permission checks on each endpoint

---

## Summary of Changes

### Tables Updated/Created
| Table | Status | Purpose |
|-------|--------|---------|
| admin_users | Updated | Add app_scope, supervisor_id for role hierarchy |
| complaint_assignments | Created | Track complaint assignments and SLA tracking |
| complaint_escalations | Updated | Current escalation state |
| complaint_escalation_history | Updated | Audit trail with admin transitions |
| escalation_sla_config | Created | Configurable SLA times |

### Authentication Changes
| Component | Change |
|-----------|--------|
| JWT Token | Add role, app_scope, supervisor_id |
| AdminLogin | Return role and app_scope in response |
| Middleware | Add CheckComplaintAccess and role-based filtering |

### New Endpoints
- GET /api/complaints/assigned
- GET /api/complaints/escalated
- GET /api/complaints/all
- PUT /api/complaints/{id}/status
- POST /api/complaints/{id}/resolve
- POST /api/complaints/{id}/escalate

### Scheduler Task
- Run every 1 hour
- Auto-escalate complaints past deadline
- Update assignments and log events

