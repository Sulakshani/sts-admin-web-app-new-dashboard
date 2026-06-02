package database

import (
	"database/sql"
	"log"
	"sts-backend/internal/config"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) {
	var err error
	DB, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	DB.SetMaxOpenConns(cfg.MaxConnections)
	DB.SetMaxIdleConns(cfg.MaxIdleConnections)
	DB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the database")

	// Auto-migrate tables
	_, err = DB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS bus_staff (
			id uuid default uuid_generate_v4() primary key,
			staff_type text not null,
			emergency_contact_name text,
			emergency_contact text,
			license_number text,
			license_expiry_date date,
			experience_years int,
			verification_status text default 'Pending',
			verification_notes text,
			created_at timestamp with time zone default timezone('utc'::text, now())
		);

		CREATE TABLE IF NOT EXISTS bus_staff_employment (
			id uuid default uuid_generate_v4() primary key,
			staff_id uuid references bus_staff(id),
			employment_status text default 'Active',
			hire_date date,
			created_at timestamp with time zone default timezone('utc'::text, now())
		);

		CREATE TABLE IF NOT EXISTS lounge_owners (
			id uuid default uuid_generate_v4() primary key,
			manager_full_name text,
			verification_status text default 'Pending',
			verification_notes text
		);

		CREATE TABLE IF NOT EXISTS lounge_marketplace_categories (
			id uuid default uuid_generate_v4() primary key,
			name text
		);

		CREATE TABLE IF NOT EXISTS lounges (
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
			verification text default 'Pending',
			verification_note text
		);

		ALTER TABLE lounges ADD COLUMN IF NOT EXISTS owner_id uuid REFERENCES lounge_owners(id);
		ALTER TABLE lounges ADD COLUMN IF NOT EXISTS marketplace_category_id uuid REFERENCES lounge_marketplace_categories(id);
		ALTER TABLE lounges ADD COLUMN IF NOT EXISTS is_operational boolean DEFAULT true;

		-- Admin Users Table (for authentication system)
		CREATE TABLE IF NOT EXISTS admin_users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('super_admin', 'supervisor', 'admin')),
			app_scope TEXT CHECK (app_scope IN ('bus', 'driver', 'lounges', 'passenger') OR app_scope IS NULL),
			supervisor_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
			function_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL
		);

		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS function_permissions jsonb NOT NULL DEFAULT '[]'::jsonb;
		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL;
		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'admin';
		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS app_scope TEXT;
		ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS supervisor_id UUID REFERENCES admin_users(id) ON DELETE SET NULL;

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'admin_users_role_check'
				  AND conrelid = 'admin_users'::regclass
			) THEN
				ALTER TABLE admin_users
					ADD CONSTRAINT admin_users_role_check
					CHECK (role IN ('super_admin', 'supervisor', 'admin'));
			END IF;

			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'admin_users_app_scope_check'
				  AND conrelid = 'admin_users'::regclass
			) THEN
				ALTER TABLE admin_users
					ADD CONSTRAINT admin_users_app_scope_check
					CHECK (app_scope IN ('bus', 'driver', 'lounges', 'passenger') OR app_scope IS NULL);
			END IF;
		END $$;
		
		ALTER TABLE users ADD COLUMN IF NOT EXISTS function_permissions jsonb NOT NULL DEFAULT '[]'::jsonb;

		-- Complaint escalation compatibility migrations for legacy databases
		ALTER TABLE complaint_escalations ADD COLUMN IF NOT EXISTS source_app TEXT;
		ALTER TABLE complaint_escalations ADD COLUMN IF NOT EXISTS previous_assigned_admin_id UUID REFERENCES admin_users(id);
		ALTER TABLE complaint_escalations ADD COLUMN IF NOT EXISTS assigned_to_admin_id UUID REFERENCES admin_users(id);

		UPDATE complaint_escalations ce
		SET source_app = COALESCE(
			ce.source_app,
			(
				SELECT CASE
					WHEN 'driver' = ANY(u.roles) THEN 'driver'
					WHEN 'lounge_owner' = ANY(u.roles) THEN 'lounges'
					WHEN 'passenger' = ANY(u.roles) THEN 'passenger'
					WHEN 'bus_owner' = ANY(u.roles) OR 'conductor' = ANY(u.roles) THEN 'bus'
					ELSE 'bus'
				END
				FROM report_issues ri
				JOIN users u ON u.id = ri.reported_by_id
				WHERE ri.id = ce.complaint_id
			),
			'bus'
		)
		WHERE ce.source_app IS NULL OR ce.source_app = '';

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'complaint_escalations_source_app_check'
				  AND conrelid = 'complaint_escalations'::regclass
			) THEN
				ALTER TABLE complaint_escalations
					ADD CONSTRAINT complaint_escalations_source_app_check
					CHECK (source_app IN ('bus', 'driver', 'lounges', 'passenger'));
			END IF;
		END $$;

		ALTER TABLE complaint_escalation_history ADD COLUMN IF NOT EXISTS from_admin_id UUID REFERENCES admin_users(id);
		ALTER TABLE complaint_escalation_history ADD COLUMN IF NOT EXISTS to_admin_id UUID REFERENCES admin_users(id);

		ALTER TABLE lounge_owners ADD COLUMN IF NOT EXISTS email text;
		ALTER TABLE lounge_owners ADD COLUMN IF NOT EXISTS contact_number text;
		ALTER TABLE lounge_owners ADD COLUMN IF NOT EXISTS nic text;

		-- Add created_at to buses table if it doesn't exist
		ALTER TABLE buses ADD COLUMN IF NOT EXISTS created_at timestamp with time zone default timezone('utc'::text, now());

		-- Make bus_owners.user_id nullable (can be assigned later)
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'bus_owners' AND column_name = 'user_id') THEN
				ALTER TABLE bus_owners ALTER COLUMN user_id DROP NOT NULL;
			END IF;
		END $$;

		-- Update bus_staff_employment check constraint to include 'inactive'
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.table_constraints 
					   WHERE table_name = 'bus_staff_employment' AND constraint_name = 'bus_staff_employment_employment_status_check') THEN
				ALTER TABLE bus_staff_employment DROP CONSTRAINT bus_staff_employment_employment_status_check;
			END IF;
			
			ALTER TABLE bus_staff_employment ADD CONSTRAINT bus_staff_employment_employment_status_check 
				CHECK (employment_status IN ('pending', 'active', 'inactive', 'terminated', 'resigned', 'suspended'));
		END $$;

		-- Lounge Booking Tables
		CREATE TABLE IF NOT EXISTS bookings (
			bus_booking_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
			passenger_name TEXT NOT NULL,
			passenger_phone TEXT,
			booking_reference TEXT UNIQUE,
			scheduled_arrival TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
		);

		CREATE TABLE IF NOT EXISTS lounge_bookings (
			lounge_booking_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
			bus_booking_id UUID REFERENCES bookings(bus_booking_id),
			lounge_name TEXT NOT NULL,
			scheduled_arrival TIMESTAMP WITH TIME ZONE,
			pricing_type TEXT,
			number_of_guests INT DEFAULT 1,
			selected_amenities JSONB DEFAULT '[]'::jsonb,
			booking_type TEXT,
			total_amount DECIMAL(10, 2) DEFAULT 0,
			payment_status TEXT DEFAULT 'Pending',
			status TEXT DEFAULT 'Pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW()),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
		);

		-- Add scheduled_arrival column if missing
		ALTER TABLE lounge_bookings ADD COLUMN IF NOT EXISTS scheduled_arrival TIMESTAMP WITH TIME ZONE;

		-- Add missing columns for lounge bookings
		ALTER TABLE lounge_bookings ADD COLUMN IF NOT EXISTS primary_guest_name TEXT;
		ALTER TABLE lounge_bookings ADD COLUMN IF NOT EXISTS primary_guest_phone TEXT;
		ALTER TABLE lounge_bookings ADD COLUMN IF NOT EXISTS booking_reference TEXT;
		ALTER TABLE lounge_bookings ADD COLUMN IF NOT EXISTS master_booking_id UUID;

		-- Convert enum columns to TEXT if they exist
		DO $$ 
		BEGIN
			-- Convert pricing_type from enum to TEXT
			IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lounge_booking_duration_type') THEN
				ALTER TABLE lounge_bookings ALTER COLUMN pricing_type TYPE TEXT USING pricing_type::TEXT;
				DROP TYPE lounge_booking_duration_type CASCADE;
			END IF;
			
			-- Convert booking_type from enum to TEXT
			IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lounge_booking_type') THEN
				ALTER TABLE lounge_bookings ALTER COLUMN booking_type TYPE TEXT USING booking_type::TEXT;
				DROP TYPE lounge_booking_type CASCADE;
			END IF;
		END $$;

		-- Make lounge_id nullable or drop it if it exists
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_bookings' AND column_name = 'lounge_id') THEN
				ALTER TABLE lounge_bookings ALTER COLUMN lounge_id DROP NOT NULL;
			END IF;
			
			-- Make booking_reference nullable
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_bookings' AND column_name = 'booking_reference') THEN
				ALTER TABLE lounge_bookings ALTER COLUMN booking_reference DROP NOT NULL;
			END IF;
			
			-- Make user_id nullable
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_bookings' AND column_name = 'user_id') THEN
				ALTER TABLE lounge_bookings ALTER COLUMN user_id DROP NOT NULL;
			END IF;
			
			-- Drop the bus_link_check constraint if it exists
			IF EXISTS (SELECT 1 FROM information_schema.table_constraints 
					   WHERE table_name = 'lounge_bookings' AND constraint_name = 'lounge_bookings_bus_link_check') THEN
				ALTER TABLE lounge_bookings DROP CONSTRAINT lounge_bookings_bus_link_check;
			END IF;
			
			-- Drop the lounge_type_check constraint if it exists
			IF EXISTS (SELECT 1 FROM information_schema.table_constraints 
					   WHERE table_name = 'lounge_bookings' AND constraint_name = 'lounge_bookings_lounge_type_check') THEN
				ALTER TABLE lounge_bookings DROP CONSTRAINT lounge_bookings_lounge_type_check;
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS lounge_booking_pre_orders (
			pre_order_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
			lounge_booking_id UUID REFERENCES lounge_bookings(lounge_booking_id) ON DELETE CASCADE UNIQUE,
			product_name TEXT,
			quantity INT DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::text, NOW())
		);

		-- Make product_id nullable if it exists (for marketplace items without product reference)
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_booking_pre_orders' AND column_name = 'product_id') THEN
				ALTER TABLE lounge_booking_pre_orders ALTER COLUMN product_id DROP NOT NULL;
			END IF;
			
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_booking_pre_orders' AND column_name = 'product_type') THEN
				ALTER TABLE lounge_booking_pre_orders ALTER COLUMN product_type DROP NOT NULL;
			END IF;
			
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_booking_pre_orders' AND column_name = 'unit_price') THEN
				ALTER TABLE lounge_booking_pre_orders ALTER COLUMN unit_price DROP NOT NULL;
			END IF;
			
			IF EXISTS (SELECT 1 FROM information_schema.columns 
					   WHERE table_name = 'lounge_booking_pre_orders' AND column_name = 'total_price') THEN
				ALTER TABLE lounge_booking_pre_orders ALTER COLUMN total_price DROP NOT NULL;
			END IF;
		END $$;

		CREATE INDEX IF NOT EXISTS idx_lounge_bookings_bus_booking ON lounge_bookings(bus_booking_id);
		CREATE INDEX IF NOT EXISTS idx_lounge_bookings_payment_status ON lounge_bookings(payment_status);
		CREATE INDEX IF NOT EXISTS idx_lounge_bookings_status ON lounge_bookings(status);
		CREATE INDEX IF NOT EXISTS idx_lounge_bookings_created_at ON lounge_bookings(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_lounge_booking_pre_orders_booking ON lounge_booking_pre_orders(lounge_booking_id);
		
		-- Admin Users Indexes
		CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);
		CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
		CREATE INDEX IF NOT EXISTS idx_admin_users_app_scope ON admin_users(app_scope);
		CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);
		CREATE INDEX IF NOT EXISTS idx_admin_users_supervisor_id ON admin_users(supervisor_id);
		CREATE INDEX IF NOT EXISTS idx_admin_users_created_at ON admin_users(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_complaint_escalations_source_app ON complaint_escalations(source_app);
	`)
	if err != nil {
		log.Printf("Error creating/updating tables: %v", err)
	} else {
		log.Println("Database schema updated successfully")
	}
}
