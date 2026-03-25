-- 添加页脚右侧链接配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES ('blog.footer_links', '[{"name":"订阅","url":"/subscribe"},{"name":"源码","url":"https://github.com/talen8/FlecBlog"}]', 'blog', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
