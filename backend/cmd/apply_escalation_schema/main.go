package main

import (
	"fmt"
	"log"
	"os"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
)

func main() {
	// Use the same config as the main application
	cfg := config.LoadConfig()
	database.Init(cfg)

	// Read SQL file
	sqlFile := "complaint_escalation_schema.sql"
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	sqlContent := string(sqlBytes)

	// Execute SQL
	log.Println("Applying complaint escalation schema...")
	_, err = database.DB.Exec(sqlContent)
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	log.Println("✅ Complaint escalation schema applied successfully!")

	// Verify tables were created
	var tableName string
	err = database.DB.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'complaint_escalations' AND table_schema = 'public'").Scan(&tableName)
	if err != nil {
		log.Printf("Warning: Could not verify table creation: %v", err)
	} else {
		fmt.Printf("✅ Table 'complaint_escalations' created\n")
	}

	err = database.DB.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'complaint_escalation_history' AND table_schema = 'public'").Scan(&tableName)
	if err != nil {
		log.Printf("Warning: Could not verify table creation: %v", err)
	} else {
		fmt.Printf("✅ Table 'complaint_escalation_history' created\n")
	}

	log.Println("Migration complete!")
}
