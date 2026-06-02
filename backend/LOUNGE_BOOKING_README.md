# Lounge Booking Management Backend

This directory contains the backend implementation for the Lounge Booking Management system.

## Overview

The backend provides RESTful APIs for managing lounge bookings with data connected from the main bookings table and lounge-specific tables.

## Database Schema

### Tables

1. **bookings** - Main booking table from bus system
   - `bus_booking_id` (UUID, Primary Key)
   - `passenger_name` (TEXT)
   - `passenger_phone` (TEXT)
   - `booking_reference` (TEXT, Unique)
   - `scheduled_arrival` (TIMESTAMP)
   - `created_at` (TIMESTAMP)

2. **lounge_bookings** - Lounge-specific booking information
   - `lounge_booking_id` (UUID, Primary Key)
   - `bus_booking_id` (UUID, Foreign Key → bookings)
   - `lounge_name` (TEXT)
   - `pricing_type` (TEXT) - Duration field
   - `number_of_guests` (INT)
   - `selected_amenities` (JSONB) - Array of amenities
   - `booking_type` (TEXT) - e.g., Regular, VIP, Premium, Corporate
   - `total_amount` (DECIMAL)
   - `payment_status` (TEXT) - Pending, Paid, Failed
   - `status` (TEXT) - Confirmed, Pending, Cancelled, Completed
   - `created_at`, `updated_at` (TIMESTAMP)

3. **lounge_booking_pre_orders** - Marketplace pre-orders
   - `pre_order_id` (UUID, Primary Key)
   - `lounge_booking_id` (UUID, Foreign Key → lounge_bookings)
   - `product_name` (TEXT) - Market place product
   - `quantity` (INT)
   - `created_at` (TIMESTAMP)

### Database Setup

Run the schema file to create all necessary tables:

```bash
psql -U your_user -d your_database -f lounge_booking_schema.sql
```

## API Endpoints

### Base URL
```
http://localhost:8080/api
```

### Endpoints

#### 1. Get All Lounge Bookings
```
GET /lounge-bookings
```
Returns all lounge bookings with joined data from bookings and pre-orders tables.

**Response:**
```json
[
  {
    "lounge_booking_id": "uuid",
    "bus_booking_id": "uuid",
    "passenger_name": "John Doe",
    "passenger_phone": "+1234567890",
    "booking_reference": "REF-1001",
    "lounge_name": "Alpha Lounge",
    "scheduled_arrival": "2025-09-02T10:00:00Z",
    "pricing_type": "2 hours",
    "number_of_guests": 2,
    "selected_amenities": ["WiFi", "Premium meals", "spa service"],
    "product_name": "VIP Package",
    "booking_type": "VIP",
    "total_amount": 150.00,
    "payment_status": "Paid",
    "status": "Confirmed",
    "created_at": "2025-09-01T09:00:00Z"
  }
]
```

#### 2. Get Lounge Booking by ID
```
GET /lounge-bookings/:id
```

#### 3. Create Lounge Booking
```
POST /lounge-bookings
Content-Type: application/json

{
  "bus_booking_id": "uuid",
  "lounge_name": "Alpha Lounge",
  "pricing_type": "2 hours",
  "number_of_guests": 2,
  "selected_amenities": ["WiFi", "Premium meals"],
  "product_name": "VIP Package",
  "booking_type": "VIP",
  "total_amount": 150.00,
  "payment_status": "Pending",
  "status": "Pending"
}
```

#### 4. Update Lounge Booking
```
PUT /lounge-bookings/:id
Content-Type: application/json

{
  "lounge_name": "Beta Lounge",
  "pricing_type": "3 hours",
  "number_of_guests": 3,
  "selected_amenities": ["WiFi", "Shower", "Meal"],
  "product_name": "Premium Package",
  "booking_type": "Premium",
  "total_amount": 200.00,
  "payment_status": "Paid",
  "status": "Confirmed"
}
```

#### 5. Update Payment Status
```
PATCH /lounge-bookings/:id/payment-status
Content-Type: application/json

{
  "status": "Paid"
}
```
Valid values: `Pending`, `Paid`, `Failed`

#### 6. Update Booking Status
```
PATCH /lounge-bookings/:id/booking-status
Content-Type: application/json

{
  "status": "Confirmed"
}
```
Valid values: `Confirmed`, `Pending`, `Cancelled`, `Completed`

#### 7. Delete Lounge Booking
```
DELETE /lounge-bookings/:id
```

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    # Main server file with routes
├── internal/
│   ├── models/
│   │   └── lounge_booking.go         # LoungeBooking model
│   ├── database/
│   │   └── lounge_booking_repository.go  # Database operations
│   ├── services/
│   │   └── lounge_booking_service.go     # Business logic layer
│   └── handlers/
│       └── lounge_booking_handler.go     # HTTP handlers
├── lounge_booking_schema.sql         # Database schema
└── LOUNGE_BOOKING_README.md          # This file
```

## Data Mapping

The API combines data from multiple tables:

| Frontend Field | Database Source |
|---------------|-----------------|
| L_Booking id | lounge_bookings.lounge_booking_id |
| Passenger Name | bookings.passenger_name (via bus_booking_id) |
| Passenger Phone | bookings.passenger_phone (via bus_booking_id) |
| Ref NUM | bookings.booking_reference |
| Lounge Name | lounge_bookings.lounge_name |
| Date & Time | bookings.scheduled_arrival |
| Duration | lounge_bookings.pricing_type |
| No of Guests | lounge_bookings.number_of_guests |
| Additional features | lounge_bookings.selected_amenities (JSONB array) |
| Market place | lounge_booking_pre_orders.product_name |
| Booking Type | lounge_bookings.booking_type |
| Total Amount | lounge_bookings.total_amount |
| Payment Status | lounge_bookings.payment_status |
| Booking Status | lounge_bookings.status |

## Features

- ✅ Full CRUD operations for lounge bookings
- ✅ Join queries across bookings, lounge_bookings, and lounge_booking_pre_orders tables
- ✅ Status management (Payment Status & Booking Status)
- ✅ JSONB support for amenities array
- ✅ Automatic timestamp tracking (created_at, updated_at)
- ✅ Data validation and error handling
- ✅ RESTful API design
- ✅ CORS enabled

## Running the Server

```bash
cd backend
go run cmd/server/main.go
```

Server will start on port `8080` (default).

## Testing

You can test the APIs using:
- Postman
- cURL
- Frontend application

Example cURL request:
```bash
curl -X GET http://localhost:8080/api/lounge-bookings
```

## Notes

- All timestamps are in UTC
- UUIDs are auto-generated for new bookings
- Selected amenities should be sent as a JSON array
- Payment status defaults to "Pending" if not specified
- Booking status defaults to "Pending" if not specified
