# Team Assignment & Category Matching Guide

## ✅ Automatic Matching Implemented

The system now automatically matches complaint categories with assigned teams without requiring database editing!

## 🔄 How It Works

### 1. **Issue Type → Category Mapping**

When a complaint is created with an `issue_type`, it's automatically mapped to an escalation category:

| Issue Type | Escalation Category | Assigned Team (Level 1) |
|-----------|-------------------|------------------------|
| `bus_delay` | Operations & Scheduling | Operations Team |
| `maintenance_issue` | Vehicle & Facility | Maintenance Team |
| `flat_wheel` | Vehicle & Facility | Maintenance Team |
| `equipment_malfunction` | Vehicle & Facility | Maintenance Team |
| `passenger_complaint` | Service Issue | Customer Service |
| `safety_concern` | Safety & Security | Safety & Security |
| Any other type | Other | General Support |

### 2. **Auto-Initialization**

When you view complaints, the system automatically:

1. ✅ **Converts** the issue_type to the proper category
2. ✅ **Checks** if escalation exists for this complaint
3. ✅ **Creates** escalation tracking if it doesn't exist (for pending/in-progress complaints)
4. ✅ **Assigns** to Level 1 team based on category
5. ✅ **Sets** next escalation date (5 days from creation)
6. ✅ **Sends** SMS to assigned team

### 3. **Team Assignment by Level**

The system uses the configuration in `backend/internal/config/assignment.go`:

```go
"Service Issue": {
    1: {"Customer Service", "Support Agent", "94771234567"},
    2: {"Customer Service", "Team Lead", "94771234568"},
    3: {"Customer Service", "Manager", "94771234569"},
},
```

This ensures:
- **Consistent Naming**: Team names are consistent across all complaints in the same category
- **Automatic Assignment**: No manual database editing needed
- **Level Progression**: Teams change as complaints escalate

## 📊 Complete Category-to-Team Mapping

### Service Issue (Passenger Complaints)
- **Level 1**: Customer Service - Support Agent
- **Level 2**: Customer Service - Team Lead
- **Level 3**: Customer Service - Manager

### Operations & Scheduling (Bus Delays)
- **Level 1**: Operations Team - Operations Officer
- **Level 2**: Operations Team - Operations Manager
- **Level 3**: Operations Team - Operations Director

### Vehicle & Facility (Maintenance/Equipment)
- **Level 1**: Maintenance Team - Maintenance Technician
- **Level 2**: Maintenance Team - Maintenance Supervisor
- **Level 3**: Maintenance Team - Maintenance Manager

### Safety & Security (Safety Concerns)
- **Level 1**: Safety & Security - Security Officer
- **Level 2**: Safety & Security - Security Manager
- **Level 3**: Safety & Security - Security Director

### Other (Unknown Types)
- **Level 1**: General Support - Support Agent
- **Level 2**: General Support - Support Manager
- **Level 3**: General Support - General Manager

## 🎯 What This Means

### Before (Manual Process)
1. ❌ Complaint created with issue_type = "bus_delay"
2. ❌ assignedTeam field might be empty or incorrect
3. ❌ Need to manually update database to set correct team
4. ❌ Escalation not initialized
5. ❌ No SMS notifications

### After (Automatic Process)
1. ✅ Complaint created with issue_type = "bus_delay"
2. ✅ System automatically maps to "Operations & Scheduling"
3. ✅ Auto-assigns to "Operations Team" (Level 1)
4. ✅ Escalation initialized with due date
5. ✅ SMS sent to Operations Officer: 94771234570
6. ✅ Frontend shows correct category and team

## 🔧 Customizing Team Assignments

To change team names or phone numbers, edit: `backend/internal/config/assignment.go`

```go
"Operations & Scheduling": {
    1: {"Your Team Name", "Your Role", "94771234567"},
    2: {"Your Team Name", "Senior Role", "94771234568"},
    3: {"Your Team Name", "Manager Role", "94771234569"},
},
```

**Important**: Keep team names consistent across all levels for the same category!

## 📱 Example Flow

### Example 1: Bus Delay Complaint

```
1. Bus driver reports delay (issue_type: "bus_delay")
   ↓
2. System maps to "Operations & Scheduling"
   ↓
3. Auto-assigns to "Operations Team" (Level 1)
   ↓
4. SMS sent to Operations Officer (94771234570)
   ↓
5. Escalation due in 5 days
   ↓
6. If unresolved, auto-escalates to Level 2 → Operations Manager
```

### Example 2: Maintenance Issue

```
1. Report flat wheel (issue_type: "flat_wheel")
   ↓
2. System maps to "Vehicle & Facility"
   ↓
3. Auto-assigns to "Maintenance Team" (Level 1)
   ↓
4. SMS sent to Maintenance Technician (94771234573)
   ↓
5. Can be manually escalated to Maintenance Supervisor
```

## 🎨 UI Display

In the Complaint Management UI, you'll now see:

- **Category**: Properly formatted (e.g., "Operations & Scheduling")
- **Assigned Team**: Matching the category (e.g., "Operations Team")
- **Escalation Level**: Current level badge (1, 2, or 3)
- **Current Team**: Team at current escalation level

## ✨ Benefits

1. **No Database Editing**: Everything is automatic
2. **Consistent Naming**: Team names match across all complaints
3. **Proper Escalation**: Teams change as complaints escalate
4. **SMS Notifications**: Right person gets notified at each level
5. **Category Matching**: Issue types correctly map to categories
6. **Auto-Initialization**: Escalation starts automatically

## 🔍 Verification

To verify the matching is working:

1. **View Any Complaint** in the Complaint Management page
2. **Check the Category** - should be properly formatted
3. **Check Assigned Team** - should match the category
4. **Look for Escalation Info** - should show current level and team
5. **Try Manual Escalation** - team should change to next level

## 📋 Issue Type Examples

Here are common issue types and how they're categorized:

```
bus_delay              → Operations & Scheduling → Operations Team
maintenance_issue      → Vehicle & Facility     → Maintenance Team
flat_wheel            → Vehicle & Facility     → Maintenance Team
equipment_malfunction → Vehicle & Facility     → Maintenance Team
passenger_complaint   → Service Issue          → Customer Service
safety_concern        → Safety & Security      → Safety & Security
driver_behavior       → Other                  → General Support
route_change          → Other                  → General Support
```

## 🚀 Setup Checklist

- [x] Backend updated with auto-matching logic
- [x] Complaint service initialized with escalation support
- [x] Team assignments configured in assignment.go
- [x] SMS notification integrated
- [x] Frontend displays escalation info
- [x] Manual escalation works
- [x] Auto-escalation scheduled

## ✅ You're All Set!

The system now automatically handles category-to-team matching. No manual intervention needed!
