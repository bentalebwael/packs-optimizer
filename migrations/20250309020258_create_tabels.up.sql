-- Create pack_configurations table
CREATE TABLE IF NOT EXISTS pack_configurations (
    id SERIAL PRIMARY KEY,
    pack_sizes BIGINT ARRAY NOT NULL,
    signature TEXT NOT NULL,
    active BOOLEAN DEFAULT false
);

-- Create unique index on signature
CREATE UNIQUE INDEX IF NOT EXISTS idx_pack_configurations_signature ON pack_configurations(signature);

-- Create order_calculations table
CREATE TABLE IF NOT EXISTS order_calculations (
    id SERIAL PRIMARY KEY,
    order_quantity INTEGER NOT NULL,
    result JSON NOT NULL,
    total_items INTEGER,
    total_packs INTEGER,
    configuration_id INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (configuration_id) REFERENCES pack_configurations(id)
);

-- Create index on configuration_id for better query performance
CREATE INDEX IF NOT EXISTS idx_order_calculations_configuration_id ON order_calculations(configuration_id);