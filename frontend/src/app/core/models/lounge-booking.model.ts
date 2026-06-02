export interface LoungeBooking {
  lounge_booking_id: string; // id from lounge_bookings table
  bus_booking_id: string | null; // connection to bookings table (nullable)
  passenger_name?: string; // from bookings table
  passenger_phone?: string; // from bookings table
  booking_reference?: string; // ref_num from bookings table
  lounge_name: string; // from lounge_bookings table
  scheduled_arrival: string; // Date & Time from bookings table (ISO string)
  pricing_type: string; // Duration from lounge_bookings table
  number_of_guests: number; // from lounge_bookings table
  selected_amenities?: string[]; // Additional features from lounge_bookings table
  product_name?: string; // Market place from lounge_booking_pre_orders table
  booking_type?: string; // from lounge_bookings table
  total_amount: number; // from lounge_bookings table
  payment_status: 'pending' | 'paid' | 'failed'; // from lounge_bookings table
  status: 'confirmed' | 'pending' | 'cancelled' | 'completed'; // Booking Status from lounge_bookings table
  created_at: string; // ISO
}