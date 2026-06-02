package main

import (
	"database/sql"
	"fmt"
	"log"
	"sts-backend/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check the specific bus_owner that was just updated
	busOwnerID := "1f83d849-12d9-448a-9679-569157cf74c0"
	
	var id, companyName, verificationStatus string
	query := `SELECT id, company_name, verification_status FROM bus_owners WHERE id = $1`
	
	err = db.QueryRow(query, busOwnerID).Scan(&id, &companyName, &verificationStatus)
	if err != nil {
		log.Fatalf("Failed to query bus_owner: %v", err)
	}

	fmt.Println("\n=== BUS OWNER VERIFICATION STATUS ===")
	fmt.Printf("ID: %s\n", id)
	fmt.Printf("Company: %s\n", companyName)
	fmt.Printf("Verification Status: '%s'\n", verificationStatus)
	fmt.Println("=====================================\n")

	// Also check all bus_owners with their verification_status
	fmt.Println("=== ALL BUS OWNERS ===")
	rows, err := db.Query(`SELECT id, company_name, COALESCE(verification_status::text, 'NULL') FROM bus_owners ORDER BY company_name LIMIT 10`)
	if err != nil {
		log.Fatalf("Failed to query all bus_owners: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, company, status string
		if err := rows.Scan(&id, &company, &status); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("- %s: %s (status: %s)\n", company, id, status)
	}
	fmt.Println("=====================")
}
