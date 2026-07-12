# Spec 1: 系统级业务 Spec

> 基于 CyberVerse 框架落地"景区导览服务 AI 数字人"，面向第十五届中国软件杯 A5 赛题。
> 一人开发，预计 5-6 周。

---

## 第 1 节：业务范围与功能边界

### 游客端能力

| # | 功能 | 实现方式 | CyberVerse 复用度 |
|---|------|---------|-------------------|
| F1 | 语音对话 | omni 端到端语音（Qwen-Omni） | 100% 复用 |
| F2 | 文本对话 | 混合输入（语音+文本同轮） | 100% 复用 |
| F3 | 智能问答（景区知识） | RAG 检索 + LLM 生成 | 复用，灌入景区文档 |
| F4 | 个性化推荐 | persona prompt + 知识库路线文档 | 复用，prompt 工程 |
| F5 | 拍照识景 | 图片上传 → Qwen-VL API → 返回介绍 | 需新增交互入口 |
| F6 | 后台任务 | PersonaAgent + SubAgent | 100% 复用 |
| F7 | 语音打断 | omni 插件原生支持 | 100% 复用 |

### 管理端能力

| # | 功能 | 实现方式 | CyberVerse 复用度 |
|---|------|---------|-------------------|
| F8 | 知识库管理 | 现有 knowledge API + RAG engine | 100% 复用 |
| F9 | 数字人形象管理 | 现有 character CRUD + persona 配置 | 100% 复用 |
| F10 | 对话日志查看 | **新增**：结构化对话记录存储 + 查询 API | 新增 |
| F11 | 感受度报告 | **新增**：SubAgent 调 LLM 分析对话 → 图表报告 | 复用任务框架，新增分析逻辑 |
| F12 | 数据大屏 | **新增**：聚合统计 API + ECharts 前端 | 新增 |
| F13 | 景点管理 | **新增**：景点 CRUD API + 管理端页面 | 新增 |
| F14 | 路线管理 | **新增**：路线 CRUD API + 管理端页面 | 新增 |

### 基础设施（两端共享）

| # | 能力 | 说明 |
|---|------|------|
| I1 | 对话日志结构化存储 | 会话开始/结束/每轮对话写 SQLite |
| I2 | 会话统计聚合 | 按时段/角色/维度聚合查询 |
| I3 | RAG 准确率保障 | FAQ 问答对 + prompt 约束 + 混合检索（可选） |
| I4 | 情感标注 | 会话结束时批量调用 LLM 判定情感极性，写入会话记录（sessions 表） |

### 明确不在范围

- GPS 定位自动讲解（P1，赛后迭代）
- GraphRAG（当前场景不需要，单跳检索 + FAQ 已足够）
- 专用 NLP 情感分析模型（用 LLM 调用代替，更简单且效果更好）

---

## 第 2 节：数据模型设计

### 2.1 存储布局

```
data/
├── characters/<id>/          ← 现有，不变
│   ├── character.json
│   ├── knowledge/
│   │   ├── sources.json
│   │   ├── sources/
│   │   └── chroma/
│   └── history/
├── tasks/                    ← 现有，不变
│   ├── tasks.db
│   └── artifacts/
├── cyberverse.db             ← 新增：统一数据库（一个库多表）
│   ├── sessions              ← 对话会话
│   ├── conversation_logs     ← 对话日志
│   ├── attractions           ← 景点数据
│   ├── routes                ← 路线数据
│   └── route_steps           ← 路线途经点
└── reports/                  ← 新增：感受度报告产物
    └── <report_id>/
```

### 2.2 对话与会话表（新增，SQLite — cyberverse.db）

感受度报告和数据大屏的共同前置依赖。

