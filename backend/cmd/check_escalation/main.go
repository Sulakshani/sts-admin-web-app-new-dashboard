package main

import (
	"database/sql"
	"fmt"
	"log"
	"sts-backend/internal/config"
	"sts-backend/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔍 Escalation System Diagnostic Tool")
	fmt.Println("====================================\n")

	// Load config
	cfg := config.LoadConfig()
	database.Init(cfg)

	// Test database connection
	fmt.Println("1. Testing database connection...")
	err := database.DB.Ping()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	fmt.Println("✅ Database connected successfully\n")

	// Check if escalation tables exist
	fmt.Println("2. Checking escalation tables...")
	if !tableExists(database.DB, "complaint_escalations") {
		fmt.Println("❌ Table 'complaint_escalations' does NOT exist")
		fmt.Println("   → Run: psql -U your_user -d your_db -f complaint_escalation_schema.sql")
		return
	}
	fmt.Println("✅ Table 'complaint_escalations' exists")

	if !tableExists(database.DB, "complaint_escalation_history") {
		fmt.Println("❌ Table 'complaint_escalation_history' does NOT exist")
		fmt.Println("   → Run: psql -U your_user -d your_db -f complaint_escalation_schema.sql")
		return
	}
	fmt.Println("✅ Table 'complaint_escalation_history' exists\n")

	// Count escalations
	fmt.Println("3. Checking existing escalations...")
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM complaint_escalations").Scan(&count)
	if err != nil {
		log.Printf("❌ Error counting escalations: %v", err)
	} else {
		fmt.Printf("✅ Found %d escalation records\n\n", count)
	}

	// Count complaints without escalation
	fmt.Println("4. Checking complaints without escalation...")
	var missingCount int
	query := `
		SELECT COUNT(*) 
		FROM report_issues ri
		WHERE NOT EXISTS (
			SELECT 1 FROM complaint_escalations ce 
			WHERE ce.complaint_id = ri.id
		)
		AND ri.status NOT IN ('resolved', 'closed')
	`
	err = database.DB.QueryRow(query).Scan(&missingCount)
	if err != nil {
		log.Printf("❌ Error counting missing escalations: %v", err)
	} else {
		if missingCount > 0 {
			fmt.Printf("⚠️  Found %d active complaints without escalation\n", missingCount)
			fmt.Println("   These will be auto-initialized when viewed in the UI")
		} else {
			fmt.Println("✅ All active complaints have escalation tracking")
		}
	}
	fmt.Println()

	// Check SMS configuration
	fmt.Println("5. Checking SMS configuration...")
	if cfg.ESMSAPIKey == "" {
		fmt.Println("⚠️  ESMS_API_KEY not configured (SMS notifications disabled)")
		fmt.Println("   → Add ESMS_API_KEY to your .env file")
	} else {
		fmt.Println("✅ ESMS_API_KEY configured")
	}
	
	if cfg.ESMSSenderID == "" {
		fmt.Println("⚠️  ESMS_SENDER_ID not configured")
	} else {
		fmt.Printf("✅ ESMS_SENDER_ID: %s\n", cfg.ESMSSenderID)
	}
	fmt.Println()

	// Test escalation assignment
	fmt.Println("6. Testing escalation assignment configuration...")
	categories := []string{"Service Issue", "Operations & Scheduling", "Vehicle & Facility", "Safety & Security", "Other"}
	allOk := true
	for _, category := range categories {
		assignment := config.GetAssignmentForCategory(category, 1)
		if assignment == nil {
			fmt.Printf("❌ No assignment for category: %s (Level 1)\n", category)
			allOk = false
		} else {
			fmt.Printf("✅ %s → %s (%s)\n", category, assignment.TeamName, assignment.PhoneNumber)
		}
	}
	if allOk {
		fmt.Println("\n✅ All category assignments configured correctly")
	}
	fmt.Println()

	// Sample recent complaints
	fmt.Println("7. Checking recent complaints...")
	sampleQuery := `
		SELECT 
			ri.id,
			ri.issue_type,
			ri.status,
			CASE 
				WHEN ce.complaint_id IS NOT NULL THEN 'YES'
				ELSE 'NO'
			END as has_escalation,
			COALESCE(ce.current_level::text, 'N/A') as level,
			COALESCE(ce.current_team, 'N/A') as team
		FROM report_issues ri
		LEFT JOIN complaint_escalations ce ON ri.id = ce.complaint_id
		WHERE ri.status NOT IN ('resolved', 'closed')
		ORDER BY ri.created_at DESC
		LIMIT 5
	`

	rows, err := database.DB.Query(sampleQuery)
	if err != nil {
		log.Printf("❌ Error querying complaints: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("\nRecent Active Complaints:")
		fmt.Println("ID | Issue Type | Status | Has Escalation | Level | Team")
		fmt.Println("---+-----------+--------+----------------+-------+-----")
		
		hasData := false
		for rows.Next() {
			hasData = true
			var id, issueType, status, hasEscalation, level, team string
			if err := rows.Scan(&id, &issueType, &status, &hasEscalation, &level, &team); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			shortID := id
			if len(id) > 8 {
				shortID = id[:8] + "..."
			}
			fmt.Printf("%s | %s | %s | %s | %s | %s\n", 
				shortID, issueType, status, hasEscalation, level, team)
		}
		
		if !hasData {
			fmt.Println("(No active complaints found)")
		}
	}

	fmt.Println("\n====================================")
	fmt.Println("✅ Diagnostic complete!")
	fmt.Println("\nIf you see errors above:")
	fmt.Println("1. Run complaint_escalation_schema.sql to create tables")
	fmt.Println("2. Configure SMS credentials in .env file")
	fmt.Println("3. Restart the server")
	fmt.Println("4. View complaints in the UI to auto-initialize escalation")
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		log.Printf("Error checking table %s: %v", tableName, err)
		return false
	}
	return exists
}
