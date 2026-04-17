-- 导入映射配置表
CREATE TABLE IF NOT EXISTS meta_mappings (
    id SERIAL PRIMARY KEY,
    template_key VARCHAR(50) NOT NULL,           -- 平台类型：vuepress, hexo, hugo, jekyll, custom
    template_name VARCHAR(100) NOT NULL,     -- 平台显示名称
    source_field VARCHAR(100) NOT NULL,      -- 源字段名（如 re, star, abbrlink）
    target_field VARCHAR(100) NOT NULL,      -- 目标字段名（如 slug, is_top, is_essence）
    field_type VARCHAR(20) NOT NULL,         -- 字段类型：string, boolean, date, array
    transform_rule TEXT,                     -- 转换规则（JSON格式）
    is_active BOOLEAN DEFAULT TRUE,          -- 是否启用
    is_system BOOLEAN DEFAULT FALSE,         -- 是否系统预设
    sort_order INT DEFAULT 0,                -- 排序
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_key, source_field)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_meta_mappings_template ON meta_mappings(template_key);
CREATE INDEX IF NOT EXISTS idx_meta_mappings_active ON meta_mappings(is_active);