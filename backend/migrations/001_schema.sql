-- Aalto Talent Network - 完整数据库 schema + 元数据种子
-- 合并自原 001-010 迁移；种子数据原 002_seed.sql 已并入本文件
-- 原 002_portfolio_wizard.sql（projects：short_caption、media_urls、related_services、co_creators、is_published、is_public）已并入下方 CREATE TABLE projects，不再单独迁移
-- 原 002_user_services_price_type_multi.sql（price_type VARCHAR(128)，多值如 hourly,project,negotiable）已体现在下方 user_services 定义
-- 原 003_users_preferred_name.sql（users.preferred_name）已并入下方 users 表

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 通用 updated_at 触发器
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
CREATE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ========== 用户 ==========
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    name VARCHAR(100) NOT NULL,
    preferred_name VARCHAR(100),
    avatar_url TEXT,
    banner_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'alumni', 'org_person', 'org_team')),
    bio TEXT,
    profile_visibility VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (profile_visibility IN ('public', 'aalto_only', 'private')),
    is_verified_email BOOLEAN NOT NULL DEFAULT FALSE,
    role_verified BOOLEAN NOT NULL DEFAULT FALSE,
    oauth_provider VARCHAR(50),
    oauth_subject VARCHAR(255),
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- 学生/校友
    student_id VARCHAR(255),
    school VARCHAR(255),
    faculty VARCHAR(255),
    major VARCHAR(255),
    skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    courses JSONB DEFAULT '[]'::jsonb,
    portfolio_visibility VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (portfolio_visibility IN ('public', 'aalto_only', 'private')),
    -- 组织
    organization_name VARCHAR(255),
    organization_bio TEXT,
    contact_title VARCHAR(100),
    is_affiliated_with_school BOOLEAN NOT NULL DEFAULT FALSE,
    org_size INTEGER,
    -- 档案增强
    website TEXT,
    linkedin TEXT,
    behance TEXT,
    languages TEXT,
    professional_interests TEXT,
    guided_profile_questions JSONB DEFAULT '{}'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_ci ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_profile_visibility ON users(profile_visibility);
CREATE INDEX IF NOT EXISTS idx_users_portfolio_visibility ON users(portfolio_visibility);
CREATE INDEX IF NOT EXISTS idx_users_school ON users(school) WHERE school IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_faculty ON users(faculty) WHERE faculty IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_major ON users(major) WHERE major IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_organization_name ON users(organization_name) WHERE organization_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_affiliated_with_school ON users(is_affiliated_with_school);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_unique ON users(oauth_provider, oauth_subject)
    WHERE oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_skills_gin ON users USING GIN (skills);

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 项目 ==========
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    short_caption VARCHAR(120),
    description TEXT,
    year INTEGER CHECK (year IS NULL OR (year >= 1900 AND year <= EXTRACT(YEAR FROM NOW()) + 1)),
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    media_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    related_services JSONB NOT NULL DEFAULT '[]'::jsonb,
    co_creators JSONB NOT NULL DEFAULT '[]'::jsonb,
    cover_image_url TEXT,
    project_link TEXT,
    client_name VARCHAR(255),
    service_category VARCHAR(100),
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_year ON projects(year) WHERE year IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 收藏 ==========
CREATE TABLE IF NOT EXISTS saved_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('project', 'opportunity', 'user', 'event')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_id, item_type)
);

CREATE INDEX IF NOT EXISTS idx_saved_items_user_id ON saved_items(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_items_item ON saved_items(item_id, item_type);
CREATE INDEX IF NOT EXISTS idx_saved_items_created_at ON saved_items(created_at);

-- ========== 通知 ==========
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

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

-- ========== 文件 ==========
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('avatar', 'project_cover', 'profile_banner', 'post_image', 'resume', 'ai_output', 'other')),
    object_key VARCHAR(500) NOT NULL,
    url TEXT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(type);
CREATE INDEX IF NOT EXISTS idx_files_user_id_type ON files(user_id, type);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);

DROP TRIGGER IF EXISTS update_files_updated_at ON files;
CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 机会 ==========
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

