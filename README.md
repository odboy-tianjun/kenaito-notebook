# Kenaito Notebook

一个基于 Go + Wails + Vue 2 + Gin Server 的多层级嵌套 Markdown 笔记应用。

> [!WARNING]
> **本项目为纯 AI 开发**
>
> 本项目的全部代码、文档与提交内容均由 AI 生成，未经人工逐行审查。使用时请自行评估风险，生产环境部署前务必进行充分的安全审计与测试。

## 📖 项目简介

Kenaito Notebook 是一款轻量级、高性能的 Markdown 笔记应用，支持无限层级嵌套的文件夹结构，帮助你高效组织和管理笔记内容。采用前后端分离架构，Wails 桌面客户端通过 REST API 与独立的 Gin 服务端通信。支持用户注册登录，笔记数据按用户隔离。

## 🏗️ 架构概览

```
┌─────────────────────────────────┐
│      Wails 桌面客户端            │
│  app.go (HTTP 客户端 + JWT)      │
│  + Wails 原生功能 (文件对话框等)    │
│  Vue 2 + Element UI + mavon     │
└──────────┬──────────────────────┘
           │ HTTP REST API (Bearer Token)
           ▼
┌─────────────────────────────────┐
│      Server (Gin)  :8080        │
│  /api/auth/*   (公开)            │
│  /api/notes/*  (需登录)           │
│  /api/files/*  (需登录)           │
│  /api/export/* (需登录)           │
├─────────────────────────────────┤
│  AuthService (bcrypt + JWT)     │
│  Database (SQLite)              │
│  FileService                    │
│  ExportService                  │
└─────────────────────────────────┘
```

## 🛠️ 技术栈

| 层级           | 技术                          | 版本      |
|--------------|-----------------------------|---------|
| 桌面框架        | Wails                       | v2.12.0 |
| 客户端后端      | Go                          | 1.25+   |
| 服务端框架      | Gin                         | v1.10.0 |
| 数据库         | SQLite (modernc.org/sqlite) | v1.53.0 |
| 密码哈希        | bcrypt (golang.org/x/crypto)| v0.x    |
| 前端          | Vue                         | 2.7.x   |
| UI 框架       | Element UI                  | 2.15.x  |
| Markdown 编辑 | mavon-editor                | 2.10.x  |
| Markdown 解析 | goldmark                    | v1.8.x  |

## ✨ 功能特性

### 用户系统

- 🔐 **用户注册/登录** - 用户名 + 密码 + 邮箱注册，JWT Token 认证
- 🔑 **安全存储** - 密码 bcrypt 加密，Token 7 天有效期
- 🚪 **登录/注销** - 优雅的登录界面，一键注销

### 核心功能

- 📁 **多层级嵌套** - 支持无限层级的文件夹和笔记组织
- ✏️ **Markdown 编辑** - 实时预览、语法高亮、工具栏快捷操作
- 🔍 **全文搜索** - 支持标题和内容的模糊搜索
- 🏷️ **标签管理** - 为笔记添加自定义标签分类
- 💾 **自动保存** - 编辑内容 500ms 防抖自动保存
- 🗑️ **回收站** - 软删除 + 恢复 + 永久删除

### 导出功能

- 📤 **HTML 导出** - 生成格式完整的 HTML 页面
- 📝 **Markdown 导出** - 导出原始 Markdown 文件
- 📦 **ZIP 导出** - 文件夹/全部笔记打包导出

### Markdown 增强

- 🎨 **代码高亮** - 支持多种编程语言的语法高亮
- 📊 **表格支持** - Markdown 表格渲染
- 🖼️ **图片上传** - 支持粘贴和上传图片，本地文件服务存储
- 🔗 **链接处理** - 自动识别和处理链接
- 📐 **数学公式** - KaTeX 数学公式支持

## 📦 安装与使用

### 环境要求

- **Go** >= 1.25
- **Node.js** >= 18
- **Wails CLI** v2.x

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 克隆项目

```bash
git clone <repository-url>
cd kenaito-notebook
```

### 安装依赖

