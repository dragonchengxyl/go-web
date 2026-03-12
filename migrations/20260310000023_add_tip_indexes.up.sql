-- Migration 023: Add tip metadata to orders
-- Tips are stored as orders with metadata->>'type' = 'tip'
-- No schema change needed since orders.metadata is JSONB.
-- This migration adds an index for efficient tip queries.

CREATE INDEX IF NOT EXISTS idx_orders_metadata_type ON orders ((metadata->>'type')) WHERE metadata->>'type' IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_metadata_to_user ON orders ((metadata->>'to_user_id')) WHERE metadata->>'to_user_id' IS NOT NULL;
