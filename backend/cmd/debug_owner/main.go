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

	// Check a specific lounge owner ID from the previous debug output
	targetID := "6dfc1063-e3cc-4524-9b98-97cfe258c89f"

	var name string
	err = db.QueryRow("SELECT manager_full_name FROM lounge_owners WHERE id = $1", targetID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("Owner with ID %s NOT FOUND in lounge_owners table.\n", targetID)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Owner found: %s\n", name)
	}
}