```sql
CREATE TABLE sessions (
    id            TEXT PRIMARY KEY,
    character_id  TEXT NOT NULL,
    started_at    TEXT NOT NULL,
    ended_at      TEXT,
    duration_s    INTEGER,
    turn_count    INTEGER DEFAULT 0,
    input_mode    TEXT DEFAULT "voice",
    visitor_id    TEXT,
    sentiment     TEXT                   -- "positive" / "neutral" / "negative"（会话结束时 LLM 批量判定）
);

CREATE INDEX idx_sessions_character ON sessions(character_id);
CREATE INDEX idx_sessions_started ON sessions(started_at);
CREATE INDEX idx_sessions_sentiment ON sessions(sentiment);

CREATE TABLE conversation_logs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id    TEXT NOT NULL,
    character_id  TEXT NOT NULL,
    turn_seq      INTEGER NOT NULL,
    role          TEXT NOT NULL,          -- "user" / "assistant"
    content       TEXT NOT NULL,
    content_type  TEXT DEFAULT "text",    -- "text" / "voice" / "image"
    created_at    TEXT NOT NULL,
    metadata_json TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX idx_log_session ON conversation_logs(session_id);
CREATE INDEX idx_log_character ON conversation_logs(character_id);
CREATE INDEX idx_log_created ON conversation_logs(created_at);
```

**sessions 与 conversation_logs 的关系**：一对多。sessions 一条记录 = 一次对话会话的元数据（何时开始/结束、时长、情感标签）。conversation_logs 一条记录 = 一轮对话的具体内容（谁说了什么）。`sentiment` 是 session 级别字段——会话结束时 LLM 读取该 session 所有 user 消息，判定整体情感极性，写入 sessions 表。

### 2.3 角色模型扩展

在现有 `Character` 结构体上新增字段（JSON 存储，向后兼容）：

```go
type Character struct {
    // ... 现有字段保持不变 ...

    // 景区导览扩展字段
    ScenicCategory    string   `json:"scenic_category,omitempty"`    // "导游" / "讲解员" / "客服"
    RecommendedRoutes []string `json:"recommended_routes,omitempty"` // 推荐路线 ID 列表
}
```

- `RecommendedRoutes` 存储 route ID 列表，引用 cyberverse.db routes 表
- 一个角色可关联多条路线；角色与景点的关联通过路线间接建立（角色 → recommended_routes → routes → route_steps → attractions）

### 2.4 景点数据（新增，SQLite 表）

```sql
CREATE TABLE attractions (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    category    TEXT,                   -- "自然风光" / "历史古迹" / "亲子游乐"
    location    TEXT,                   -- 自然语言描述，如 "景区东部，莲花湖畔"
    image_url   TEXT,
    tags        TEXT,                   -- JSON 数组字符串，如 '["文化","历史"]'
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);
```

### 2.5 路线数据（新增，SQLite 表）

```sql
CREATE TABLE routes (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    duration    TEXT,                   -- "约2小时"
    tags        TEXT,                   -- JSON 数组字符串
    difficulty  TEXT,                   -- "easy" / "moderate"
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE route_steps (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    route_id      TEXT NOT NULL,
    attraction_id TEXT NOT NULL,
    order_num     INTEGER NOT NULL,
    highlight     TEXT,
    stay_minutes  INTEGER,
    FOREIGN KEY (route_id) REFERENCES routes(id),
    FOREIGN KEY (attraction_id) REFERENCES attractions(id)
);

CREATE INDEX idx_route_steps_route ON route_steps(route_id);
```

Route 通过 `route_steps` 表关联 Attraction，一个路线包含多个途经点，每个途经点引用一个景点。

### 2.6 景点元数据 vs 知识库的区别

| | 景点元数据（`attractions` 表） | 景点知识库（`knowledge/`） |
|---|---|---|
| 内容 | 结构化字段：名称、简介、分类、所在位置、标签 | 非结构化文档：景区介绍 PDF、历史典故 TXT、FAQ |
| 用途 | 管理端景点列表展示、路线规划、地图标注 | RAG 检索——数字人回答游客问题时的上下文来源 |
| 格式 | SQLite 表记录 | 文档原文 + Chroma 向量索引 |
| CRUD | 管理端 API 读写 cyberverse.db | 管理端通过现有 knowledge API 上传/删除 |

两者通过景点 ID 关联：管理员创建景点（写 attractions 表），同时可上传该景点的介绍文档到知识库（进 RAG）。

---

---

## 第 3 节：API 契约

### 3.0 设计决策：统一对话数据源

**`cyberverse.db` 作为对话数据的唯一存储**，现有 CyberVerse 的文本文件历史存储被替换。

