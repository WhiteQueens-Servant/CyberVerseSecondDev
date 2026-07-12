# 服务器端综合验证指南

> **目标读者**：服务器远端 Claude Code
> **项目**：CyberVerse 景区 AI 数字人（第十五届软件杯 A5 赛题）
> **分支**：`feature/scenic-guide`
> **编写日期**：2026-06-18

---

## 一、架构总览

```
浏览器
  │
  ├─ 游客端: /scenic → ScenicLandingPage → /session/:id (light theme)
  ├─ 管理端: /admin/* → AdminLayout (sidebar) → dashboard/characters/attractions/...
  └─ 根路径: / → redirect /scenic
  │
  │ HTTP :8080 (REST + WS + 前端静态文件托管)
  ▼
Go API Server
  ├─ api/          REST 路由 (+ spaFallback 托管 frontend/dist/)
  ├─ scenic/       景区数据层 (SQLite: data/cyberverse.db)
  ├─ orchestrator/ 会话编排 + 情感分析 + LLM 上下文注入
  ├─ character/    角色磁盘存储 (data/characters/<id>/)
  ├─ ws/           WebSocket hub
  └─ inference/    gRPC 客户端 → Python :50051
  │
  │ gRPC :50051
  ▼
Python Inference Server
  ├─ voice_llm/persona_agent.py  PersonaAgent (工具调用: 景点/路线/RAG)
  ├─ rag/engine.py               RAG 向量检索 (Chroma)
  └─ plugins/                    LLM/TTS/ASR/Avatar 插件
```

---

## 二、SQL 脚本执行

### 2.1 数据库初始化（自动）

数据库表在 Go 服务器首次启动时**自动创建**，无需手动执行 SQL。

**位置**：`server/internal/scenic/db.go` → `initSchema()`

**自动创建的 6 张表**：

```sql
attractions       — 景点 (id, name, description, category, location, image_url, tags)
routes            — 路线 (id, name, description, duration, difficulty, tags)
route_steps       — 路线站点 (route_id, attraction_id, step_order, duration_minutes, highlight)
sessions          — 会话 (id, character_id, started_at, ended_at, duration_s, turn_count, sentiment)
conversation_logs — 对话日志 (session_id, role, content, normalized_content, timestamp)
reports           — 分析报告 (id, status, date_from, date_to, character_id, content)
```

**验证**：

```bash
# 启动服务器后检查
sqlite3 data/cyberverse.db ".tables"
# 预期：attractions  conversation_logs  reports  route_steps  routes  sessions
```

### 2.2 Demo 数据导入（手动）

**SQL 文件**：`docs/second-development-plan/seed_jinan_demo.sql`

**内容**：10 个济南景点 + 4 条游览路线（含 route_steps）

**执行命令**：

```bash
cd /path/to/CyberVerse-main
mkdir -p data
sqlite3 data/cyberverse.db < docs/second-development-plan/seed_jinan_demo.sql
```

**验证**：

```bash
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM attractions;"
# 预期：10

sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM routes;"
# 预期：4

sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM route_steps;"
# 预期：11

sqlite3 data/cyberverse.db "SELECT name, category FROM attractions;"
# 预期：趵突泉/历史古迹, 大明湖/自然风光, 千佛山/历史古迹, ...
```

**注意**：
- SQL 脚本使用 `INSERT INTO`（非 `INSERT OR REPLACE`），重复执行会因主键冲突报错
- 如需重置：先 `DELETE FROM route_steps; DELETE FROM routes; DELETE FROM attractions;` 再重新导入
- 表结构由 Go 服务器自动创建，SQL 脚本只插入数据

### 2.3 其他数据库

```bash
# SubAgent 任务数据库（原 CyberVerse，自动创建）
sqlite3 data/tasks/tasks.db ".tables"
# 预期：artifacts  schema_migrations  task_events  tasks

# LangGraph 检查点（自动创建）
ls data/tasks/langgraph_checkpoints.db
```

---

## 三、已实现功能清单

### 3.1 Go 后端（Phase 1-3 + T-23）

