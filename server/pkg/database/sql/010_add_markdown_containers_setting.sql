-- 添加 Markdown 容器配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES (
  'blog.markdown_containers',
  '[
    {"name":"tip","target":"note","params":"info","system":"vuepress"},
    {"name":"warning","target":"note","params":"warning","system":"vuepress"},
    {"name":"danger","target":"note","params":"error","system":"vuepress"}
  ]',
  'blog',
  TRUE,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;

