package main

import (
	"fmt"
	"log"
	"sts-backend/internal/config"
	"sts-backend/internal/database"
)

func main() {
	cfg := config.LoadConfig()
	database.Init(cfg)

	// Check if report_issues table has data
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM report_issues").Scan(&count)
	if err != nil {
		log.Fatalf("Error querying report_issues: %v", err)
	}
	
	fmt.Printf("Total complaints in report_issues table: %d\n", count)

	// Check if there are users with driver/conductor roles
	var driverCount, conductorCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE 'driver' = ANY(roles)").Scan(&driverCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE 'conductor' = ANY(roles)").Scan(&conductorCount)
	
	fmt.Printf("Users with driver role: %d\n", driverCount)
	fmt.Printf("Users with conductor role: %d\n", conductorCount)

	// Check complaints from drivers/conductors
	var driverComplaintCount, conductorComplaintCount int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM report_issues ri 
		JOIN users u ON ri.reported_by_id = u.id 
		WHERE 'driver' = ANY(u.roles)
	`).Scan(&driverComplaintCount)
	
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM report_issues ri 
		JOIN users u ON ri.reported_by_id = u.id 
		WHERE 'conductor' = ANY(u.roles)
	`).Scan(&conductorComplaintCount)
	
	fmt.Printf("Complaints from drivers: %d\n", driverComplaintCount)
	fmt.Printf("Complaints from conductors: %d\n", conductorComplaintCount)
}