| 功能 | 文件 | API 端点 | 状态 |
|------|------|---------|------|
| 景点 CRUD | `scenic/attraction.go` + `api/scenic_handler.go` | GET/POST/PUT/DELETE `/api/v1/attractions` | ✅ |
| 路线 CRUD | `scenic/route.go` + `api/scenic_handler.go` | GET/POST/PUT/DELETE `/api/v1/routes` | ✅ |
| 会话+对话日志 | `scenic/session.go` + `api/analytics_handler.go` | GET `/api/v1/analytics/sessions`, GET `/api/v1/analytics/sessions/{id}` | ✅ |
| Dashboard 聚合 | `scenic/dashboard.go` + `api/analytics_handler.go` | GET `/api/v1/analytics/dashboard` | ✅ |
| 报告 CRUD | `scenic/report.go` + `api/analytics_handler.go` | GET/POST/DELETE `/api/v1/reports` | ✅ |
| 角色扩展字段 | `character/store.go` | — | ✅ |
| 会话生命周期 DB 钩子 | `orchestrator/orchestrator.go` | — | ✅ |
| 对话日志双写 | `orchestrator/orchestrator.go` (`syncSessionToDB`) | — | ✅ |
| LLM 上下文注入 | `orchestrator/orchestrator.go` (`buildScenicContext`) | — | ✅ |
| 情感分析 | `orchestrator/orchestrator.go` (`AnalyzeSessionSentiment`) | — | ✅ |
| 前端静态文件托管 | `api/router.go` (`spaFallback`) | — | ✅ |

### 3.2 Python 推理层（Phase 4）

| 功能 | 文件 | 状态 |
|------|------|------|
| `get_attractions` 工具 | `persona_agent.py` | ✅ |
| `get_routes` 工具 | `persona_agent.py` | ✅ |
| `get_attraction_detail` 工具 | `persona_agent.py` | ✅ |
| `retrieve_character_knowledge` 工具（RAG） | `persona_agent.py` | ✅ |
| RAG 预检索（每条消息自动触发） | `persona_agent.py` (`_rag_response_instructions`) | ✅ |

### 3.3 前端（Phase 5-6 + T-22 + T-27）

| 功能 | 文件 | 路由 | 状态 |
|------|------|------|------|
| 景区入口页 | `ScenicLandingPage.vue` | `/scenic` | ✅ |
| 对话页双主题 | `SessionPage.vue` | `/session/:id` | ✅ |
| ChatPanel 浅色主题 | `ChatPanel.vue` | — | ✅ |
| ScenicCard 展开/折叠 | `ScenicCard.vue` | — | ✅ |
| 角色编辑页景区扩展 | `CharacterEditPage.vue` | `/admin/characters/:id/edit` | ✅ |
| 知识库管理 | `KnowledgeSourceManager.vue` | 嵌入角色编辑页 | ✅ |
| 景点管理页 | `AttractionManagePage.vue` | `/admin/attractions` | ✅ |
| 路线管理页 | `RouteManagePage.vue` | `/admin/routes` | ✅ |
| 数据大屏 | `DashboardPage.vue` | `/admin/dashboard` | ✅ |
| 会话日志页 | `SessionLogPage.vue` | `/admin/sessions` | ✅ |
| 分析报告页 | `ReportPage.vue` | `/admin/reports` | ✅ |
| 系统设置页 | `SettingsPage.vue` | `/admin/settings` | ✅ |
| 管理端侧边栏布局 | `AdminLayout.vue` | `/admin` | ✅ |
| 根路径重定向 | `router/index.ts` | `/ → /scenic` | ✅ |

---

## 四、核心链路详解

### 4.1 游客发起对话的完整链路

```
① 浏览器访问 / → 重定向 /scenic → ScenicLandingPage
② 页面加载：并行请求 getCharacters() + listAttractions() + listRoutes()
③ 渲染：导游角色卡片 + 景点列表 + 路线列表
④ 用户点击"开始语音讲解"
⑤ POST /api/v1/sessions { character_id, mode:"omni" }
⑥ Go 服务器：
   - charStore.Get(characterID) → 从磁盘读取最新 character.json
   - sessionMgr.Create() → 创建会话
   - HydrateVoiceDialogContext() → 加载历史对话
⑦ 返回 session_id → 浏览器跳转 /session/:id
⑧ SessionPage 判断 isScenicMode (returnPath === '/scenic') → 浅色主题
⑨ WebSocket 建立 → client_media_ready
⑩ orchestrator 组装 system_prompt：
    scenic_system_prompt（如有，覆盖 system_prompt）
    + buildScenicContext()（景点/路线数据，从 SQLite 实时查询）
⑪ 调用 Python 推理服务生成开场白
⑫ 用户语音/文字输入 → ASR → LLM（含 RAG 预检索 + 工具调用）→ TTS → 数字人
```

### 4.2 系统提示词组装链路