```bash
# 安装客户端前端依赖
cd frontend && npm install && cd ..

# 安装 Server 依赖（会自动下载）
cd server && go mod tidy && cd ..
```

### 启动

> ⚠️ 必须先启动 Server，再启动 Wails 客户端。

```bash
# 1. 启动 Server（默认监听 :8080）
cd server
go run .

# 2. 新开终端，启动 Wails 客户端
cd kenaito-notebook
wails dev
```

### 构建生产版本

```bash
# 构建 Server
cd server && go build -o kenaito-notebook-server . && cd ..

# 构建 Wails 客户端
wails build
```

构建产物：

```
server/kenaito-notebook-server       # Server 可执行文件
build/bin/kenaito-notebook.app       # macOS 客户端
build/bin/kenaito-notebook.exe       # Windows 客户端
build/bin/kenaito-notebook           # Linux 客户端
```

## 📁 项目结构

```
kenaito-notebook/
├── app.go                      # Wails 客户端 (HTTP 客户端 + 原生功能)
├── main.go                     # Wails 入口
├── go.mod                      # 客户端模块 (仅依赖 wails/v2)
├── wails.json                  # Wails 项目配置
│
├── models/                     # 共享数据模型
│   └── note.go                 # Note / NoteTree 定义
│
├── server/                     # 独立服务端 (Gin)
│   ├── main.go                 # Gin 启动入口 (路由注册 + 认证中间件)
│   ├── config.go               # 配置管理
│   ├── go.mod                  # 服务端模块
│   ├── models/                 # 数据模型
│   │   ├── note.go             # Note / NoteTree
│   │   └── user.go             # User 用户模型
│   ├── services/               # 业务服务层
│   │   ├── database.go         # SQLite 数据库服务 (笔记 + 用户表)
│   │   ├── fileservice.go      # 文件/图片上传服务
│   │   ├── export.go           # 导出服务 (HTML/Markdown/ZIP)
│   │   └── auth.go             # 认证服务 (bcrypt + JWT)
│   └── handlers/               # REST API 处理器
│       ├── note.go             # 笔记 CRUD API
│       ├── file.go             # 文件上传 API
│       ├── export.go           # 导出 API
│       └── auth.go             # 认证 API (注册/登录/中间件)
│
├── frontend/                   # 前端项目
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── src/
│       ├── main.js
│       ├── App.vue             # 主组件 (登录态判断 + 布局)
│       ├── style.css
│       └── components/
│           ├── LoginView.vue   # 登录/注册页面
│           ├── NoteTree.vue    # 树形导航组件
│           └── TrashView.vue   # 回收站视图
│
├── build/                      # 构建输出
│   └── bin/
│
└── README.md
```

## 🔧 数据存储

### 数据库位置

SQLite 数据库和资源文件存储在用户主目录下：

```
~/.kenaito-notebook/
├── notes.db                   # SQLite 数据库
└── assets/                    # 图片资源
    └── 2026/06/               # 按日期分组
        ├── a1b2c3d4.png
        └── e5f6g7h8.jpg
```

### 数据库表结构

```sql
-- 用户表
CREATE TABLE users
(
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    username       TEXT NOT NULL UNIQUE,
    email          TEXT NOT NULL,
    hashed_password TEXT NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_username ON users (username);

-- 笔记表
CREATE TABLE notes
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id  INTEGER       DEFAULT 0, -- 父级 ID，0 为根目录
    title      TEXT NOT NULL DEFAULT '',
    content    TEXT          DEFAULT '',
    sort_order INTEGER       DEFAULT 0,
    is_folder  BOOLEAN       DEFAULT FALSE,
    tag        TEXT          DEFAULT '',
    created_at DATETIME      DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME      DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME      DEFAULT NULL  -- 软删除标记
);

CREATE INDEX idx_notes_parent_id ON notes (parent_id);
CREATE INDEX idx_notes_sort_order ON notes (sort_order);
CREATE INDEX idx_notes_deleted_at ON notes (deleted_at);
```

## 📡 API 文档

### Server REST API

