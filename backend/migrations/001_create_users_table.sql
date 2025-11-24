-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    nickname VARCHAR(100) DEFAULT NULL,
    avatar_url TEXT DEFAULT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'company', 'admin')),
    student_id VARCHAR(64),
    school VARCHAR(255),
    faculty VARCHAR(255),
    major VARCHAR(255) DEFAULT NULL,
    availability VARCHAR(100) DEFAULT NULL,
    projects JSONB NOT NULL DEFAULT '[]'::jsonb,
    skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    bio TEXT DEFAULT NULL,
    is_verified_email BOOLEAN NOT NULL DEFAULT FALSE,
    email_verification_token TEXT,
    email_verification_expires TIMESTAMP WITH TIME ZONE,
    oauth_provider VARCHAR(50),
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index on role for role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_school ON users(school) WHERE school IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token
    ON users(email_verification_token) WHERE email_verification_token IS NOT NULL;

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

