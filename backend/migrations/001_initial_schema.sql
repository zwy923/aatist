-- Initial database schema for Aalto Talent Network
-- This migration creates all tables and indexes in one go

-- Create users table with all fields
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    nickname VARCHAR(100) DEFAULT NULL,
    avatar_url TEXT DEFAULT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'company', 'admin')),
    student_id VARCHAR(255),
    school VARCHAR(255),
    faculty VARCHAR(255),
    major VARCHAR(255) DEFAULT NULL,
    -- Skills (JSONB with Skill + Level structure)
    skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    bio TEXT DEFAULT NULL,
    -- Availability fields
    weekly_hours INTEGER,
    emotional_status VARCHAR(50) DEFAULT NULL,
    weekly_availability JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- Profile visibility
    profile_visibility VARCHAR(20) NOT NULL DEFAULT 'public' 
        CHECK (profile_visibility IN ('public', 'aalto_only', 'private')),
    -- Email verification (token stored in Redis, only status in DB)
    is_verified_email BOOLEAN NOT NULL DEFAULT FALSE,
    -- OAuth (provider + subject for unique identity)
    oauth_provider VARCHAR(50),
    oauth_subject VARCHAR(255),
    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for users table
-- Email: case-insensitive unique index (email should be stored in lowercase)
-- This ensures "AAA@example.com" and "aaa@example.com" are treated as the same
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_ci ON users (LOWER(email));

-- Role for role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
-- Profile visibility for visibility queries
CREATE INDEX IF NOT EXISTS idx_users_profile_visibility ON users(profile_visibility);
-- School, faculty, major for public profile searches
CREATE INDEX IF NOT EXISTS idx_users_school ON users(school) WHERE school IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_faculty ON users(faculty) WHERE faculty IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_major ON users(major) WHERE major IS NOT NULL;
-- OAuth provider and subject for OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_oauth_provider_subject ON users(oauth_provider, oauth_subject) 
    WHERE oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL;
-- Unique constraint for OAuth identity (partial unique index)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_unique ON users(oauth_provider, oauth_subject) 
    WHERE oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL;
-- Time-based queries
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
-- GIN indexes for JSONB fields (for efficient JSON queries and searches)
CREATE INDEX IF NOT EXISTS idx_users_skills_gin ON users USING GIN (skills);
CREATE INDEX IF NOT EXISTS idx_users_weekly_availability_gin ON users USING GIN (weekly_availability);

-- Create function to update updated_at timestamp
-- Drop function if exists to avoid type conflicts, then recreate
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
CREATE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at for users
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create projects table (independent from users.projects JSONB)
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    year INTEGER CHECK (year IS NULL OR (year >= 1900 AND year <= EXTRACT(YEAR FROM NOW()) + 1)),
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    cover_image_url TEXT,
    project_link TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for projects
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_year ON projects(year) WHERE year IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);

-- Create trigger for projects updated_at
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create saved_items table
-- Note: item_id is polymorphic (can reference projects, opportunities, or users)
-- Foreign key constraints are not used here due to polymorphic design
-- Consider adding triggers to validate item existence if needed
CREATE TABLE IF NOT EXISTS saved_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('project', 'opportunity', 'user')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_id, item_type)
);

-- Create indexes for saved_items
CREATE INDEX IF NOT EXISTS idx_saved_items_user_id ON saved_items(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_items_item ON saved_items(item_id, item_type);
CREATE INDEX IF NOT EXISTS idx_saved_items_created_at ON saved_items(created_at);

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    data JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

-- Create trigger for notifications updated_at
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create files table for file-service
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('avatar', 'project_cover', 'post_image', 'resume', 'ai_output', 'other')),
    object_key VARCHAR(500) NOT NULL,
    url TEXT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    metadata TEXT DEFAULT NULL, -- JSON string for additional metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for files table
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(type);
CREATE INDEX IF NOT EXISTS idx_files_user_id_type ON files(user_id, type);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);

-- Create trigger for files updated_at
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

