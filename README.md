# Today Eat 后端服务 🍽️

## 项目简介 📚
Today Eat 后端是为微信小程序“今天吃什么”提供数据支持和业务逻辑的服务端，基于 Go 语言开发，采用 Gin 框架，支持个性化菜品推荐、用户管理、聊天互动等功能。

## 主要功能 ✨
- 微信小程序用户微信登录、信息管理
- 菜品信息增删查改与个性化推荐
- 聊天机器人 WebSocket 支持
- 用户点赞、评分、推荐历史、定制推荐等
- RESTful API 设计，接口文档清晰

## 技术栈 🛠️
- **Go 1.18+**
- **Gin** Web 框架
- **MySQL** 数据库
- **Gorilla WebSocket** 实现聊天
- **Session/Cookie** 用户状态管理

## 目录结构 📂
- backend/
  - main.go              入口文件
  - config/              配置文件目录
  - user/                用户相关接口
  - recommend/           推荐与菜品相关接口
  - chat/                聊天相关接口
  - data/                静态资源（如头像）

## 快速启动 🚀
1. 安装 Go 1.18+ 和 MySQL 数据库。
2. 配置数据库和微信/AI参数（见 config/ 目录下示例配置文件）。
3. 安装依赖：
   ```bash
   go mod tidy
   ```
4. 启动服务：
   ```bash
   go run backend/main.go
   ```
5. 服务默认监听 8080 端口。

## 常用接口文档 📖
- 微信登录：`POST /api/user/wxlogin`
- 获取菜品：`GET /api/dishes`
- 随机推荐：`GET /api/dish/random?user_id=xxx`
- 聊天 WebSocket：`GET /api/chat/ws`
- 用户点赞：`POST /api/like/like`
- 评分接口：`POST /api/rating`
- 更多接口详见代码注释与接口文档

---
如有问题请联系开发者。🤝

前端仓库：[TodayEat今天吃什么——微信小程序前端](https://github.com/RuiHanna/Today-Eat-frontend)