Server 默认运行在 `http://localhost:8080`，所有端点前缀为 `/api`。

#### 认证

| 方法   | 路径                  | 请求体                                     | 返回             | 说明       |
|------|---------------------|------------------------------------------|----------------|----------|
| POST | `/api/auth/register`| `{username, password, email}`            | `{token}`      | 用户注册    |
| POST | `/api/auth/login`   | `{username, password}`                   | `{token}`      | 用户登录    |
| GET  | `/api/auth/me`      | - (需 Bearer Token)                      | `{id, username, email, created_at}` | 获取当前用户 |

> 🔒 除 `/api/auth/register` 和 `/api/auth/login` 外，所有 API 需在请求头中携带 `Authorization: Bearer <token>`。

#### 笔记管理

| 方法     | 路径                       | 请求体                                          | 返回            | 说明       |
|--------|--------------------------|-----------------------------------------------|---------------|----------|
| GET    | `/api/notes/tree`        | -                                             | `[]NoteTree`  | 获取笔记树    |
| GET    | `/api/notes/:id`         | -                                             | `Note`        | 获取单个笔记   |
| POST   | `/api/notes`             | `{parentId, title, isFolder}`                 | `Note`        | 创建笔记/文件夹 |
| PUT    | `/api/notes/:id`         | `{title, content, tag}`                       | `{message}`   | 更新笔记     |
| DELETE | `/api/notes/:id`         | -                                             | `{message}`   | 软删除(移入回收站)|
| GET    | `/api/notes/search?q=xxx`| -                                             | `[]Note`      | 搜索笔记     |

#### 回收站

| 方法     | 路径                    | 返回          | 说明       |
|--------|-----------------------|-------------|----------|
| GET    | `/api/trash`          | `[]Note`    | 获取回收站   |
| POST   | `/api/trash/:id/restore`| `{message}` | 恢复笔记    |
| DELETE | `/api/trash/:id`      | `{message}` | 永久删除    |
| DELETE | `/api/trash`          | `{message}` | 清空回收站   |

#### 文件管理

| 方法   | 路径                   | 请求体                      | 返回        | 说明      |
|------|----------------------|--------------------------|-----------|---------|
| POST | `/api/files/image`   | `{data: base64}`         | `{url}`   | 上传图片    |
| POST | `/api/files/upload`  | `{data: base64, filename}`| `{url}`   | 上传文件    |
| GET  | `/api/files/server-url`| -                        | `{url}`   | 文件服务器地址 |
| GET  | `/api/files/assets-path`| -                       | `{path}`  | 资源目录路径  |

#### 导出

| 方法   | 路径                              | 请求体                      | 返回          | 说明        |
|------|---------------------------------|--------------------------|-------------|-----------|
| POST | `/api/export/html`              | `{title, content}`       | `{filePath}`| 导出 HTML    |
| POST | `/api/export/markdown`          | `{title, content}`       | `{filePath}`| 导出 Markdown|
| POST | `/api/export/folder-zip/:folderId`| -                       | `{filePath}`| 文件夹导出 ZIP |
| POST | `/api/export/all-zip`           | -                        | `{filePath}`| 全部导出 ZIP  |

### Wails 客户端 API

Wails 客户端通过 `window.go.main.App.*` 调用，方法名与 Server API 对应：

```javascript
// 用户认证
const token = await window.go.main.App.Register(username, password, email)
const token = await window.go.main.App.Login(username, password)
window.go.main.App.Logout()
const user = await window.go.main.App.GetUser()
const loggedIn = await window.go.main.App.IsLoggedIn()

// 获取笔记树
const tree = await window.go.main.App.GetNoteTree()

// 获取单个笔记
const note = await window.go.main.App.GetNote(noteId)

// 创建笔记
await window.go.main.App.CreateNote(parentId, title, isFolder)

// 更新笔记
await window.go.main.App.UpdateNote(id, title, content, tag)

// 删除笔记
await window.go.main.App.DeleteNote(id)

// 搜索笔记
const results = await window.go.main.App.SearchNotes(keyword)

// 保存图片
const url = await window.go.main.App.SaveImage(base64Data)

// 导出
await window.go.main.App.ExportHTML(title, content)
await window.go.main.App.ExportMarkdown(title, content)
```

