-- Initial database schema for Aalto Talent Network
-- This migration creates all tables and indexes in one go

-- Create users table with all fields
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) DEFAULT NULL, -- OAuth users may not have password
    name VARCHAR(100) NOT NULL,
    avatar_url TEXT DEFAULT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'alumni', 'org_person', 'org_team')),
    bio TEXT DEFAULT NULL,
    -- Profile visibility (all roles)
    profile_visibility VARCHAR(20) NOT NULL DEFAULT 'public' 
        CHECK (profile_visibility IN ('public', 'aalto_only', 'private')),
    -- Email verification
    is_verified_email BOOLEAN NOT NULL DEFAULT FALSE,
    -- Role verification: automatically true for school email domains (e.g., @aalto.fi)
    -- This indicates the user has a verified school email, but still needs email verification
    role_verified BOOLEAN NOT NULL DEFAULT FALSE,
    -- OAuth (provider + subject for unique identity)
    oauth_provider VARCHAR(50),
    oauth_subject VARCHAR(255),
    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- Student/Alumni specific fields
    student_id VARCHAR(255),
    school VARCHAR(255),
    faculty VARCHAR(255),
    major VARCHAR(255),
    skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    weekly_hours INTEGER,
    weekly_availability JSONB NOT NULL DEFAULT '[]'::jsonb,
    portfolio_visibility VARCHAR(20) NOT NULL DEFAULT 'public'
        CHECK (portfolio_visibility IN ('public', 'aalto_only', 'private')),
    -- Organization specific fields (org_person, org_team)
    organization_name VARCHAR(255),
    organization_bio TEXT,
    contact_title VARCHAR(100),
    is_affiliated_with_school BOOLEAN NOT NULL DEFAULT FALSE,
    org_size INTEGER
);

-- Create indexes for users table
-- Email: case-insensitive unique index (email should be stored in lowercase)
-- This ensures "AAA@example.com" and "aaa@example.com" are treated as the same
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_ci ON users (LOWER(email));

-- Role for role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
-- Profile visibility for visibility queries
CREATE INDEX IF NOT EXISTS idx_users_profile_visibility ON users(profile_visibility);
-- Portfolio visibility for portfolio queries
CREATE INDEX IF NOT EXISTS idx_users_portfolio_visibility ON users(portfolio_visibility);
-- School, faculty, major for public profile searches
CREATE INDEX IF NOT EXISTS idx_users_school ON users(school) WHERE school IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_faculty ON users(faculty) WHERE faculty IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_major ON users(major) WHERE major IS NOT NULL;
-- Organization fields for org searches
CREATE INDEX IF NOT EXISTS idx_users_organization_name ON users(organization_name) WHERE organization_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_affiliated_with_school ON users(is_affiliated_with_school);
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
-- Note: item_id is polymorphic (can reference projects, opportunities, users, or events)
-- Foreign key constraints are not used here due to polymorphic design
-- Consider adding triggers to validate item existence if needed
CREATE TABLE IF NOT EXISTS saved_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('project', 'opportunity', 'user', 'event')),
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

-- Enable pg_trgm extension for fuzzy search (idempotent)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Community: discussion posts table
CREATE TABLE IF NOT EXISTS discussion_posts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    like_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    tsv tsvector NOT NULL DEFAULT to_tsvector('english', ''),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Triggers for discussion_posts
DROP TRIGGER IF EXISTS update_discussion_posts_updated_at ON discussion_posts;
CREATE TRIGGER update_discussion_posts_updated_at BEFORE UPDATE ON discussion_posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP FUNCTION IF EXISTS discussion_posts_tsv_trigger() CASCADE;
CREATE FUNCTION discussion_posts_tsv_trigger()
RETURNS TRIGGER AS $$
BEGIN
    NEW.tsv := to_tsvector('english', coalesce(NEW.content, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS discussion_posts_tsv_update ON discussion_posts;
CREATE TRIGGER discussion_posts_tsv_update BEFORE INSERT OR UPDATE ON discussion_posts
    FOR EACH ROW EXECUTE FUNCTION discussion_posts_tsv_trigger();

-- Indexes for discussion_posts
CREATE INDEX IF NOT EXISTS idx_discussion_posts_user_id ON discussion_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_discussion_posts_category_created_at ON discussion_posts(category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_discussion_posts_created_at ON discussion_posts(created_at);
CREATE INDEX IF NOT EXISTS idx_discussion_posts_title_trgm ON discussion_posts USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_discussion_posts_tsv ON discussion_posts USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_discussion_posts_tags_gin ON discussion_posts USING GIN (tags);

-- Post comments table
CREATE TABLE IF NOT EXISTS post_comments (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES discussion_posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id BIGINT NULL REFERENCES post_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS update_post_comments_updated_at ON post_comments;
CREATE TRIGGER update_post_comments_updated_at BEFORE UPDATE ON post_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX IF NOT EXISTS idx_post_comments_post_id_created_at ON post_comments(post_id, created_at);
CREATE INDEX IF NOT EXISTS idx_post_comments_user_id ON post_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_post_comments_parent_id ON post_comments(parent_id);

-- Post likes table
CREATE TABLE IF NOT EXISTS post_likes (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES discussion_posts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_likes_user_id ON post_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes(post_id);

