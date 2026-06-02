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

	// Check if users table exists and has data
	var count int
	err = db.QueryRow("SELECT count(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Error querying users table: %v", err)
		// Try to list tables to see if 'users' exists or if it's in a schema like 'auth'
		rows, _ := db.Query("SELECT table_schema, table_name FROM information_schema.tables WHERE table_name = 'users'")
		defer rows.Close()
		for rows.Next() {
			var schema, name string
			rows.Scan(&schema, &name)
			fmt.Printf("Found table: %s.%s\n", schema, name)
		}
		return
	}
	fmt.Printf("User count: %d\n", count)

	if count > 0 {
		rows, err := db.Query("SELECT id, email FROM users LIMIT 5")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, email string
			rows.Scan(&id, &email)
			fmt.Printf("User: %s (%s)\n", id, email)
		}
	}
}
