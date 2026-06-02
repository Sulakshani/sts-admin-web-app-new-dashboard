-- Lounge Booking Tables Schema
-- This file contains the table definitions for lounge booking management

-- Enable UUID extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Bookings Table (Main booking table from bus system)
CREATE TABLE IF NOT EXISTS bookings (
    bus_booking_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    passenger_name TEXT NOT NULL,
    passenger_phone TEXT,
    booking_reference TEXT UNIQUE,
    scheduled_arrival TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
);

-- Lounge Bookings Table
CREATE TABLE IF NOT EXISTS lounge_bookings (
    lounge_booking_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    bus_booking_id UUID REFERENCES bookings(bus_booking_id),
    lounge_name TEXT NOT NULL,
    pricing_type TEXT, -- Duration field (e.g., "1 hour", "2 hours", "Half day")
    number_of_guests INT DEFAULT 1,
    selected_amenities JSONB DEFAULT '[]'::jsonb, -- Array of selected amenities
    booking_type TEXT, -- e.g., "Regular", "VIP", "Premium", "Corporate"
    total_amount DECIMAL(10, 2) DEFAULT 0,
    payment_status TEXT DEFAULT 'Pending', -- Pending, Paid, Failed
    status TEXT DEFAULT 'Pending', -- Confirmed, Pending, Cancelled, Completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW()),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
);

-- Lounge Booking Pre-Orders Table (for marketplace items)
CREATE TABLE IF NOT EXISTS lounge_booking_pre_orders (
    pre_order_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    lounge_booking_id UUID REFERENCES lounge_bookings(lounge_booking_id) ON DELETE CASCADE UNIQUE,
    product_name TEXT, -- Market place product name
    quantity INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_lounge_bookings_bus_booking ON lounge_bookings(bus_booking_id);
CREATE INDEX IF NOT EXISTS idx_lounge_bookings_payment_status ON lounge_bookings(payment_status);
CREATE INDEX IF NOT EXISTS idx_lounge_bookings_status ON lounge_bookings(status);
CREATE INDEX IF NOT EXISTS idx_lounge_bookings_created_at ON lounge_bookings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lounge_booking_pre_orders_booking ON lounge_booking_pre_orders(lounge_booking_id);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_lounge_booking_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = TIMEZONE('utc'::text, NOW());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_lounge_booking_timestamp
    BEFORE UPDATE ON lounge_bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_lounge_booking_timestamp();

-- Insert sample data for testing (optional)
-- INSERT INTO bookings (passenger_name, passenger_phone, booking_reference, scheduled_arrival)
-- VALUES 
--     ('John Doe', '+1234567890', 'REF-1001', NOW() + INTERVAL '2 hours'),
--     ('Jane Smith', '+0987654321', 'REF-1002', NOW() + INTERVAL '5 hours');

-- INSERT INTO lounge_bookings (bus_booking_id, lounge_name, pricing_type, number_of_guests, selected_amenities, booking_type, total_amount, payment_status, status)
-- SELECT 
--     bus_booking_id,
--     'Alpha Lounge',
--     '2 hours',
--     2,
--     '["WiFi", "Premium meals", "spa service"]'::jsonb,
--     'VIP',
--     150.00,
--     'Paid',
--     'Confirmed'
-- FROM bookings WHERE booking_reference = 'REF-1001';