```
character.json 各字段的去向：

scenic_system_prompt ──→ 非空时覆盖 system_prompt
system_prompt ─────────→ 基础人设（被覆盖或保留）
name ─────────────────→ 拼入开场白提示词 + character_manifest
speaking_style ────────→ 拼入开场白提示词 + character_manifest
personality ───────────→ 拼入 characterSystemPrompt
description ───────────→ 拼入 characterSystemPrompt
welcome_message ───────→ 拼入开场白提示词（作为参考）
voice_type ────────────→ VoiceLLMSessionConfig.Voice → TTS 音色
scenic_category ───────→ 前端 ScenicLandingPage 筛选导游角色
recommended_routes ────→ 前端展示，不注入 LLM
knowledge/ ────────────→ Chroma 向量库索引 → RAG 动态检索注入

最终 system_prompt = scenic_system_prompt（或 system_prompt）
                    + "\n\n" + buildScenicContext()
                    // buildScenicContext() 实时查询 SQLite 景点+路线
```

### 4.3 用户提问时的工具调用链路

```
用户："千佛山有什么好玩的？开放时间是什么？"
  │
  ├─ [自动] RAG 预检索（persona_agent.py:655）
  │   → Chroma 向量库搜索 → 命中知识库文档 → 注入 LLM 上下文
  │
  ├─ [LLM 决策] 调用 get_attractions()
  │   → httpx GET http://localhost:8080/api/v1/attractions
  │   → Go Server 查 SQLite attractions 表 → 返回 JSON
  │
  ├─ [LLM 决策] 调用 get_attraction_detail(attraction_id="xxx")
  │   → httpx GET http://localhost:8080/api/v1/attractions/xxx
  │   → 返回景点详情（含 description, opening_hours, ticket_price 等）
  │
  └─ [LLM 生成] 综合所有信息，流式输出语音+文字回答
```

**关键点**：
- RAG 预检索是**自动触发**的（每条消息），不需要 LLM 决定
- 景点/路线查询是**按需的**（LLM 自主决定是否调用工具）
- 工具可以**串联调用多次**
- Go API 地址通过 `GO_API_BASE` 环境变量配置，默认 `http://localhost:8080`

### 4.4 响应返回链路

```
Python Inference → gRPC stream: VoiceLLMOutput
  │
  ├─ 文字链路：transcript → Go WS broadcastJSON → 浏览器 ChatPanel 流式渲染
  │
  ├─ 语音链路：audio PCM → WebRTC → 浏览器扬声器播放
  │
  └─ 数字人链路：audio → avatar 模型 → 唇形同步视频 → RTP → 浏览器视频播放

前端 ChatPanel 额外处理：
  - 正则匹配 assistant 消息中的 "推荐路线""景点" 等关键词
  - 匹配到 → 渲染 ScenicCard（可展开/折叠的内嵌卡片）
  - 未匹配到 → 纯文本渲染
```

### 4.5 会话结束情感分析链路

```
用户断开连接
  → OnSessionEnd 回调
    ├─ PersistSessionConversation()     ← 同步，写文件
    ├─ scenicDB.EndSession()            ← 同步，更新 duration_s/turn_count
    └─ go orch.AnalyzeSessionSentiment() ← 异步 goroutine
        ├─ 取最后 20 条消息
        ├─ 构建情感分析 prompt
        ├─ 调用 inference.GenerateLLMStream() (temperature=0.1)
        ├─ 解析: positive/neutral/negative
        └─ scenicDB.UpdateSentiment() 写回 sessions 表
```

---

## 五、验收标准与命令

### 5.1 基础环境验证

```bash
# Go 编译
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 前端构建
cd /path/to/CyberVerse-main/frontend
npm ci && npm run build
# 预期：构建成功

# Python 语法检查
cd /path/to/CyberVerse-main
python -c "from inference.plugins.voice_llm.persona_agent import PERSONA_TOOL_DEFINITIONS; print(len(PERSONA_TOOL_DEFINITIONS))"
# 预期：8
```

### 5.2 启动验证

```bash
# Step 1: 构建前端
cd /path/to/CyberVerse-main/frontend && npm ci && npm run build

# Step 2: 导入 Demo 数据
cd /path/to/CyberVerse-main
mkdir -p data
sqlite3 data/cyberverse.db < docs/second-development-plan/seed_jinan_demo.sql

# Step 3: 启动 Go 服务器
make server
# 预期日志：Scenic database initialized: db=data/cyberverse.db

# Step 4: 启动 Python 推理服务
make inference
```