```
修改前（CyberVerse 现状）：
  orchestrator → 写文本文件 data/characters/<id>/history/*.txt（给 LLM 上下文）
  无分析存储

修改后：
  orchestrator → 只写 cyberverse.db（conversation_logs + sessions 表）
  orchestrator → 从 cyberverse.db 读取最近 N 轮（给 LLM 上下文注入）
  管理端 API  → 从 cyberverse.db 读取（给分析/展示/报告）
```

Session 生命周期由 orchestrator 管理（连接建立 → 开始，离开/超时 → 结束），不是按问答轮次切割。系统是 session-based，不是 user-based。

### 3.1 复用现有 API（零改动）

| 方法 | 路径 | 功能 | 对应功能 |
|------|------|------|---------|
| GET | `/api/v1/health` | 健康检查 | — |
| GET/POST/PUT/DELETE | `/api/v1/characters` | 角色 CRUD | F9 |
| GET/POST/DELETE | `/api/v1/characters/:id/knowledge` | 知识库文档管理 | F8 |
| GET | `/api/v1/settings` | 全局配置读取 | — |
| POST | `/api/v1/settings` | 全局配置写入 | — |
| WS | `/ws` | WebSocket 事件推送 | F1-F7 实时通信 |

> `/api/v1/conversations/:id`（会话历史）保留 API 路径不变，但底层改为从 cyberverse.db 读取，不再读文本文件。

### 3.2 新增 API：景点管理（F13）

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/api/v1/attractions` | 景点列表 | 支持 `?category=` 过滤 |
| GET | `/api/v1/attractions/:id` | 景点详情 | — |
| POST | `/api/v1/attractions` | 创建景点 | 写入 cyberverse.db attractions 表 |
| PUT | `/api/v1/attractions/:id` | 更新景点 | — |
| DELETE | `/api/v1/attractions/:id` | 删除景点 | 同步清理关联路线中的引用 |

### 3.3 新增 API：路线管理（F14）

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/api/v1/routes` | 路线列表 | 支持 `?tags=` 过滤 |
| GET | `/api/v1/routes/:id` | 路线详情 | — |
| POST | `/api/v1/routes` | 创建路线 | 校验 steps 中的 attraction_id 是否存在 |
| PUT | `/api/v1/routes/:id` | 更新路线 | — |
| DELETE | `/api/v1/routes/:id` | 删除路线 | 同步清理角色 recommended_routes 中的引用 |

### 3.4 新增 API：会话与对话日志（F10, F12）

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/api/v1/analytics/sessions` | 会话列表 | 支持 `?character_id=&date_from=&date_to=` 过滤 |
| GET | `/api/v1/analytics/sessions/:id` | 会话详情（含对话记录） | 返回该 session 的逐轮对话 + 元数据 |
| GET | `/api/v1/analytics/dashboard` | 数据大屏聚合数据 | 今日服务人次、热门 TOP10、满意度分布、活跃时段 |
| GET | `/api/v1/analytics/sentiment` | 情感趋势数据 | 按日/周聚合 positive/neutral/negative 占比趋势 |

`/api/v1/analytics/sessions/:id` 通过 query 参数 `?format=` 控制返回视角：

```
GET /api/v1/analytics/sessions/:id?format=llm    → 精简格式（role + content），供 LLM 上下文注入
GET /api/v1/analytics/sessions/:id?format=admin   → 完整格式（会话元数据含 sentiment + 逐轮对话含 timestamp/content_type），供管理端展示
```

原有 `/api/v1/conversations/:id` 路径保留，内部转发到同一查询（format=llm）。

**`dashboard` 返回结构**：

```json
{
  "today_sessions": 128,
  "today_duration_avg_s": 180,
  "week_sessions": 856,
  "hot_questions": [
    {"question": "门票多少钱", "count": 45},
    {"question": "卫生间在哪", "count": 38}
  ],
  "sentiment_distribution": {
    "positive": 72,
    "neutral": 23,
    "negative": 5
  },
  "hourly_distribution": [2, 1, 0, 0, 0, 3, 12, 28, 35, 42, 38, 30, 25, 22, 18, 15, 10, 8, 5, 3, 2, 1, 1, 0]
}
```

### 3.5 新增 API：感受度报告（F11）

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| POST | `/api/v1/reports/generate` | 触发报告生成 | 创建 SubAgent 任务，异步执行 |
| GET | `/api/v1/reports` | 报告列表 | — |
| GET | `/api/v1/reports/:id` | 报告详情 | 返回分析结果 + 图表数据 |

报告生成通过 PersonaAgent SubAgent 异步执行：管理端点击"生成报告"→ Go 端创建任务 → Python SubAgent 读取 cyberverse.db 对话日志 → 调 LLM 分析 → 产物写入 `data/reports/`。

### 3.6 新增 PersonaAgent 工具（F4, F5）

现有 4 个工具保持不变，新增 2 个：

| 工具名 | 参数 | 作用 | 约束 |
|--------|------|------|------|
| `get_routes` | `tags?: string[]` | 获取当前角色的推荐路线 | **必须先过滤 `RecommendedRoutes`，再按 tags 二次过滤** |
| `recognize_scenic_spot` | `image_base64: string` | 拍照识景 | 调 Qwen-VL API 返回景点介绍 |

**`get_routes` 完整逻辑**：

```
输入: tags=["亲子"]（可选）
  ↓
