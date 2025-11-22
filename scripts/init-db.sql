-- 创建数据库（如果不存在）
CREATE DATABASE trpg_solo_engine;

-- 连接到数据库
\c trpg_solo_engine;

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 注意：表结构将由GORM自动迁移创建
