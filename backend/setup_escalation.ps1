# Quick Setup Script for Complaint Escalation System
# Run this in PowerShell: .\setup_escalation.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Complaint Escalation Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check schema file
Write-Host "Step 1: Checking if escalation schema exists..." -ForegroundColor Yellow
if (-not (Test-Path "complaint_escalation_schema.sql")) {
    Write-Host "❌ ERROR: complaint_escalation_schema.sql not found!" -ForegroundColor Red
    Write-Host "Please make sure you're running this from the backend folder." -ForegroundColor Red
    pause
    exit 1
}
Write-Host "✅ Schema file found" -ForegroundColor Green
Write-Host ""

# Step 2: Get database credentials
Write-Host "Step 2: Database Configuration" -ForegroundColor Yellow
$DB_USER = Read-Host "PostgreSQL Username"
$DB_NAME = Read-Host "Database Name"
$DB_HOST = Read-Host "Host (press Enter for localhost)"
if ([string]::IsNullOrWhiteSpace($DB_HOST)) { $DB_HOST = "localhost" }
$DB_PORT = Read-Host "Port (press Enter for 5432)"
if ([string]::IsNullOrWhiteSpace($DB_PORT)) { $DB_PORT = "5432" }

Write-Host ""
Write-Host "Applying escalation schema..." -ForegroundColor Yellow

# Run psql
$env:PGPASSWORD = Read-Host "PostgreSQL Password" -AsSecureString | ConvertFrom-SecureString -AsPlainText
$psqlCommand = "psql -U $DB_USER -h $DB_HOST -p $DB_PORT -d $DB_NAME -f complaint_escalation_schema.sql"
Invoke-Expression $psqlCommand

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ ERROR: Failed to apply schema." -ForegroundColor Red
    Write-Host "Please check your database credentials and try again." -ForegroundColor Red
    pause
    exit 1
}

Write-Host ""
Write-Host "✅ Schema applied successfully!" -ForegroundColor Green
Write-Host ""

# Step 3: Run diagnostic
Write-Host "Step 3: Running diagnostic check..." -ForegroundColor Yellow
go run cmd/check_escalation/main.go

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Configure SMS in .env file (optional):" -ForegroundColor White
Write-Host "   ESMS_API_KEY=your_key_here"
Write-Host "   ESMS_SENDER_ID=BusLounge"
Write-Host ""
Write-Host "2. Update team phone numbers in:" -ForegroundColor White
Write-Host "   internal/config/assignment.go"
Write-Host ""
Write-Host "3. Restart your backend server:" -ForegroundColor White
Write-Host "   go run cmd/server/main.go"
Write-Host ""
Write-Host "4. Test in UI by viewing any complaint" -ForegroundColor White
Write-Host ""
pause
