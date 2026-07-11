-- Waitlist signups collected from the temporary user-research landing page.
CREATE TABLE IF NOT EXISTS waitlist_entries (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    interest VARCHAR(20) NOT NULL CHECK (interest IN ('talent', 'opportunity')),
    consent VARCHAR(20) CHECK (consent IN ('agreed', 'declined')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_waitlist_entries_email ON waitlist_entries(email);
