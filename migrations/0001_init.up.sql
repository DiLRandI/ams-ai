CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE asset_categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO asset_categories (name, description, is_system) VALUES
    ('IT devices', 'Computers, phones, tablets, monitors, and accessories.', true),
    ('Networking equipment', 'Routers, switches, access points, and related network devices.', true),
    ('Appliances', 'Home or office appliances.', true),
    ('Furniture', 'Office and household furniture.', true),
    ('Tools', 'Tools and equipment.', true),
    ('Vehicles', 'Vehicle assets with renewal and operating records.', true),
    ('Other / General', 'General physical assets.', true);

CREATE TABLE assets (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'general' CHECK (type IN ('general', 'vehicle')),
    category_id BIGINT NOT NULL REFERENCES asset_categories(id),
    name TEXT NOT NULL,
    brand TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    serial_number TEXT NOT NULL DEFAULT '',
    purchase_date DATE,
    purchase_price NUMERIC(12,2),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'in_repair', 'stored', 'retired', 'disposed')),
    condition TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    assigned_to TEXT NOT NULL DEFAULT '',
    assigned_user_id BIGINT REFERENCES users(id),
    notes TEXT NOT NULL DEFAULT '',
    warranty_start_date DATE,
    warranty_expiry_date DATE,
    warranty_notes TEXT NOT NULL DEFAULT '',
    archived_at TIMESTAMPTZ,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assets_category_id ON assets(category_id);
CREATE INDEX idx_assets_status ON assets(status);
CREATE INDEX idx_assets_type ON assets(type);
CREATE INDEX idx_assets_assigned_user_id ON assets(assigned_user_id);
CREATE INDEX idx_assets_created_by ON assets(created_by);
CREATE INDEX idx_assets_warranty_expiry_date ON assets(warranty_expiry_date);

CREATE TABLE asset_documents (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('bill_invoice', 'warranty', 'insurance', 'license_registration', 'service_receipt', 'manual', 'other')),
    notes TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    uploaded_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_asset_documents_asset_id ON asset_documents(asset_id);
CREATE INDEX idx_asset_documents_type ON asset_documents(type);

CREATE TABLE vehicle_profiles (
    asset_id BIGINT PRIMARY KEY REFERENCES assets(id) ON DELETE CASCADE,
    registration_number TEXT NOT NULL UNIQUE,
    vehicle_type TEXT NOT NULL DEFAULT '',
    chassis_number TEXT NOT NULL DEFAULT '',
    engine_number TEXT NOT NULL DEFAULT '',
    odometer INTEGER,
    assigned_driver TEXT NOT NULL DEFAULT '',
    next_service_date DATE,
    next_service_mileage INTEGER,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vehicle_profiles_registration_number ON vehicle_profiles(registration_number);
CREATE INDEX idx_vehicle_profiles_next_service_date ON vehicle_profiles(next_service_date);

CREATE TABLE vehicle_insurance_records (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT '',
    policy_number TEXT NOT NULL DEFAULT '',
    cost NUMERIC(12,2),
    start_date DATE,
    expiry_date DATE NOT NULL,
    document_id BIGINT REFERENCES asset_documents(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vehicle_insurance_records_asset_id ON vehicle_insurance_records(asset_id);
CREATE INDEX idx_vehicle_insurance_records_expiry_date ON vehicle_insurance_records(expiry_date);

CREATE TABLE vehicle_license_records (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    renewal_type TEXT NOT NULL DEFAULT '',
    reference_number TEXT NOT NULL DEFAULT '',
    cost NUMERIC(12,2),
    issue_date DATE,
    expiry_date DATE NOT NULL,
    document_id BIGINT REFERENCES asset_documents(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vehicle_license_records_asset_id ON vehicle_license_records(asset_id);
CREATE INDEX idx_vehicle_license_records_expiry_date ON vehicle_license_records(expiry_date);

CREATE TABLE vehicle_emission_records (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    inspection_type TEXT NOT NULL DEFAULT '',
    reference_number TEXT NOT NULL DEFAULT '',
    cost NUMERIC(12,2),
    issue_date DATE,
    expiry_date DATE NOT NULL,
    document_id BIGINT REFERENCES asset_documents(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vehicle_emission_records_asset_id ON vehicle_emission_records(asset_id);
CREATE INDEX idx_vehicle_emission_records_expiry_date ON vehicle_emission_records(expiry_date);

CREATE TABLE service_records (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    service_type TEXT NOT NULL DEFAULT 'service' CHECK (service_type IN ('service', 'repair')),
    service_date DATE NOT NULL,
    cost NUMERIC(12,2),
    vendor TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    mileage INTEGER,
    next_service_date DATE,
    next_service_mileage INTEGER,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_service_records_asset_id ON service_records(asset_id);
CREATE INDEX idx_service_records_service_date ON service_records(service_date);
CREATE INDEX idx_service_records_next_service_date ON service_records(next_service_date);

CREATE TABLE fuel_logs (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    fuel_date DATE NOT NULL,
    fuel_type TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(12,3) NOT NULL,
    cost NUMERIC(12,2) NOT NULL,
    odometer INTEGER,
    notes TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_fuel_logs_asset_id ON fuel_logs(asset_id);
CREATE INDEX idx_fuel_logs_fuel_date ON fuel_logs(fuel_date);

CREATE TABLE reminders (
    id BIGSERIAL PRIMARY KEY,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL CHECK (source_type IN ('warranty', 'insurance', 'license', 'emission', 'service')),
    source_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    due_date DATE NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('upcoming', 'due', 'overdue')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, due_date)
);

CREATE INDEX idx_reminders_asset_id ON reminders(asset_id);
CREATE INDEX idx_reminders_due_date ON reminders(due_date);
CREATE INDEX idx_reminders_state ON reminders(state);
