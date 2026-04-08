-- 已执行过旧版 001（尚无 preferred_name）的数据库：补列
-- 全新库若 001 已含 preferred_name，本语句为 no-op
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_name VARCHAR(100);
