-- Job title / position shown under poster name on opportunity cards
ALTER TABLE opportunities
    ADD COLUMN IF NOT EXISTS position TEXT NOT NULL DEFAULT '';
