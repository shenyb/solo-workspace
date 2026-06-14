-- ─── V1__init.sql ───
-- H2 初始化脚本（MySQL 兼容模式）
-- 如果使用 MySQL，对应表结构不变，迁移脚本通用

CREATE TABLE IF NOT EXISTS `app_user` (
    `id`         BIGINT       AUTO_INCREMENT PRIMARY KEY,
    `username`   VARCHAR(64)  NOT NULL COMMENT '用户名',
    `email`      VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
    `status`     TINYINT      DEFAULT 1    COMMENT '状态: 1=正常 0=禁用',
    `created_at` DATETIME     DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 初始化演示数据
INSERT INTO `app_user` (`username`, `email`, `status`) VALUES
('admin', 'admin@example.com', 1),
('test', 'test@example.com', 1);
