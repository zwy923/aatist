-- 元数据种子数据

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
