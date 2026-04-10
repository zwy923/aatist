-- Optional field of study / concentration (separate from degree programme stored in major).
ALTER TABLE users ADD COLUMN IF NOT EXISTS study_major VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_users_study_major ON users(study_major) WHERE study_major IS NOT NULL;
