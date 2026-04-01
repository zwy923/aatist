-- Allow comma-separated price modes: hourly,project,negotiable (e.g. hourly,negotiable)
ALTER TABLE user_services DROP CONSTRAINT IF EXISTS user_services_price_type_check;
ALTER TABLE user_services ALTER COLUMN price_type TYPE VARCHAR(128);
