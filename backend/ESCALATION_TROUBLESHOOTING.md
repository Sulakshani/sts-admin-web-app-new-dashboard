# Escalation Troubleshooting Guide

## ❌ "Escalation Failed" - How to Fix

If you're seeing escalation failures, follow these steps:

## 🔍 Step 1: Run Diagnostic Tool

First, check what's wrong by running the diagnostic tool:

```bash
cd backend
go run cmd/check_escalation/main.go
```

This will show you:
- ✅ Database connection status
- ✅ Which tables exist
- ✅ SMS configuration
- ✅ Team assignments
- ✅ Current complaints status

## 🛠️ Step 2: Fix Common Issues

### Issue 1: Missing Database Tables ❌

**Symptom**: Diagnostic shows "Table 'complaint_escalations' does NOT exist"

**Solution**:
```bash
# Connect to your PostgreSQL database and run the schema
psql -U your_username -d your_database_name -f backend/complaint_escalation_schema.sql

# Or using pgAdmin:
# 1. Open pgAdmin
# 2. Connect to your database
# 3. Click Tools → Query Tool
# 4. Open complaint_escalation_schema.sql
# 5. Click Execute (F5)
```

**Verify**:
```sql
-- Check tables were created
SELECT table_name 
FROM information_schema.tables 
WHERE table_name IN ('complaint_escalations', 'complaint_escalation_history');
```

### Issue 2: SMS Not Configured ⚠️

**Symptom**: Diagnostic shows "ESMS_API_KEY not configured"

**Solution**: Edit your `.env` file:
```env
# Add these lines to backend/.env
ESMS_API_URL=https://api.esms.lk/v1/sms/send
ESMS_API_KEY=your_actual_api_key_here
ESMS_SENDER_ID=BusLounge
```

**Note**: Without SMS config, escalation will work but SMS won't be sent (which is okay for testing).

### Issue 3: Database Connection Failed ❌

**Symptom**: "Database connection failed" error

**Solution**: Check your `DATABASE_URL` in `.env`:
```env
DATABASE_URL=postgresql://username:password@localhost:5432/database_name
```

Test connection:
```bash
# Try connecting manually
psql "postgresql://username:password@localhost:5432/database_name"
```

### Issue 4: Type Mismatch Errors ❌

**Symptom**: SQL errors about type mismatches

**Solution**: The schema expects UUID types. Check your report_issues table:
```sql
-- Verify ID type
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'report_issues' 
AND column_name = 'id';

-- Should show: id | uuid
```

## 📋 Step-by-Step Fix (Most Common)

### For First-Time Setup:

1. **Create the escalation tables**:
   ```bash
   # Terminal in backend folder
   psql -U your_user -d your_database < complaint_escalation_schema.sql
   ```

2. **Verify tables exist**:
   ```bash
   go run cmd/check_escalation/main.go
   ```

3. **Configure SMS (optional)**:
   ```bash
   # Edit .env file
   nano .env  # or use any text editor
   
   # Add:
   ESMS_API_KEY=your_key_here
   ESMS_SENDER_ID=BusLounge
   ```

4. **Restart the backend server**:
   ```bash
   # Stop current server (Ctrl+C)
   go run cmd/server/main.go
   
   # You should see:
   # ✅ Complaint escalation scheduler started (runs every 1 hour)
   ```

5. **Test in UI**:
   - Go to Complaint Management
   - Click "View" on any complaint
   - You should see escalation info
   - Try clicking "Escalate to Next Level"

## 🧪 Manual Testing

### Test Escalation Initialization:

```sql
-- Create a test complaint (if needed)
INSERT INTO report_issues (id, reported_by_id, issue_type, priority, status, description)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM users LIMIT 1),
    'bus_delay',
    'medium',
    'reported',
    'Test complaint for escalation'
)
RETURNING id;

-- Check if escalation auto-initializes when you view it in UI
-- Or manually initialize:
INSERT INTO complaint_escalations (
    complaint_id, 
    current_level, 
    current_team,
    next_escalation_due
)
VALUES (
    'your-complaint-id-here',
    1,
    'Operations Team',
    NOW() + INTERVAL '5 days'
);
```

### Test Manual Escalation API:

```bash
# Get complaint ID from UI or database
COMPLAINT_ID="your-complaint-id-here"

# Test manual escalation
curl -X POST http://localhost:8080/api/complaints/$COMPLAINT_ID/escalate \
  -H "Content-Type: application/json" \
  -d '{"escalated_by": "admin_test"}'

# Should return:
# {"message":"Complaint escalated successfully","complaint_id":"..."}
```

## 🔧 Advanced Debugging

### Check Backend Logs:

When you run the server, watch for these logs:

**Good logs** ✅:
```
✅ Complaint escalation scheduler started (runs every 1 hour)
🔍 Running scheduled escalation check...
⬆️  Escalated complaint xxx from level 1 to level 2
📱 Sending escalation notification (Level 2) to 94715342627
✅ SMS sent successfully to 94715342627
```

**Error logs** ❌:
```
❌ Error: pq: relation "complaint_escalations" does not exist
   → Need to run complaint_escalation_schema.sql

❌ Error: failed to initialize escalation: no escalation config
   → Check assignment.go configuration

❌ Error: SMS API error
   → Check ESMS credentials
```

### Enable Detailed Logging:

In `escalation_service.go`, errors are already logged. Check terminal output.

### Database Queries for Debugging:

```sql
-- See all escalations
SELECT * FROM complaint_escalations;

-- See escalation history
SELECT * FROM complaint_escalation_history ORDER BY escalated_at DESC;

-- Find complaints ready to escalate
SELECT 
    ce.complaint_id,
    ce.current_level,
    ce.next_escalation_due,
    ri.issue_type,
    ri.status
FROM complaint_escalations ce
JOIN report_issues ri ON ce.complaint_id = ri.id
WHERE ce.next_escalation_due <= NOW()
AND ri.status NOT IN ('resolved', 'closed');
```

## ✅ Verification Checklist

After fixing, verify everything works:

- [ ] Diagnostic tool runs without errors
- [ ] Tables `complaint_escalations` and `complaint_escalation_history` exist
- [ ] Can view complaints in UI
- [ ] Escalation info shows in complaint view modal
- [ ] "Escalate to Next Level" button works
- [ ] Backend logs show successful escalation
- [ ] SMS sent (if configured) or warning logged (if not)

## 🆘 Still Not Working?

If you've tried everything above:

1. **Share the error message**: Copy the exact error from terminal/browser console

2. **Run diagnostic and share output**:
   ```bash
   go run cmd/check_escalation/main.go > diagnostic_output.txt
   ```

3. **Check these files exist**:
   - `backend/complaint_escalation_schema.sql` ✅
   - `backend/internal/config/assignment.go` ✅
   - `backend/internal/services/escalation_service.go` ✅
   - `backend/internal/services/sms_service.go` ✅

4. **Verify database connection**:
   ```bash
   psql $DATABASE_URL -c "SELECT NOW();"
   ```

## 📱 Testing Without SMS

You can test escalation WITHOUT SMS configured:

1. Leave `ESMS_API_KEY` empty in `.env`
2. System will log: `⚠️ SMS service not configured - skipping SMS notification`
3. Escalation will still work, just no SMS sent
4. Perfect for testing!

## 🎯 Quick Fix Command

If tables are missing, run this ONE command:

```bash
psql -U your_username -d your_database_name -f backend/complaint_escalation_schema.sql && go run cmd/check_escalation/main.go
```

This will:
1. Create the tables
2. Run diagnostics
3. Show you the status

Then restart your backend server!
