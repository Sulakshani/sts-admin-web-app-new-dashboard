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
    
    query2 := `SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = ''bus_staff_employment''::regclass AND contype = ''c''`
    rows2, err := db.Query(query2)
    if err != nil { log.Fatal(err) }
    defer rows2.Close()
    
    fmt.Println("Check constraints:")
    for rows2.Next() {
        var name, def string
        rows2.Scan(&name, &def)
        fmt.Printf("%s\n%s\n\n", name, def)
    }
}
