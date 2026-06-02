-- Insert sample data for testing lounge booking management

-- Insert sample bookings
INSERT INTO bookings (passenger_name, passenger_phone, booking_reference, scheduled_arrival)
VALUES 
    ('John Doe', '+1234567890', 'REF-1001', NOW() + INTERVAL '2 hours'),
    ('Jane Smith', '+0987654321', 'REF-1002', NOW() + INTERVAL '5 hours'),
    ('Alice Johnson', '+1122334455', 'REF-1003', NOW() + INTERVAL '1 day')
ON CONFLICT (booking_reference) DO NOTHING;

-- Insert sample lounge bookings
WITH booking_ids AS (
    SELECT bus_booking_id, booking_reference FROM bookings 
    WHERE booking_reference IN ('REF-1001', 'REF-1002', 'REF-1003')
)
INSERT INTO lounge_bookings (
    bus_booking_id, 
    lounge_name, 
    pricing_type, 
    number_of_guests, 
    selected_amenities, 
    booking_type, 
    total_amount, 
    payment_status, 
    status
)
SELECT 
    b.bus_booking_id,
    'Alpha Lounge',
    '2 hours',
    2,
    '["WiFi", "Premium meals", "spa service"]'::jsonb,
    'VIP',
    150.00,
    'Paid',
    'Confirmed'
FROM booking_ids b WHERE b.booking_reference = 'REF-1001'
UNION ALL
SELECT 
    b.bus_booking_id,
    'Beta Premium Lounge',
    '3 hours',
    4,
    '["Shower", "Meal", "Airport transfer"]'::jsonb,
    'Premium',
    200.00,
    'Pending',
    'Confirmed'
FROM booking_ids b WHERE b.booking_reference = 'REF-1002'
UNION ALL
SELECT 
    b.bus_booking_id,
    'Gamma Relax Lounge',
    '1 hour',
    1,
    '["WiFi", "cargo storage"]'::jsonb,
    'Regular',
    50.00,
    'Failed',
    'Cancelled'
FROM booking_ids b WHERE b.booking_reference = 'REF-1003';

-- Insert sample pre-orders (marketplace items)
WITH lounge_booking_ids AS (
    SELECT lb.lounge_booking_id, lb.lounge_name 
    FROM lounge_bookings lb
    WHERE lb.lounge_name IN ('Alpha Lounge', 'Beta Premium Lounge')
)
INSERT INTO lounge_booking_pre_orders (lounge_booking_id, product_name, quantity)
SELECT lounge_booking_id, 'VIP Package', 1
FROM lounge_booking_ids WHERE lounge_name = 'Alpha Lounge'
ON CONFLICT (lounge_booking_id) DO NOTHING
UNION ALL
SELECT lounge_booking_id, 'Premium Meals Package', 2
FROM lounge_booking_ids WHERE lounge_name = 'Beta Premium Lounge'
ON CONFLICT (lounge_booking_id) DO NOTHING;

-- Verify the data
SELECT COUNT(*) as total_bookings FROM bookings;
SELECT COUNT(*) as total_lounge_bookings FROM lounge_bookings;
SELECT COUNT(*) as total_pre_orders FROM lounge_booking_pre_orders;
