-- Metadata tables for skills, courses, and tags
CREATE TABLE IF NOT EXISTS skills (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    school VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- e.g., 'opportunity_category', 'event_type', 'post_tag'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, type)
);

-- Seed initial data
INSERT INTO skills (name, category) VALUES 
('React', 'Frontend'),
('Vue.js', 'Frontend'),
('Angular', 'Frontend'),
('Go', 'Backend'),
('Python', 'Backend'),
('Node.js', 'Backend'),
('PostgreSQL', 'Database'),
('MongoDB', 'Database'),
('Docker', 'DevOps'),
('Kubernetes', 'DevOps')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tags (name, type) VALUES 
('Full-time', 'opportunity_type'),
('Part-time', 'opportunity_type'),
('Internship', 'opportunity_type'),
('Freelance', 'opportunity_type'),
('Software Development', 'opportunity_category'),
('Design', 'opportunity_category'),
('Marketing', 'opportunity_category'),
('General', 'post_category'),
('Question', 'post_category'),
('Announcement', 'post_category')
ON CONFLICT (name, type) DO NOTHING;
