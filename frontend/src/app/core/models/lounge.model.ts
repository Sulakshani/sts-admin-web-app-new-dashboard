export interface Lounge {
  lounge_id: string;
  lounge_owner: string;
  owner_nic?: string;
  owner_email?: string;
  owner_contact?: string;
  lounge_name: string;
  lounge_contact: string;
  address: string;
  capacity: number;
  price_per_hour: number;
  facilities: string[];
  marketplace: string;
  verification: string;
  verification_note: string;
  operational: boolean;
}


