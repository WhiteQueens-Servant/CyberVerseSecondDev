# 服务端部署与操作指南

> 本文件供服务器端 Claude Code 或运维人员参考。

---

## 1. 架构概览

CyberVerse 由三个进程组成：

| 进程 | 端口 | 作用 |
|------|------|------|
| Go API Server | 8080 | REST API + WebSocket + **前端页面** |
| Python Inference | 50051 | gRPC 推理服务（LLM/TTS/ASR/Avatar） |
| LiveKit（可选） | 7880 | WebRTC 信令 |

Go 服务器同时托管前端静态文件：API 请求走 `/api/`、`/ws/`，其他请求返回 `frontend/dist/` 中的文件，找不到的返回 `index.html`（Vue SPA 路由）。

---

## 2. 一键启动（管理端 + 游客端）

```bash
# Step 1: 构建前端（只需执行一次，后续代码变更后重新 build 即可）
cd /path/to/CyberVerse-main/frontend && npm ci && npm run build

# Step 2: 导入 Demo 数据（只需执行一次）
sqlite3 /path/to/CyberVerse-main/data/cyberverse.db < /path/to/seed_jinan_demo.sql

# Step 3: 启动服务
cd /path/to/CyberVerse-main
make server       # Go 服务器 :8080（同时提供 API + 前端页面）
make inference    # Python 推理服务 :50051（AI 对话能力）
```

**为什么只需 `make server` 就能同时启动管理端和游客端？**

Go 服务器内置了前端静态文件托管（`spaFallback` 中间件）。请求到达 :8080 时：
- `/api/`、`/ws/` 开头 → 走 API 路由（景点 CRUD、会话管理等）
- 其他路径 → 从 `frontend/dist/` 返回对应文件；找不到则返回 `index.html`，由 Vue Router 在浏览器端决定渲染哪个页面

因此不需要单独的前端服务器（Nginx/Vite），一个 `make server` 同时服务管理端页面（`/attractions`、`/routes`、`/dashboard` 等）和游客端页面（`/scenic`）。

> `make frontend`（Vite 开发服务器 :5173）仅用于本地开发时的热更新，生产环境不需要。

---

## 3. 页面入口一览

所有页面通过 Go 服务器的 8080 端口访问：

### 游客端

| URL | 页面 | 说明 |
|-----|------|------|
| `http://<host>:8080/scenic` | 景区入口页 | 展示景点/路线/导游角色，点击"开始对话"进入会话 |
| `http://<host>:8080/session/<id>` | 对话页 | 与 AI 导游的文字/语音对话 |

### 管理端

| URL | 页面 | 说明 |
|-----|------|------|
| `http://<host>:8080/` | 首页 | CyberVerse 主入口 |
| `http://<host>:8080/characters` | 角色列表 | 管理所有 AI 角色 |
| `http://<host>:8080/characters/<id>/edit` | 角色编辑 | 编辑角色，含景区配置 |
| `http://<host>:8080/attractions` | 景点管理 | CRUD 景点数据 |
| `http://<host>:8080/routes` | 路线管理 | CRUD 路线数据 |
| `http://<host>:8080/dashboard` | 数据看板 | 会话统计、热门问题、情感分布 |
| `http://<host>:8080/sessions` | 会话日志 | 历史会话记录 |
| `http://<host>:8080/reports` | 分析报告 | 生成/查看分析报告 |
| `http://<host>:8080/settings` | 系统配置 | LLM/TTS/Avatar 等全局配置 |

---

## 4. 首次配置流程

### 4.1 创建导游角色

1. 打开 `http://<host>:8080/characters` → 新建角色
2. 填写名称、描述、头像
3. 景区配置区域：
   - 数字人类别 → 选择"导游"
   - 景区人设提示词 → `"你是泉城济南的专属 AI 导游..."`
   - 推荐路线 → 勾选路线
4. 保存

### 4.2 配置知识库（推荐）

1. 角色编辑页 → 知识库区域
2. 上传文档（`.md`、`.txt`、`.pdf`、`.docx`、`.json`）：
   - 景区介绍文档：各景点详细介绍
   - FAQ 文档：20-30 条常见问答对
3. 等待 RAG 索引完成

### 4.3 验证

1. `http://<host>:8080/scenic` → 确认景点/路线/导游显示正常
2. 点击"开始对话" → 发送"推荐一条路线" → 确认 AI 回复包含路线卡片

---

## 5. 常见问题

### Q: 访问 /scenic 返回 404？

**原因**：`frontend/dist/` 不存在或未构建。

**解决**：`cd frontend && npm ci && npm run build`，重启 Go 服务器。

### Q: 页面能打开但 API 调用失败？

**排查**：
```bash
curl -s http://localhost:8080/api/v1/health   # Go 服务器是否运行
```

### Q: /scenic 显示"暂无可用的导游角色"？

**原因**：没有 `scenic_category=guide` 的角色。按照 4.1 创建。

### Q: 对话页 AI 不回复？

**原因**：Python 推理服务未启动。
```bash
make inference   # 启动推理服务
```