CREATE INDEX IF NOT EXISTS idx_opp_category ON opportunities(category);
CREATE INDEX IF NOT EXISTS idx_opp_location ON opportunities(location);
CREATE INDEX IF NOT EXISTS idx_opp_start_date ON opportunities(start_date);
CREATE INDEX IF NOT EXISTS idx_opp_published_at ON opportunities(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_opp_urgent ON opportunities(urgent);
CREATE INDEX IF NOT EXISTS idx_opp_status ON opportunities(status);
CREATE INDEX IF NOT EXISTS idx_opp_created_by ON opportunities(created_by);
CREATE INDEX IF NOT EXISTS idx_opp_budget_value ON opportunities(budget_value) WHERE budget_value IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_opp_languages ON opportunities USING GIN (languages);
CREATE INDEX IF NOT EXISTS idx_opp_tags ON opportunities USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_opp_status_published_at ON opportunities(status, published_at DESC) WHERE status = 'active';

DROP TRIGGER IF EXISTS update_opportunities_updated_at ON opportunities;
CREATE TRIGGER update_opportunities_updated_at BEFORE UPDATE ON opportunities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 机会申请 ==========
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

CREATE INDEX IF NOT EXISTS idx_opp_applications_user_id ON opportunity_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_opp_applications_opportunity_id ON opportunity_applications(opportunity_id);
CREATE INDEX IF NOT EXISTS idx_opp_applications_status ON opportunity_applications(status);
CREATE INDEX IF NOT EXISTS idx_opp_applications_created_at ON opportunity_applications(created_at);

DROP TRIGGER IF EXISTS update_opportunity_applications_updated_at ON opportunity_applications;
CREATE TRIGGER update_opportunity_applications_updated_at BEFORE UPDATE ON opportunity_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 活动 ==========
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

CREATE INDEX IF NOT EXISTS idx_events_search ON events USING GIN (to_tsvector('english', title || ' ' || organizer));
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time);
CREATE INDEX IF NOT EXISTS idx_events_published_at ON events(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_types ON events USING GIN (type_tags);
CREATE INDEX IF NOT EXISTS idx_events_schools ON events USING GIN (school_tags);
CREATE INDEX IF NOT EXISTS idx_events_languages ON events USING GIN (languages);
CREATE INDEX IF NOT EXISTS idx_events_location ON events(location);
CREATE INDEX IF NOT EXISTS idx_events_free ON events(is_free);
CREATE INDEX IF NOT EXISTS idx_events_external ON events(is_external);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_created_by ON events(created_by);
CREATE INDEX IF NOT EXISTS idx_events_status_start_time ON events(status, start_time) WHERE status = 'active';

DROP TRIGGER IF EXISTS update_events_updated_at ON events;
CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS event_interests (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_event_interests_user_id ON event_interests(user_id);
CREATE INDEX IF NOT EXISTS idx_event_interests_event_id ON event_interests(event_id);

CREATE TABLE IF NOT EXISTS event_going (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_event_going_user_id ON event_going(user_id);
CREATE INDEX IF NOT EXISTS idx_event_going_event_id ON event_going(event_id);

CREATE TABLE IF NOT EXISTS event_comments (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_comments_event_id ON event_comments(event_id);
CREATE INDEX IF NOT EXISTS idx_event_comments_user_id ON event_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_event_comments_created_at ON event_comments(created_at);

CREATE OR REPLACE FUNCTION get_event_interested_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
    SELECT COUNT(*)::BIGINT FROM event_interests WHERE event_id = event_id_param;
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_event_going_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
    SELECT COUNT(*)::BIGINT FROM event_going WHERE event_id = event_id_param;
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_event_comment_count(event_id_param BIGINT)
RETURNS BIGINT AS $$
    SELECT COUNT(*)::BIGINT FROM event_comments WHERE event_id = event_id_param;
$$ LANGUAGE sql STABLE;

-- ========== 元数据 ==========
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
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, type)
);

-- ========== 聊天 ==========
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL,
    from_user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_created
    ON chat_messages (conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_from_user
    ON chat_messages (from_user_id, created_at DESC);

-- 会话已读状态：用于计算未读消息数
CREATE TABLE IF NOT EXISTS chat_conversation_read (
    user_id BIGINT NOT NULL,
    conversation_id VARCHAR(64) NOT NULL,
    last_read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, conversation_id)
);
CREATE INDEX IF NOT EXISTS idx_chat_conversation_read_user
    ON chat_conversation_read (user_id);

-- ========== 用户服务 ==========
CREATE TABLE IF NOT EXISTS user_services (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL,
    title VARCHAR(200),
    short_description VARCHAR(500),
    description TEXT,
    experience_summary TEXT NOT NULL,
    price_type VARCHAR(128),
    price_min INTEGER,
    price_max INTEGER,
    media_urls JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_services_user_id ON user_services(user_id);

DROP TRIGGER IF EXISTS update_user_services_updated_at ON user_services;
CREATE TRIGGER update_user_services_updated_at BEFORE UPDATE ON user_services
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========== 种子数据（skills / tags）==========
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
