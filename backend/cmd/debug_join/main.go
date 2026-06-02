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

	query := `
		SELECT 
			l.id::text,
			COALESCE(lo.manager_full_name, 'NO_NAME'),
			l.lounge_owner_id::text
		FROM lounges l
		LEFT JOIN lounge_owners lo ON l.lounge_owner_id = lo.id
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Testing Join Query:")
	for rows.Next() {
		var id, ownerName, ownerID string
		if err := rows.Scan(&id, &ownerName, &ownerID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Lounge: %s, OwnerID: %s, OwnerName: %s\n", id, ownerID, ownerName)
	}
}
