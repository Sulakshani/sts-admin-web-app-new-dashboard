package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func checkBusDataMain() {
	db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()

	// Check total buses
	var totalBuses int
	db.QueryRow("SELECT COUNT(*) FROM buses").Scan(&totalBuses)
	fmt.Printf("Total buses in table: %d\n\n", totalBuses)

	// Check buses with bus_number
	var withBusNumber int
	db.QueryRow("SELECT COUNT(*) FROM buses WHERE bus_number IS NOT NULL AND bus_number != ''").Scan(&withBusNumber)
	fmt.Printf("Buses with bus_number: %d\n", withBusNumber)

	// Show sample bus_numbers
	if withBusNumber > 0 {
		fmt.Println("\nSample bus_numbers:")
		rows, _ := db.Query("SELECT id::text, bus_number, permit_id::text FROM buses WHERE bus_number IS NOT NULL AND bus_number != '' LIMIT 5")
		defer rows.Close()
		for rows.Next() {
			var id, busNum, permitID string
			rows.Scan(&id, &busNum, &permitID)
			fmt.Printf("  ID: %s | Bus#: %s | Permit: %s\n", id[:8]+"...", busNum, permitID[:8]+"...")
		}
	}

	// Check how many scheduled_trips link to buses with bus_number
	var tripsWithBusNum int
	db.QueryRow("SELECT COUNT(DISTINCT st.id) FROM scheduled_trips st LEFT JOIN buses b ON st.permit_id = b.permit_id WHERE b.bus_number IS NOT NULL AND b.bus_number != ''").Scan(&tripsWithBusNum)
	fmt.Printf("\nScheduled trips linked to buses with bus_number: %d\n", tripsWithBusNum)
}
