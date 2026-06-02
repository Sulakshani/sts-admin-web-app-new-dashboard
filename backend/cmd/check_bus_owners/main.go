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
		log.Fatal(err)
	}
	defer db.Close()

	// Check if bus_owners table exists
	var exists bool
	err = db.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'bus_owners')`).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bus_owners table exists: %v\n\n", exists)

	if !exists {
		fmt.Println("bus_owners table does not exist!")
		return
	}

	// Count bus owners
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM bus_owners`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total bus owners in database: %d\n\n", count)

	// Get all bus owners
	rows, err := db.Query(`
		SELECT 
			id,
			COALESCE(user_id::text, 'NULL') as user_id,
			COALESCE(company_name, 'NULL') as company_name,
			COALESCE(business_email, 'NULL') as business_email,
			COALESCE(business_phone, 'NULL') as business_phone,
			COALESCE(identity_or_incorporation_no, 'NULL') as identity_or_incorporation_no,
			COALESCE(verification_status::text, 'NULL') as verification_status
		FROM bus_owners
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Bus Owners (first 10):")
	fmt.Println("----------------------------------------")
	for rows.Next() {
		var id, userID, company, email, phone, nic, status string
		rows.Scan(&id, &userID, &company, &email, &phone, &nic, &status)
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("  Company: %s\n", company)
		fmt.Printf("  Email: %s\n", email)
		fmt.Printf("  Phone: %s\n", phone)
		fmt.Printf("  NIC: %s\n", nic)
		fmt.Printf("  Status: %s\n", status)
		fmt.Println("----------------------------------------")
	}
}
