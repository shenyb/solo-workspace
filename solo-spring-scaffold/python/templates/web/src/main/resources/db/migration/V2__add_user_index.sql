-- ─── V2__add_user_index.sql ───
-- 为 app_user 表添加业务索引
-- 按需修改，第二个迁移脚本作为示例供用户参考

-- 用户名唯一索引（典型业务场景）
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_user_username ON app_user (username);

-- 邮箱索引（常用于登录/检索）
CREATE INDEX IF NOT EXISTS idx_app_user_email ON app_user (email);

-- 状态 + 创建时间复合索引（列表查询优化）
CREATE INDEX IF NOT EXISTS idx_app_user_status_created ON app_user (status, created_at);
