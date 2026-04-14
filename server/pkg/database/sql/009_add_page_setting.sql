-- 添加页面配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES ('blog.moments_size', '30', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;

INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES ('blog.message_content', '时光流逝，岁月如歌。
愿每句话都传递温暖。
欢迎来到留言天地。
分享你的心声与故事。
生活中的点滴感悟。
对未来的美好期许。
让文字连接彼此心灵。
期待你的真挚留言。', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;

-- 添加首页布局配置
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES ('blog.home_layout', 'waterfall', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
