# AI Teach System

## 项目简介

SZU 毕业设计 —— 面向编程类课程的人工智能辅助学习系统的教师端设计与开发

## 技术栈

- 后端：Gin + Gorm + MySQL


## 功能

- 采用 cron + goroutine 定时异步的方式，自动从 LeetCode 题库抓取题目数据，并同步到数据库中


## 项目结构
- `controllers/`: 控制器
- `models/`: 数据库模型
- `routes/`: 路由
- `utils/`: 工具函数
- `config/`: 配置文件
- `main.go`: 主函数

## 快速开始

1. 安装依赖

```bash
go mod tidy
```

2. 初始化数据库

```SQL
CREATE DATABASE ai_teach_system;
```

3. 配置数据库

```bash
cp .env.example .env
```

在 `.env` 文件中配置数据库连接信息

4. 运行项目

```bash
go run main.go
```
