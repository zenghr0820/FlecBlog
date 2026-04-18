-- 添加作者卡片背景图片配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES 
('blog.author_card_bg', '', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
