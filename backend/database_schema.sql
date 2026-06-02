-- Enable UUID extension
create extension if not exists "uuid-ossp";

-- Lounges Table
create table if not exists lounges (
  lounge_id uuid default uuid_generate_v4() primary key,
  owner text,
  owner_nic text,
  owner_email text,
  owner_contact text,
  name text,
  address text,
  lounge_contact text,
  capacity int,
  price_per_hour float,
  lounge_status text default 'open',
  operating_hours text,
  amenities text[],
  services text[],
  images text[],
  created_at timestamp with time zone default timezone('utc'::text, now()),
  verification text default 'Pending', -- Pending, Verified, Rejected
  verification_note text
);

-- Buses Table
create table if not exists buses (
  bus_id uuid default uuid_generate_v4() primary key,
  bus_number text,
  company text,
  nic_number text,
  email text,
  contact text,
  permitNum text,
  regnum text,
  capacity int,
  type text,
  assigned_route_id text,
  approvedFare float,
  is_active boolean default true,
  verificationStatus text default 'Pending', -- Pending, Verified, Rejected
  documents text[]
);

-- You can add other tables here following the same pattern

-- Drivers Table
create table if not exists drivers (
  driver_id uuid default uuid_generate_v4() primary key,
  first_name text,
  last_name text,
  email text,
  phone text,
  license_number text,
  license_expiry date,
  experience_years int,
  is_active boolean default true,
  assigned_bus_id text,
  hire_date date,
  verification text default 'Pending',
  verification_note text
);

-- Conductors Table
create table if not exists conductors (
  conductor_id uuid default uuid_generate_v4() primary key,
  full_name text,
  nic text,
  phone_number text,
  experience_years int,
  license_number text,
  license_expiry_date date,
  verification_status text default 'Pending',
  verification_note text,
  status text default 'Active',
  assigned_bus_id text,
  hired_date date
);

-- Bus Staff Table (Drivers and Conductors)
create table if not exists bus_staff (
    id uuid default uuid_generate_v4() primary key,
    staff_type text not null, -- 'Driver' or 'Conductor'
    emergency_contact_name text, -- Used as Name
    emergency_contact text, -- Used as Contact Num
    license_number text,
    license_expiry_date date,
    experience_years int,
    verification_status text default 'Pending',
    verification_notes text,
    created_at timestamp with time zone default timezone('utc'::text, now())
);

-- Bus Staff Employment Table
create table if not exists bus_staff_employment (
    id uuid default uuid_generate_v4() primary key,
    staff_id uuid references bus_staff(id),
    employment_status text default 'Active',
    hire_date date,
    created_at timestamp with time zone default timezone('utc'::text, now())
);

-- Lounge Owners Table
create table if not exists lounge_owners (
    id uuid default uuid_generate_v4() primary key,
    manager_full_name text,
    verification_status text default 'Pending',
    verification_notes text
);

-- Lounge Marketplace Categories Table
create table if not exists lounge_marketplace_categories (
    id uuid default uuid_generate_v4() primary key,
    name text
);

-- Update Lounges Table to include new columns
alter table lounges add column if not exists owner_id uuid references lounge_owners(id);
alter table lounges add column if not exists marketplace_category_id uuid references lounge_marketplace_categories(id);
alter table lounges add column if not exists is_operational boolean default true;

-- Update Lounge Owners Table to include contact info
alter table lounge_owners add column if not exists email text;
alter table lounge_owners add column if not exists contact_number text;
alter table lounge_owners add column if not exists nic text;

-- Add verification_documents column to bus_owners if it doesn't exist
alter table bus_owners add column if not exists verification_documents text;