1. Go 端查询 cyberverse.db：读取当前角色的 RecommendedRoutes
2. 从 routes 表加载对应路线
3. 如果 tags 非空 → 过滤出 tags 匹配的路线
4. 返回过滤结果
```

如果游客不指定偏好，直接问"有什么路线"，则跳过第 3 步，返回角色的全部推荐路线。

**`recognize_scenic_spot` 逻辑**：

```
输入: image_base64（游客拍照）
  ↓
1. 调用 Qwen-VL API（多模态模型，满足赛题"至少1个多模态大模型"要求）
2. Prompt: "这张照片中的景点是什么？请介绍该景点的历史、特色和游览建议。"
3. 返回景点介绍文本
4. 同时查询 cyberverse.db attractions 表尝试匹配已知景点，如果命中则附加结构化信息
```

### 3.7 API 与前端页面的对应关系

| 页面 | 调用的 API |
|------|-----------|
| 游客端 - 对话页 | WS（实时通信） + 会话由 orchestrator 自动管理 |
| 游客端 - 拍照识景 | WS（传递图片） → PersonaAgent `recognize_scenic_spot` |
| 管理端 - 景点管理页 | CRUD `/attractions` |
| 管理端 - 路线管理页 | CRUD `/routes` |
| 管理端 - 知识库管理页 | 现有 `/characters/:id/knowledge` |
| 管理端 - 角色管理页 | 现有 `/characters`（含新增的 scenic_category/recommended_routes 字段） |
| 管理端 - 数据大屏页 | GET `/analytics/dashboard` |
| 管理端 - 感受度报告页 | POST `/reports/generate` + GET `/reports` |
| 管理端 - 对话日志页 | GET `/analytics/sessions` + GET `/analytics/sessions/:id` |

---

## 第 4 节：后端扩展点

### 4.0 设计决策修正

基于前三节讨论，以下修正纳入本节：

**修正 1：统一数据库**。景点和路线数据也存入 SQLite，与对话日志共用同一个数据库（`cyberverse.db`），一个库多表管理。去除 JSON 文件存储。

```
cyberverse.db
├── sessions           ← 对话会话
├── conversation_logs  ← 对话日志
├── attractions        ← 景点数据
├── routes             ← 路线数据
└── route_steps        ← 路线途经点
```

**修正 2：情感标注时机**。不在每轮对话后触发 LLM，改为会话结束时批量分析。一个 session 只需 1 次额外 LLM 调用。

**修正 3：get_routes 工具处理器简化**。Go 端新增 `GET /api/v1/character-routes` 接口，内部完成 RecommendedRoutes 过滤 + tags 过滤。Python 工具处理器只需一次 HTTP 调用，不需要知道 RecommendedRoutes 的存在。

### 4.1 Go 服务器扩展

#### 4.1.1 新增模块

| 模块 | 路径 | 职责 | 参考模式 |
|------|------|------|---------|
| `analytics/` | `server/internal/analytics/` | cyberverse.db 读写（sessions, conversation_logs, attractions, routes, route_steps） | 复制 `agenttask/store.go` 的 SQLite 模式 |

景点和路线不再独立建模块，统一由 `analytics/` 包管理数据库表。

#### 4.1.2 修改现有模块

**`server/internal/orchestrator/orchestrator.go`**（核心改动）

orchestrator 的对话读写路径从文本文件切换到 SQLite：

```
改动点 1：会话开始
  现状：无记录
  改为：INSERT INTO sessions (id, character_id, started_at, ...)

