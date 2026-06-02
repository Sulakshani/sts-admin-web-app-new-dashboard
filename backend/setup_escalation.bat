@echo off
REM Quick Setup Script for Complaint Escalation System
REM Run this in PowerShell: .\setup_escalation.bat

echo ========================================
echo Complaint Escalation Setup
echo ========================================
echo.

echo Step 1: Checking if escalation schema exists...
if not exist "complaint_escalation_schema.sql" (
    echo ERROR: complaint_escalation_schema.sql not found!
    echo Please make sure you're running this from the backend folder.
    pause
    exit /b 1
)

echo Step 2: Applying escalation schema to database...
echo Please enter your database connection details:
echo.
set /p DB_USER="PostgreSQL Username: "
set /p DB_NAME="Database Name: "
set /p DB_HOST="Host (default: localhost): "
if "%DB_HOST%"=="" set DB_HOST=localhost
set /p DB_PORT="Port (default: 5432): "
if "%DB_PORT%"=="" set DB_PORT=5432

echo.
echo Running schema...
psql -U %DB_USER% -h %DB_HOST% -p %DB_PORT% -d %DB_NAME% -f complaint_escalation_schema.sql

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Failed to apply schema. Please check your database credentials.
    pause
    exit /b 1
)

echo.
echo ✅ Schema applied successfully!
echo.

echo Step 3: Running diagnostic check...
go run cmd/check_escalation/main.go

echo.
echo ========================================
echo Setup Complete!
echo ========================================
echo.
echo Next steps:
echo 1. Configure SMS in .env file (optional):
echo    ESMS_API_KEY=your_key_here
echo    ESMS_SENDER_ID=BusLounge
echo.
echo 2. Update team phone numbers in:
echo    internal/config/assignment.go
echo.
echo 3. Restart your backend server:
echo    go run cmd/server/main.go
echo.
echo 4. Test in UI by viewing any complaint
echo.
pause
