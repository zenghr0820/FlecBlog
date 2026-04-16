-- 添加浅色和深色主题背景图片配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES 
('blog.background_image_light', '', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('blog.background_image_dark', '', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
