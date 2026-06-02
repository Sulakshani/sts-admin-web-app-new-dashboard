# 🚨 ESCALATION FAILED? - Quick Fix Guide

## The Most Common Reason

**❌ The escalation database tables don't exist yet!**

## ✅ Quick Fix (Choose One)

### Option 1: Automatic Setup (Windows)

```powershell
# In PowerShell, navigate to backend folder:
cd backend

# Run the setup script:
.\setup_escalation.ps1
```

### Option 2: Manual Setup

```bash
# 1. Navigate to backend folder
cd backend

# 2. Apply the escalation schema to your database
psql -U your_username -d your_database_name -f complaint_escalation_schema.sql

# 3. Run diagnostic to verify
go run cmd/check_escalation/main.go

# 4. Restart backend server
go run cmd/server/main.go
```

### Option 3: Using pgAdmin (GUI)

1. Open **pgAdmin**
2. Connect to your database
3. Right-click your database → **Query Tool**
4. Click **Open File** → Select `backend/complaint_escalation_schema.sql`
5. Click **Execute** (or press F5)
6. You should see: "Query returned successfully"

## 🔍 Verify It Worked

Run the diagnostic tool:
```bash
cd backend
go run cmd/check_escalation/main.go
```

You should see:
```
✅ Database connected successfully
✅ Table 'complaint_escalations' exists
✅ Table 'complaint_escalation_history' exists
✅ All category assignments configured correctly
```

## 📋 Complete Troubleshooting

If the quick fix doesn't work, see: **[ESCALATION_TROUBLESHOOTING.md](ESCALATION_TROUBLESHOOTING.md)**

## 🎯 What Each File Does

- `complaint_escalation_schema.sql` - Creates the database tables ⭐ **RUN THIS FIRST**
- `check_escalation_setup.sql` - SQL queries to check status
- `cmd/check_escalation/main.go` - Diagnostic tool  
- `setup_escalation.ps1` - Automated setup script
- `ESCALATION_TROUBLESHOOTING.md` - Detailed troubleshooting guide

## ⚡ Super Quick Test

After setup, test immediately:

1. **Start backend**: `go run cmd/server/main.go`
2. **Open UI**: Go to Complaint Management
3. **View any complaint**: Click the eye icon
4. **Look for**: "Escalation Status" section
5. **Try**: Click "Escalate to Next Level" button

If you see the escalation info → ✅ **IT WORKS!**

If you still see errors → 📖 Read [ESCALATION_TROUBLESHOOTING.md](ESCALATION_TROUBLESHOOTING.md)
