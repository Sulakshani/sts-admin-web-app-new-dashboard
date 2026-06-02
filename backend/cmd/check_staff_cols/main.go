package main
import (
    "database/sql"
    "fmt"
    "log"
    "os"
    _ "github.com/lib/pq"
)
func main() {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil { log.Fatal(err) }
    defer db.Close()
    
    query := `SELECT column_name, data_type, is_nullable 
              FROM information_schema.columns 
              WHERE table_name = 'bus_staff' 
              ORDER BY ordinal_position`
    
    rows, err := db.Query(query)
    if err != nil { log.Fatal(err) }
    defer rows.Close()
    
    fmt.Println("bus_staff table columns:")
    for rows.Next() {
        var col, dtype, nullable string
        rows.Scan(&col, &dtype, &nullable)
        fmt.Printf("  %s (%s) - Nullable: %s\n", col, dtype, nullable)
    }
}