### 5.3 API 端点验证

```bash
# 健康检查
curl -s http://localhost:8080/api/v1/health
# 预期：{"status":"ok",...}

# 景点列表（应有 10 条 Demo 数据）
curl -s http://localhost:8080/api/v1/attractions | python3 -m json.tool
# 预期：10 个景点

# 路线列表（应有 4 条 Demo 数据）
curl -s http://localhost:8080/api/v1/routes | python3 -m json.tool
# 预期：4 条路线

# Dashboard
curl -s http://localhost:8080/api/v1/analytics/dashboard | python3 -m json.tool
# 预期：5 个聚合指标（初始值为 0）

# 角色列表
curl -s http://localhost:8080/api/v1/characters | python3 -m json.tool
# 预期：现有角色列表

# 现有端点不受影响
curl -s http://localhost:8080/api/v1/characters | python3 -m json.tool
# 预期：正常返回
```

### 5.4 前端页面验证

```bash
# 游客端入口
curl -s http://localhost:8080/scenic | head -5
# 预期：返回 index.html（Vue SPA）

# 管理端入口
curl -s http://localhost:8080/admin | head -5
# 预期：返回 index.html

# 根路径重定向
curl -s -I http://localhost:8080/
# 预期：浏览器端会重定向到 /scenic（SPA 层面，curl 不跟随 JS 重定向）

# 静态资源
curl -s -I http://localhost:8080/assets/index-*.js
# 预期：200 + Content-Type: application/javascript
```

### 5.5 端到端业务流程验证

#### 验证 1：管理端创建景点 → 游客端可见

```bash
# 1. 管理端创建景点
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"测试景点","description":"测试描述","category":"测试","tags":["测试"]}' | python3 -m json.tool
# 记录返回的 id

# 2. 游客端查询
curl -s http://localhost:8080/api/v1/attractions | python3 -m json.tool
# 预期：列表中包含"测试景点"

# 3. 清理
curl -s -X DELETE http://localhost:8080/api/v1/attractions/<id>
```

#### 验证 2：管理员编辑角色 → 游客端对话生效

```bash
# 1. 更新角色的 scenic_system_prompt
curl -s -X PUT http://localhost:8080/api/v1/characters/<char_id> \
  -H 'Content-Type: application/json' \
  -d '{"scenic_category":"guide","scenic_system_prompt":"你是济南泉城的专属导游，用亲切的济南话讲解。"}'
# 预期：返回更新后的角色对象

# 2. 发起对话（通过浏览器访问 /scenic → 开始对话）
# 预期：数字人使用新的景区人设回复

# 3. 清空 scenic_system_prompt
curl -s -X PUT http://localhost:8080/api/v1/characters/<char_id> \
  -H 'Content-Type: application/json' \
  -d '{"scenic_system_prompt":""}'
# 预期：对话回退到原 system_prompt
```

#### 验证 3：对话 → 会话日志 → 情感分析

```bash
# 1. 通过浏览器发起一次对话，发送几条消息后结束

# 2. 查询会话记录
curl -s http://localhost:8080/api/v1/analytics/sessions | python3 -m json.tool
# 预期：包含刚结束的会话

# 3. 查询会话详情（含对话记录）
curl -s http://localhost:8080/api/v1/analytics/sessions/<session_id> | python3 -m json.tool
# 预期：session 元数据 + messages 数组

# 4. 检查情感分析结果（等待几秒让异步 goroutine 完成）
sqlite3 data/cyberverse.db "SELECT id, sentiment FROM sessions ORDER BY started_at DESC LIMIT 1;"
# 预期：sentiment 为 positive/neutral/negative（非全 neutral）

# 5. Dashboard 情感分布
curl -s http://localhost:8080/api/v1/analytics/dashboard | python3 -m json.tool
# 预期：sentiment_distribution 中有非零值
```

#### 验证 4：知识库 RAG 检索

```bash
# 1. 通过管理端角色编辑页上传知识文档
# 2. 发起对话，询问知识库中的内容
# 预期：数字人基于知识库内容回答（而非编造）
```

#### 验证 5：ScenicCard 渲染

```bash
# 1. 发起对话，问"推荐一条路线"
# 预期：ChatPanel 中出现路线卡片（绿色左边框，可展开/折叠）

# 2. 问"有什么好玩的景点"
# 预期：ChatPanel 中出现景点卡片（蓝色左边框）
```

### 5.6 回归验证

