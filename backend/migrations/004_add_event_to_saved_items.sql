-- Add 'event' type to saved_items table
-- This allows users to save events in addition to projects, opportunities, and users

-- Update the CHECK constraint to include 'event'
ALTER TABLE saved_items DROP CONSTRAINT IF EXISTS saved_items_item_type_check;
ALTER TABLE saved_items ADD CONSTRAINT saved_items_item_type_check 
    CHECK (item_type IN ('project', 'opportunity', 'user', 'event'));

