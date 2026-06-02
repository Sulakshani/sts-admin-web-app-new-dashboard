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

	tables := []string{"lounge_owners", "lounge_marketplace_categories"}

	for _, table := range tables {
		fmt.Printf("\nColumns in %s table:\n", table)
		rows, err := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = $1", table)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var name, dtype string
			if err := rows.Scan(&name, &dtype); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s (%s)\n", name, dtype)
		}
		rows.Close()
	}
}
