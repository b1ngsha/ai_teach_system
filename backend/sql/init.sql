-- 创建数据库
DROP DATABASE IF EXISTS ai_teach_system;
CREATE DATABASE IF NOT EXISTS ai_teach_system DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE ai_teach_system;

-- 题目表
CREATE TABLE IF NOT EXISTS problems (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    leetcode_id INT UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    title_slug VARCHAR(255) NOT NULL,
    difficulty ENUM('Easy', 'Medium', 'Hard') NOT NULL,
    content TEXT NOT NULL,
    sample_testcases TEXT,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    INDEX idx_problems_deleted_at (deleted_at)
);

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    INDEX idx_tags_deleted_at (deleted_at)
);

-- 题目-标签关联表
CREATE TABLE IF NOT EXISTS problem_tags (
    problem_id BIGINT,
    tag_id BIGINT,
    PRIMARY KEY (problem_id, tag_id),
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 知识点表
CREATE TABLE IF NOT EXISTS knowledge_points (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME(3),
    updated_at DATETIME(3),
    deleted_at DATETIME(3),
    INDEX idx_knowledge_points_deleted_at (deleted_at)
);

-- 题目-知识点关联表
CREATE TABLE IF NOT EXISTS problem_knowledge_points (
    problem_id BIGINT,
    knowledge_point_id BIGINT,
    PRIMARY KEY (problem_id, knowledge_point_id),
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
    FOREIGN KEY (knowledge_point_id) REFERENCES knowledge_points(id) ON DELETE CASCADE
); 