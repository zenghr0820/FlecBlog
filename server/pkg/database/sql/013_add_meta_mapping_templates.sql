-- 映射模版表（一个模版包含多条 meta_mappings）
CREATE TABLE IF NOT EXISTS meta_mapping_templates (
    id SERIAL PRIMARY KEY,
    template_key VARCHAR(50) NOT NULL UNIQUE,      -- 模版 key：导入时选择的标识
    template_name VARCHAR(100) NOT NULL,       -- 模版显示名称
    description TEXT,                          -- 描述（可选）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

