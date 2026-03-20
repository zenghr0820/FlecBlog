-- 添加 AI 提示词配置项
INSERT INTO settings (key, value, "group", is_public, created_at, updated_at)
VALUES
('ai.summary_prompt', '你是一位博客作者，请根据文章内容生成中文摘要。要求：1. 以作者视角介绍文章；2. 控制在50到100字之间；3. 只输出摘要正文，不要附加解释；4. 内容简洁准确，覆盖核心信息；5. 不要空泛拔高文章意义。', 'ai', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ai.ai_summary_prompt', '你是一名AI助手，请根据文章内容生成中文总结。要求：1. 以旁观者视角总结并推荐文章；2. 控制在150到200字之间；3. 只输出总结正文，不要附加解释；4. 保持语言自然、信息完整；5. 采用“这篇文章...”的表述方式。', 'ai', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ai.title_prompt', '你是一位资深技术作者，请根据文章内容生成1个中文标题。要求：1. 突出主题亮点和核心价值；2. 控制在15到25字之间；3. 尽量不用标点符号；4. 只返回标题本身，不要解释。', 'ai', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
