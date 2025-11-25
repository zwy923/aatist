-- Opportunities module migration
-- Creates opportunities table, favorites table, and applications table

-- Create opportunities table
CREATE TABLE IF NOT EXISTS opportunities (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    organization TEXT NOT NULL,
    category TEXT NOT NULL,
    budget_type TEXT NOT NULL CHECK (budget_type IN ('hourly', 'fixed')),
    budget_value NUMERIC,
    location TEXT NOT NULL,
    duration_months INT,
    languages TEXT[] DEFAULT '{}',
    start_date DATE,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    urgent BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    tags TEXT[] DEFAULT '{}',
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for opportunities table
-- Category index for filtering
CREATE INDEX IF NOT EXISTS idx_opp_category ON opportunities(category);
-- Location index for filtering
CREATE INDEX IF NOT EXISTS idx_opp_location ON opportunities(location);
-- Start date index for date range queries
CREATE INDEX IF NOT EXISTS idx_opp_start_date ON opportunities(start_date);
-- Published at index for sorting
CREATE INDEX IF NOT EXISTS idx_opp_published_at ON opportunities(published_at DESC);
-- Urgent index for filtering
CREATE INDEX IF NOT EXISTS idx_opp_urgent ON opportunities(urgent);
-- Status index for filtering active/closed
CREATE INDEX IF NOT EXISTS idx_opp_status ON opportunities(status);
-- Created by index for user's opportunities
CREATE INDEX IF NOT EXISTS idx_opp_created_by ON opportunities(created_by);
-- Budget value index for budget range queries
CREATE INDEX IF NOT EXISTS idx_opp_budget_value ON opportunities(budget_value) WHERE budget_value IS NOT NULL;
-- GIN indexes for array fields (languages, tags)
CREATE INDEX IF NOT EXISTS idx_opp_languages ON opportunities USING GIN (languages);
CREATE INDEX IF NOT EXISTS idx_opp_tags ON opportunities USING GIN (tags);
-- Composite index for common filter combinations
CREATE INDEX IF NOT EXISTS idx_opp_status_published_at ON opportunities(status, published_at DESC) WHERE status = 'active';

-- Create trigger for opportunities updated_at
DROP TRIGGER IF EXISTS update_opportunities_updated_at ON opportunities;
CREATE TRIGGER update_opportunities_updated_at BEFORE UPDATE ON opportunities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Note: Favorites/saved items are handled by user-service's saved_items table
-- which supports saving opportunities, users, projects, and future events

-- Create opportunity_applications table (for apply functionality)
CREATE TABLE IF NOT EXISTS opportunity_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id BIGINT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    message TEXT,
    cv_url TEXT,
    portfolio_url TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'accepted', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, opportunity_id)
);

-- Create indexes for opportunity_applications
CREATE INDEX IF NOT EXISTS idx_opp_applications_user_id ON opportunity_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_opp_applications_opportunity_id ON opportunity_applications(opportunity_id);
CREATE INDEX IF NOT EXISTS idx_opp_applications_status ON opportunity_applications(status);
CREATE INDEX IF NOT EXISTS idx_opp_applications_created_at ON opportunity_applications(created_at);

-- Create trigger for opportunity_applications updated_at
DROP TRIGGER IF EXISTS update_opportunity_applications_updated_at ON opportunity_applications;
CREATE TRIGGER update_opportunity_applications_updated_at BEFORE UPDATE ON opportunity_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