改动点 2：每轮对话写入
  现状：append 到 data/characters/<id>/history/session_*.txt
  改为：INSERT INTO conversation_logs (session_id, character_id, turn_seq, role, content, ...)

改动点 3：LLM 上下文加载
  现状：读取文本文件最后 N 行
  改为：SELECT * FROM conversation_logs WHERE session_id=? ORDER BY turn_seq DESC LIMIT N

改动点 4：会话结束
  现状：无记录
  改为：UPDATE sessions SET ended_at=?, duration_s=?, turn_count=? WHERE id=?

改动点 5：情感标注（会话结束时批量触发）
  现状：无
  改为：session 结束时，读取该 session 所有 user 消息 → 一次 LLM 调用判断整体情感极性
        → UPDATE sessions SET sentiment=? WHERE id=?
```

**`server/internal/character/store.go`**（小改动）

在 `Character` 结构体中新增 3 个字段：

```go
ScenicCategory    string   `json:"scenic_category,omitempty"`
RecommendedRoutes []string `json:"recommended_routes,omitempty"`
```

**`server/internal/api/router.go`**（路由注册）

新增路由：

```go
// 景点管理（F13）
mux.HandleFunc("GET /api/v1/attractions", r.handleListAttractions)
mux.HandleFunc("GET /api/v1/attractions/{id}", r.handleGetAttraction)
mux.HandleFunc("POST /api/v1/attractions", r.handleCreateAttraction)
mux.HandleFunc("PUT /api/v1/attractions/{id}", r.handleUpdateAttraction)
mux.HandleFunc("DELETE /api/v1/attractions/{id}", r.handleDeleteAttraction)

// 路线管理（F14）
mux.HandleFunc("GET /api/v1/routes", r.handleListRoutes)
mux.HandleFunc("GET /api/v1/routes/{id}", r.handleGetRoute)
mux.HandleFunc("POST /api/v1/routes", r.handleCreateRoute)
mux.HandleFunc("PUT /api/v1/routes/{id}", r.handleUpdateRoute)
mux.HandleFunc("DELETE /api/v1/routes/{id}", r.handleDeleteRoute)

// 角色推荐路线（供 PersonaAgent 工具调用）
mux.HandleFunc("GET /api/v1/character-routes", r.handleCharacterRoutes)

// 分析统计（F10, F12）
mux.HandleFunc("GET /api/v1/analytics/sessions", r.handleListSessions)
mux.HandleFunc("GET /api/v1/analytics/sessions/{id}", r.handleGetSessionDetail)
mux.HandleFunc("GET /api/v1/analytics/dashboard", r.handleDashboard)
mux.HandleFunc("GET /api/v1/analytics/sentiment", r.handleSentimentTrend)

// 报告（F11）
mux.HandleFunc("POST /api/v1/reports/generate", r.handleGenerateReport)
mux.HandleFunc("GET /api/v1/reports", r.handleListReports)
mux.HandleFunc("GET /api/v1/reports/{id}", r.handleGetReport)
```

#### 4.1.3 Go 端改动量估算

| 模块 | 新增/修改 | 估算行数 |
|------|----------|---------|
| `analytics/` | 新增 | ~400 行（SQLite store，含 sessions/logs/attractions/routes 表操作 + 聚合查询） |
| `orchestrator/` | 修改 | ~110 行（读写路径切换到 SQLite + 新增 tool_result WebSocket 事件推送） |
| `character/` | 修改 | ~10 行（新增 3 个字段） |
| `api/` 新 handlers | 新增 | ~450 行（景点/路线/角色路线/分析/报告 handler） |
| `api/router.go` | 修改 | ~35 行（路由注册） |
| `orchestrator/` 新增 tool_result 事件 | 新增 | ~30 行（PersonaAgent 工具调用结果通过 wsHub 推送给前端，用于卡片渲染） |
| **合计** | | **~1005 行** |

### 4.2 Python 推理服务扩展

#### 4.2.1 新增 PersonaAgent 工具定义

**`inference/plugins/voice_llm/persona_agent.py`**，在 `PERSONA_TOOL_DEFINITIONS` 中新增：

```python
ToolDefinition(
    name="get_routes",
    description="获取当前景区的游览路线推荐。返回结构化路线数据。只返回当前角色被授权推荐的路线。",
    parameters={
        "type": "object",
        "properties": {
            "tags": {
                "type": "array",
                "items": {"type": "string"},
                "description": "按标签过滤路线，如 ['亲子', '文化']。为空则返回该角色的所有推荐路线。",
            },
        },
        "required": [],
    },
)

