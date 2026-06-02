export interface Bus {
  id: string;
  bus_number: string;
  company_name: string;
  identify_or_incorporation_no: string;
  business_email: string;
  business_phone: string;
  permit_number: string;
  license_plate: string;
  total_seats: number;
  bus_type: string;
  custom_route_name: string;
  fare_per_seat: number;
  status: string;
  verification_status: 'Verified' | 'Pending' | 'Rejected';
  owner_verification_status?: 'Verified' | 'Pending' | 'Rejected';
  verification_documents?: string[]; // Array of document URLs or file names
}
