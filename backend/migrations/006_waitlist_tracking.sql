-- Ad-attribution tracking for the waitlist research landing page.
-- page_views logs every visit (regardless of signup) so conversion rate per
-- campaign can be computed; the same UTM/referrer/user-agent fields are also
-- attached to each waitlist entry so converted signups carry their source too.

CREATE TABLE IF NOT EXISTS page_views (
    id BIGSERIAL PRIMARY KEY,
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    utm_content VARCHAR(255),
    utm_term VARCHAR(255),
    referrer TEXT,
    user_agent TEXT,
    tab VARCHAR(20) CHECK (tab IN ('talent', 'opportunity')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_page_views_utm_source ON page_views(utm_source);
CREATE INDEX IF NOT EXISTS idx_page_views_utm_campaign ON page_views(utm_campaign);

ALTER TABLE waitlist_entries
    ADD COLUMN IF NOT EXISTS utm_source VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_medium VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_campaign VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_content VARCHAR(255),
    ADD COLUMN IF NOT EXISTS utm_term VARCHAR(255),
    ADD COLUMN IF NOT EXISTS referrer TEXT,
    ADD COLUMN IF NOT EXISTS user_agent TEXT;

CREATE INDEX IF NOT EXISTS idx_waitlist_entries_utm_source ON waitlist_entries(utm_source);
