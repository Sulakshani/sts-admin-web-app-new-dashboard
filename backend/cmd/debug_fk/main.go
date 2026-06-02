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

	rows, err := db.Query("SELECT id, owner_id, lounge_owner_id FROM lounges LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Checking foreign keys in lounges table:")
	for rows.Next() {
		var id string
		var ownerID, loungeOwnerID sql.NullString
		if err := rows.Scan(&id, &ownerID, &loungeOwnerID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Lounge ID: %s, owner_id: %v, lounge_owner_id: %v\n", id, ownerID.String, loungeOwnerID.String)
	}
}
