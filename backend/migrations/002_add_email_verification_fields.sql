-- Add email verification fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email_verification_token TEXT,
ADD COLUMN IF NOT EXISTS email_verification_expires TIMESTAMP WITH TIME ZONE;

-- Create index on email_verification_token for faster lookups (optional)
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users(email_verification_token) 
WHERE email_verification_token IS NOT NULL;

