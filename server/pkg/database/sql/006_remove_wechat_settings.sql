-- 移除微信公众号配置（功能已调整为仅本地渲染，不再需要 API 凭证）
DELETE FROM settings WHERE "group" = 'wechat';