ToolDefinition(
    name="recognize_scenic_spot",
    description="当用户发送照片或图片时使用。识别照片中的景点并返回介绍。",
    parameters={
        "type": "object",
        "properties": {
            "image_base64": {
                "type": "string",
                "description": "图片的 base64 编码",
            },
        },
        "required": ["image_base64"],
    },
)
```

#### 4.2.2 工具处理器

**`get_routes` 处理器**：

```
输入: tags=["亲子"]（可选）
  ↓
1. 从会话上下文获取 character_id（orchestrator 在会话开始时已绑定，纯数据读取）
2. HTTP 调用 Go 服务：GET /api/v1/character-routes?character_id={id}&tags=亲子
3. Go 服务内部完成：查 character.RecommendedRoutes → 查 routes 表 → tags 过滤 → 返回 JSON
4. Python 直接把 JSON 结果返回给 PersonaAgent
```

Python 端不需要知道 RecommendedRoutes 的存在，不需要做任何过滤逻辑。

**`recognize_scenic_spot` 处理器**：

```
输入: image_base64（游客拍照）
  ↓
1. 调用 Qwen-VL API（多模态模型，满足赛题"至少1个多模态大模型"要求）
2. Prompt: "这张照片中的景点是什么？请介绍该景点的历史、特色和游览建议。"
3. 同时 HTTP 调用 Go 服务 GET /api/v1/attractions 尝试匹配已知景点
4. 返回景点介绍文本 + 结构化信息（如命中）
```

#### 4.2.3 PersonaAgent 系统指令更新

在 `PERSONA_AGENT_INSTRUCTIONS` 中增加景区导游场景行为约束。

#### 4.2.4 Python 端改动量估算

| 模块 | 新增/修改 | 估算行数 |
|------|----------|---------|
| `persona_agent.py` 工具定义 | 修改 | ~40 行 |
| 工具处理器 | 新增/修改 | ~100 行 |
| 系统指令 | 修改 | ~20 行 |
| **合计** | | **~160 行** |

### 4.3 get_routes 完整数据流

```
游客语音："推荐一条适合带小孩的路线"
  ↓
[Qwen-Omni] 语音转文本（已有）
  ↓
PersonaAgent（LLM）判断：调用 get_routes 工具
  ↓
Python 工具处理器：
  ① 从会话上下文读 character_id（纯数据，不调 LLM）
  ② HTTP → Go：GET /api/v1/character-routes?character_id=guide_xiaojing&tags=亲子
  ③ Go 内部：查 RecommendedRoutes → 查 routes → tags 过滤 → 返回 JSON
  ④ 返回 JSON 给 PersonaAgent
  ↓
PersonaAgent（LLM）收到结构化数据，生成自然语言回答：
  "我推荐您试试亲子休闲线，全程约1.5小时，途经莲花湖和儿童乐园，路线平坦适合推婴儿车。"
  ↓
[Qwen-TTS] 文本转语音（已有）
  ↓
[WebRTC] 音频推送给游客（已有）
```

整条链路中**没有额外 LLM 调用**。所有 LLM 调用都是 PersonaAgent 和已有语音链路的正常运行。

---

---

## 第 5 节：非功能约束

### 5.1 赛题硬性指标

| 指标 | 要求 | 实现保障 |
|------|------|---------|
| 多模态大模型 | 至少使用 1 个 | Qwen-VL（拍照识景 F5）+ Qwen-Omni（语音+视觉输入） |
| 问答准确率 | ≥ 90% | FAQ 问答对 + RAG 检索 + prompt 约束 + 测试用例验证 |
| 语音问答延迟 | ≤ 5 秒 | Qwen-Omni 端到端实时流式，CyberVerse 已验证延迟 < 3 秒 |
| 口型同步自然度 | 专家主观评估 | Avatar 关闭时为纯语音模式；有 GPU 时启用 FlashHead/LiveAct |

### 5.2 延迟预算分解

```
语音问答全链路延迟目标：≤ 5 秒

