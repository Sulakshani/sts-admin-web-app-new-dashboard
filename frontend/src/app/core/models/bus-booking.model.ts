export interface BusBooking {
  booking_id: string;
  scheduled_trip_id: string;
  bus_id: string;
  passenger_name: string;
  passenger_phone: string;
  booking_reference: string;
  route: string;
  departure_datetime: string;
  bus_type: string;
  seat_number: string;
  total_fare: number;
  payment_status: string;
  booking_status: string;
  created_at: string;
  bus_number: string;
  license_plate: string;
  number_of_seats: number;
}