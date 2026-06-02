package main
import ("database/sql"; "fmt"; "os"; _ "github.com/lib/pq")
func main() {
db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
defer db.Close()
var busNum string
query := `SELECT 
CASE 
WHEN bus.bus_number IS NOT NULL AND bus.bus_number != '' THEN bus.bus_number
WHEN rp.permit_number IS NOT NULL AND rp.permit_number != '' THEN rp.permit_number
ELSE 'BUS-' || SUBSTRING(st.id::text, 1, 8)
END as bus_number
FROM bus_bookings bb
INNER JOIN bookings b ON bb.booking_id = b.id
INNER JOIN scheduled_trips st ON bb.scheduled_trip_id = st.id
LEFT JOIN bus_owner_routes bor ON st.bus_owner_route_id = bor.id
LEFT JOIN route_permits rp ON st.permit_id = rp.id
LEFT JOIN master_routes mr ON rp.master_route_id = mr.id
LEFT JOIN buses bus ON st.permit_id = bus.permit_id
ORDER BY bb.created_at DESC
LIMIT 5`
rows, _ := db.Query(query)
defer rows.Close()
i := 1
for rows.Next() {
rows.Scan(&busNum)
fmt.Printf("%d. '%s'\n", i, busNum)
i++
}
}