Qwen-Omni 端到端（语音→语音）：~2-3 秒（CyberVerse 已有数据）
  其中：
    ASR：~0.3 秒（Qwen-Omni 内置，不单独计时）
    LLM 首 token：~0.5-1 秒
    TTS 首包：~0.3 秒
    网络往返：~0.5 秒

PersonaAgent 工具调用（如 get_routes）：+0.5-1 秒
  其中：
    Go 端数据库查询：~5 毫秒
    HTTP 往返：~50 毫秒
    PersonaAgent 组织回答：~0.5 秒（LLM 推理）

总计：~2.5-4 秒，满足 ≤ 5 秒要求
```

### 5.3 并发与容量

| 参数 | 值 | 来源 |
|------|-----|------|
| 最大并发会话 | 4 | `session.max_concurrent` 配置 |
| 单会话最长时间 | 3600 秒 | `session.max_duration_s` 配置 |
| 空闲超时 | 300 秒 | `session.idle_timeout_s` 配置 |
| 景点数据量级 | 几十到上百条 | 单景区场景 |
| 路线数据量级 | 几条到十几条 | 单景区场景 |
| 对话日志量级 | 每天几百到几千条 | 比赛 demo 场景 |

SQLite 对这个量级的数据毫无压力，聚合查询在毫秒级完成。

### 5.4 部署模式

比赛 demo 采用单机部署（和 CyberVerse 现有部署方式一致）：

```
一台 Linux 服务器（或 WSL2）
├── Go API 服务 :8080（REST + WebSocket + TURN :8443）
├── Python 推理服务 :50051（gRPC）
├── SQLite 文件：data/cyberverse.db + data/tasks/tasks.db
├── 知识库文件：data/characters/<id>/knowledge/
└── 前端：小程序（游客端） + Vue 管理端（:5173 dev / Nginx 生产）
```

### 5.5 RAG 准确率保障措施

赛题要求问答准确率 ≥ 90%，需要系统性保障：

| 措施 | 说明 | 实施阶段 |
|------|------|---------|
| FAQ 问答对 | 预置 50-100 条高频问答，直接灌入知识库 | 数据准备阶段 |
| prompt 约束 | PersonaAgent 系统指令约束"只基于检索结果回答，无相关信息则说不知道" | 代码阶段 |
| chunk 参数调优 | chunk_size=900, overlap=120（现有配置） | 配置阶段 |
| 测试用例验证 | 准备 50-100 个测试问答对，跑准确率测试 | 验证阶段 |
| 混合检索（可选） | BM25 关键词检索 + 向量检索结果合并 | 如果纯向量不够再加 |

### 5.6 初赛提交物对应

| 提交物 | 内容 | 来源 |
|--------|------|------|
| 源码 | GitHub 仓库 | CyberVerse fork + 二次开发代码 |
| 部署手册 | 环境要求、安装步骤、配置说明 | 现有 CyberVerse 部署手册 + 景区配置补充 |
| 设计文档 | 需求分析、架构、模块设计、数据库、API | 本 Spec 落盘文件 |
| 方案 PPT | 项目介绍 + 技术方案 + 演示亮点 | 从 Spec 提炼 |
| 演示视频 | ≤ 7 分钟 | 录屏：语音对话 + 知识库问答 + 拍照识景 + 管理后台 + 大屏 |

### 5.7 测试策略

| 测试类型 | 覆盖范围 | 工具 |
|---------|---------|------|
| 单元测试 | Go 端：analytics store、character 扩展、API handlers | `go test`（复用现有） |
| 单元测试 | Python 端：工具处理器 | `pytest`（复用现有） |
| 集成测试 | API 端到端：景点/路线/会话 CRUD | `pytest` + `requests` |
| 准确率测试 | 50-100 个问答对，验证 RAG 命中率 | 自定义测试脚本 |
| 延迟测试 | 语音问答全链路延迟 ≤ 5 秒 | 录屏计时 + 日志分析 |

---

*Spec 1 系统级业务 Spec 完成。*