## 🎨 界面说明

### 登录/注册页面

首次打开应用会显示登录/注册页面，支持：
- 用户名 + 密码登录
- 用户名 + 密码 + 邮箱注册
- 登录/注册模式一键切换
- 表单验证（用户名长度、密码长度、邮箱格式、确认密码）

### 主界面布局

```
┌─────────────────────────────────────────────────────────────┐
│  📝 Kenaito Notebook                    👤 username [退出]  │
├─────────────┬───────────────────────────────────────────────┤
│  🔍 搜索...  │  笔记标题                                      │
├─────────────┤  [导出 ▼]                                      │
│ [新建笔记]   │  [标签]                                        │
│ [新建文件夹]  │ ┌─────────────────────────────────────────────┐
├─────────────┤ │                                             │
│ 📁 文件夹    │ │           Markdown 编辑器                    │
│   📄 笔记1  │ │           (实时预览)                         │
│   📁 子文件夹│ │                                             │
│     📄 笔记2│ │                                             │
│ 📄 笔记3    │ │                                             │
│             │ └─────────────────────────────────────────────┘
└─────────────┴───────────────────────────────────────────────┘
```

### 操作说明

| 操作       | 方式           |
|----------|--------------|
| 登录/注册   | 首页自动显示     |
| 选择笔记     | 单击笔记项        |
| 展开/收起文件夹 | 单击文件夹左侧箭头    |
| 新建笔记     | 点击「新建笔记」按钮   |
| 新建文件夹    | 点击「新建文件夹」按钮  |
| 在文件夹内新建  | 右键文件夹 → 选择操作 |
| 删除笔记     | 右键笔记 → 删除    |
| 搜索       | 在搜索框输入关键词    |
| 导出       | 点击右上角「导出」按钮  |
| 退出登录     | 点击右上角「退出」按钮  |

## 🚀 开发指南

### 开发环境配置

1. **安装 Go**: https://golang.org/dl/
2. **安装 Node.js**: https://nodejs.org/
3. **安装 Wails**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

### 添加新功能

#### 添加 Server API

1. 在 `server/handlers/` 中添加处理器
2. 在 `server/main.go` 的路由组中注册路由（公开接口放 `auth` 组，需登录接口放 `protected` 组）
3. 在 `app.go` 中添加对应的 HTTP 调用方法
4. 在 `frontend/wailsjs/go/main/App.js` 和 `App.d.ts` 中添加绑定声明
5. 前端通过 `window.go.main.App.MethodName()` 调用

#### 认证机制

- 注册：POST `/api/auth/register`，返回 JWT Token，自动保存到客户端
- 登录：POST `/api/auth/login`，返回 JWT Token，自动保存到客户端
- 客户端 `doRequest()` 自动在请求头中添加 `Authorization: Bearer <token>`
- Server 中间件校验 Token 有效性，无效/过期返回 401
- Token 使用 HMAC-SHA256 签名，有效期 7 天

### 代码规范

- Go 代码遵循 `gofmt` 标准
- Vue 代码使用 ESLint 格式化
- 提交信息使用中文描述

## 🐛 常见问题

### Q: 数据库文件在哪里？

A: 数据库存储在 `~/.kenaito-notebook/notes.db`

### Q: 如何备份笔记？

A: 备份 `~/.kenaito-notebook/` 目录即可

### Q: 图片上传后存储在哪里？

A: 图片存储在 `~/.kenaito-notebook/assets/` 目录，按日期分组

### Q: Server 启动失败怎么办？

A: 确认 8080 端口未被占用。可通过环境变量 `SERVER_PORT=9090` 修改端口。

### Q: Wails 客户端连接不上 Server？

A: 确保 Server 已启动。默认连接 `http://localhost:8080`。

### Q: 忘记密码了怎么办？

A: 目前暂不支持密码重置功能，请重新注册账号。

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📮 联系方式

如有问题或建议，请提交 Issue。
