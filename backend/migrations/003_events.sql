-- Events module migration
-- Creates events table, event_interests, event_going, and event_comments tables

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    organizer TEXT NOT NULL,
    type_tags TEXT[] NOT NULL DEFAULT '{}',
    school_tags TEXT[] NOT NULL DEFAULT '{}',
    is_external BOOLEAN NOT NULL DEFAULT FALSE,
    is_free BOOLEAN NOT NULL DEFAULT FALSE,
    location TEXT NOT NULL,
    languages TEXT[] NOT NULL DEFAULT '{}',
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    max_participants INT,
    description TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'canceled')),
    cover_image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

-- Create indexes for events table
-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_events_search ON events USING GIN (to_tsvector('english', title || ' ' || organizer));
-- Time-based indexes
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time);
CREATE INDEX IF NOT EXISTS idx_events_published_at ON events(published_at DESC);
-- Array field indexes (GIN for array operations)
CREATE INDEX IF NOT EXISTS idx_events_types ON events USING GIN (type_tags);
CREATE INDEX IF NOT EXISTS idx_events_schools ON events USING GIN (school_tags);
CREATE INDEX IF NOT EXISTS idx_events_languages ON events USING GIN (languages);
-- Filter indexes
CREATE INDEX IF NOT EXISTS idx_events_location ON events(location);
CREATE INDEX IF NOT EXISTS idx_events_free ON events(is_free);
CREATE INDEX IF NOT EXISTS idx_events_external ON events(is_external);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_created_by ON events(created_by);
-- Composite index for common queries (active events sorted by start_time)
CREATE INDEX IF NOT EXISTS idx_events_status_start_time ON events(status, start_time) WHERE status = 'active';

-- Create trigger for events updated_at
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create event_interests table (users marking Interested)
CREATE TABLE IF NOT EXISTS event_interests (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, event_id)
);

-- Create indexes for event_interests
CREATE INDEX IF NOT EXISTS idx_event_interests_user_id ON event_interests(user_id);
CREATE INDEX IF NOT EXISTS idx_event_interests_event_id ON event_interests(event_id);
CREATE INDEX IF NOT EXISTS idx_event_interests_created_at ON event_interests(created_at);

-- Create event_going table (users marking Going)
CREATE TABLE IF NOT EXISTS event_going (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, event_id)
);

-- Create indexes for event_going
CREATE INDEX IF NOT EXISTS idx_event_going_user_id ON event_going(user_id);
CREATE INDEX IF NOT EXISTS idx_event_going_event_id ON event_going(event_id);
CREATE INDEX IF NOT EXISTS idx_event_going_created_at ON event_going(created_at);

-- Create event_comments table
CREATE TABLE IF NOT EXISTS event_comments (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for event_comments
CREATE INDEX IF NOT EXISTS idx_event_comments_event_id ON event_comments(event_id);
CREATE INDEX IF NOT EXISTS idx_event_comments_user_id ON event_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_event_comments_created_at ON event_comments(created_at);

-- Create function to get interested count for an event
CREATE OR REPLACE FUNCTION get_event_interested_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM event_interests WHERE event_id = event_id_param);
END;
$$ LANGUAGE plpgsql STABLE;

-- Create function to get going count for an event
CREATE OR REPLACE FUNCTION get_event_going_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM event_going WHERE event_id = event_id_param);
END;
$$ LANGUAGE plpgsql STABLE;

-- Create function to get comment count for an event
CREATE OR REPLACE FUNCTION get_event_comment_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM event_comments WHERE event_id = event_id_param);
END;
$$ LANGUAGE plpgsql STABLE;

