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

	tables := []string{"bus_owners", "route_permits", "master_routes"}

	for _, table := range tables {
		fmt.Printf("\nConstraints for %s table:\n", table)
		rows, err := db.Query(`
			SELECT column_name, is_nullable 
			FROM information_schema.columns 
			WHERE table_name = $1
		`, table)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var name, nullable string
			if err := rows.Scan(&name, &nullable); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s: %s\n", name, nullable)
		}
		rows.Close()
	}
}
