# AI Teach System

## 项目简介

SZU 毕业设计 —— 面向编程类课程的人工智能辅助学习系统的教师端设计与开发

本系统旨在利用人工智能技术辅助编程类课程的教学，提供代码生成、代码纠错、代码分析等功能，帮助教师更高效地进行教学活动。

## 技术栈

- 后端：Gin + Gorm + MySQL
- AI服务：通义千问API
- 存储：阿里云OSS
- 部署：Docker + Docker Compose

## 功能

- 采用 cron + goroutine 定时异步的方式，自动从 LeetCode 题库抓取题目数据，并同步到数据库中
- 用户认证：JWT认证机制，支持用户注册和登录
- 题目管理：支持按难度、知识点筛选题目，查看题目详情
- AI辅助功能：
  - 代码生成：根据题目要求自动生成最优解答代码
  - 代码纠错：分析用户代码中的错误并提供修正建议
  - 代码分析：对代码进行深度分析，提供知识点讲解
- 课程管理：支持课程详情查看和知识点管理
- 跨域支持：内置CORS中间件，支持前后端分离开发

## 项目结构
- `controllers/`: 控制器，处理HTTP请求
- `models/`: 数据库模型定义
- `routes/`: 路由配置和中间件
- `services/`: 业务逻辑层
- `utils/`: 工具函数和辅助方法
- `config/`: 配置文件和环境变量管理
- `constants/`: 常量定义
- `tasks/`: 定时任务和异步处理
- `tests/`: 单元测试和集成测试
- `cmd/`: 命令行工具
- `main.go`: 应用入口

## 快速开始

1. 安装依赖

```bash
go mod tidy
```

2. 初始化数据库
```sql
CREATE DATABASE ai_teach_system;
 ```

3. 配置环境变量
```bash
cp .env.example .env
 ```

在 .env 文件中配置以下信息：

- 数据库连接信息
- JWT密钥
- 阿里云OSS配置
- 通义千问API密钥
- LeetCode会话信息
4. 运行项目
```bash
go run main.go
 ```

## Docker部署
1. 构建并启动容器
```bash
docker-compose up -d
 ```

2. 查看日志
```bash
docker-compose logs -f
 ```

## 开发指南
### 代码规范
项目使用pre-commit钩子确保代码质量：

- 自动格式化Go代码
- 运行go vet检查常见错误
- 使用golangci-lint进行静态代码分析
安装pre-commit:

```bash
pip install pre-commit
pre-commit install
 ```

### 测试
运行单元测试：

```bash
go test ./tests/...
 ```