```bash
# 原有功能不受影响
curl -s http://localhost:8080/api/v1/health                    # 健康检查
curl -s http://localhost:8080/api/v1/characters                 # 角色列表
curl -s http://localhost:8080/api/v1/settings                   # 系统设置

# WebSocket 对话正常
# 通过浏览器发起语音/文字对话，确认 AI 正常回复

# SubAgent 任务正常
# 发起一个后台任务（如"帮我搜索知乎热榜"），确认任务创建和完成
```

---

## 六、路由结构（最新）

```
/                           → redirect /scenic
/scenic                     → ScenicLandingPage.vue（游客入口）
/session/:id                → SessionPage.vue（对话页）
/kanshan                    → KanshanLandingPage.vue（看山品牌）
/admin                      → AdminLayout.vue（侧边栏壳）
  /admin/dashboard          → DashboardPage.vue
  /admin/characters         → CharacterListPage.vue
  /admin/characters/:id/edit → CharacterEditPage.vue
  /admin/attractions        → AttractionManagePage.vue
  /admin/routes             → RouteManagePage.vue
  /admin/sessions           → SessionLogPage.vue
  /admin/reports            → ReportPage.vue
  /admin/settings           → SettingsPage.vue
  /admin/launch/:id         → LaunchConfigPage.vue
```

**注意**：原 `LandingPage.vue` 已成为死代码（不在路由中），可安全删除。

---

## 七、文件系统结构

```
data/                              # 运行时创建，不提交 git
├── cyberverse.db                  # 主 SQLite（景区+会话）
├── cyberverse.db-wal              # WAL 日志
├── characters/                    # 角色磁盘存储
│   └── <character-id>/
│       ├── character.json         # 角色配置（含 scenic 字段）
│       ├── avatar.*               # 头像
│       ├── images/                # 多张图片
│       ├── models/                # 数字人模型
│       └── knowledge/             # 知识库文件（RAG 索引）
└── tasks/
    ├── tasks.db                   # SubAgent 任务
    ├── artifacts/                 # 任务产出
    └── langgraph_checkpoints.db   # PersonaAgent 状态
```

---

## 八、已知限制与注意事项

1. **ScenicCard 是前端正则解析**：ChatPanel 用正则从 LLM 文字输出中检测路线/景点关键词，不是后端结构化返回。如果 LLM 回复格式不符合预期，卡片不会出现。

2. **情感分析是异步 best-effort**：会话关闭后几秒内完成，失败仅 log，不影响会话正常关闭。

3. **RAG 预检索每条消息触发**：自动从 Chroma 向量库检索，无需 LLM 决定。如果知识库为空，检索结果为空，不影响对话。

4. **景区数据实时注入**：`buildScenicContext()` 每次消息处理时从 SQLite 查询，修改景点/路线后立即生效。

5. **scenic_system_prompt 覆盖关系**：非空时完全替代 system_prompt（不是追加）。

6. **Demo SQL 重复执行会报错**：使用 `INSERT INTO`，主键冲突。需先清空再重新导入。

7. **预存的 TypeScript 错误**：CharacterEditPage.vue 和 SessionLogPage.vue 有 4 个预存 TS 错误（ScenicRoute export 等），不影响构建和运行。

---

## 九、快速验收检查清单

```bash
# □ Go 编译通过
go build -tags livekit ./cmd/cyberverse-server/

# □ 前端构建通过
npm run build

# □ 数据库表自动创建
sqlite3 data/cyberverse.db ".tables"  # 6 张表

# □ Demo 数据导入成功
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM attractions;"  # 10
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM routes;"       # 4

# □ API 端点全部可用
curl -s http://localhost:8080/api/v1/attractions     # 200
curl -s http://localhost:8080/api/v1/routes           # 200
curl -s http://localhost:8080/api/v1/analytics/dashboard  # 200
curl -s http://localhost:8080/api/v1/reports          # 200

# □ 前端页面可访问
curl -s http://localhost:8080/scenic | grep -q "index.html"  # 返回 SPA 入口
curl -s http://localhost:8080/admin | grep -q "index.html"

# □ 对话功能正常
# 浏览器访问 /scenic → 开始对话 → AI 回复

# □ ScenicCard 渲染
# 对话中问"推荐路线" → 出现路线卡片

# □ 情感分析
# 对话结束后 sqlite3 查 sessions.sentiment 非全 neutral

# □ 现有功能不受影响
curl -s http://localhost:8080/api/v1/health  # ok
curl -s http://localhost:8080/api/v1/characters  # 正常
```
