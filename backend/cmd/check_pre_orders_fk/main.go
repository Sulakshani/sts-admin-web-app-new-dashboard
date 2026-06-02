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

	// Get all foreign key constraints on lounge_booking_pre_orders table
	rows, err := db.Query(`
		SELECT
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = 'lounge_booking_pre_orders'
	`)
	if err != nil {
		fmt.Println("Error querying foreign keys:", err)
		return
	}
	defer rows.Close()

	fmt.Println("Foreign key constraints on lounge_booking_pre_orders:")
	for rows.Next() {
		var constraintName, tableName, columnName, foreignTableName, foreignColumnName string
		if err := rows.Scan(&constraintName, &tableName, &columnName, &foreignTableName, &foreignColumnName); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  - %s: %s(%s) -> %s(%s)\n", constraintName, tableName, columnName, foreignTableName, foreignColumnName)
	}
}
