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

	// Query to show bus with both route_permits.status and bus_owners.verification_status
	query := `
		SELECT 
			b.id::text as bus_id,
			b.bus_number,
			bo.company_name,
			bo.id::text as bus_owner_id,
			rp.id::text as permit_id,
			rp.status::text as route_permit_status,
			bo.verification_status::text as bus_owner_verification_status
		FROM buses b
		LEFT JOIN bus_owners bo ON b.bus_owner_id = bo.id
		LEFT JOIN route_permits rp ON b.permit_id = rp.id
		ORDER BY b.created_at DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n=== BUS VERIFICATION STATUS COMPARISON ===")
	fmt.Println("Showing: route_permits.status vs bus_owners.verification_status")
	fmt.Println("==================================================")

	for rows.Next() {
		var busID, busNumber, companyName, busOwnerID, permitID, permitStatus, ownerStatus string
		
		err := rows.Scan(&busID, &busNumber, &companyName, &busOwnerID, &permitID, &permitStatus, &ownerStatus)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Show if they match or not
		match := "✓ SYNCED"
		if permitStatus != ownerStatus {
			match = "✗ NOT SYNCED!"
		}

		fmt.Printf("\nBus: %s (%s)\n", busNumber, companyName)
		fmt.Printf("  Bus Owner ID: %s\n", busOwnerID)
		fmt.Printf("  Permit ID: %s\n", permitID)
		fmt.Printf("  route_permits.status: %s\n", permitStatus)
		fmt.Printf("  bus_owners.verification_status: %s\n", ownerStatus)
		fmt.Printf("  Status: %s\n", match)
		fmt.Println("--------------------------------------------------")
	}

	fmt.Println("\n✓ Check complete!")
	fmt.Println("If statuses don't match, approve the bus again from notification panel.")
	fmt.Println("==================================================\n")
}
