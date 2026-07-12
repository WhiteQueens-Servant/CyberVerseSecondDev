# Writing Plans: 景区导览 AI 数字人实现计划

> 基于 Spec 1/2/3 的完整实现计划。一人开发，盲开模式（Go 端无法本地编译，服务器验证）。

---

## 第 0 节：二开风险评估

### 风险矩阵

| 风险等级 | 模块 | 风险描述 | 缓解策略 |
|---------|------|---------|---------|
| 🔴 高 | `orchestrator.go` 对话持久化切换 | 4259 行核心模块，5 个热路径改动点，改错导致整个对话链路崩溃 | 双写过渡期：先加 SQLite 写入，保留旧文件写入，验证后再移除 |
| 🟡 中 | LiveKit 小程序适配 | 官方无小程序 SDK，社区方案质量不确定 | 备选方案：WebSocket 音频流（放弃 LiveKit） |
| 🟡 中 | Go 端盲开 | Windows 无法编译（缺 gcc/opus），代码正确性无法本地验证 | `go vet` 静态检查 + 服务器上编译验证 |
| 🟢 低 | 新增 analytics/ 模块 | 纯新增，参照 agenttask/store.go 模式 | 无 |
| 🟢 低 | 景点/路线 CRUD API | 纯新增 handler | 参照现有 characters handler |
| 🟢 低 | Character 结构体扩展 | JSON 向后兼容，加 omitempty 字段 | 无 |
| 🟢 低 | PersonaAgent 新加工具 | 现有分发模式简单，新增 if 分支 | 无 |
| 🟢 低 | 管理端 Vue 页面 | 纯新增，参照现有页面模式 | 无 |
| 🟢 低 | 小程序（除 LiveKit） | 标准微信小程序开发 | 无 |

### 整体风险评级：中等偏低

**理由**：70%+ 工作量是"新增模块"而非"修改现有模块"，CyberVerse 插件架构扩展性好，核心改动点（orchestrator）虽复杂但改动模式统一（文件读写 → SQLite 读写）。

---

## 第 1 节：开发环境与盲开策略

### 1.1 盲开策略

```
本地开发（Windows）：
  ✅ Python 推理服务 — 可启动（已验证端口 50051）
  ✅ Vue 管理端 — 可启动（npm run dev）
  ✅ 微信小程序 — 可开发（微信开发者工具）
  ❌ Go 服务器 — 无法编译（缺 gcc/opus C 库）

服务器验证（Linux）：
  ✅ 全部三个服务 — Docker 或直接部署
  ✅ 端到端联调 — LiveKit + WebRTC + 全链路
```

### 1.2 Go 盲开工作流

```
1. 写代码（本地编辑器）
2. go vet ./... 静态检查（不需要链接 opus，纯语法/类型检查）
3. git push
4. 服务器上 go build + 运行 + 测试
5. 发现问题 → 本地修 → 重复
```

### 1.3 依赖安装清单

**Python（本地 conda cyberverse 环境）**：
```bash
pip install aiosqlite  # SubAgent 任务持久化（如未安装）
```

**前端（npm）**：
```bash
cd frontend
npm install echarts vue-echarts  # 数据大屏图表
```

**小程序（微信开发者工具）**：
- LiveKit 小程序 SDK 或备选方案
- 无需 npm，直接在开发者工具中配置

---

## 第 2 节：实现阶段划分

### 总览

```
Phase 1: Go 后端扩展（~1005 行）
  ├── 1A: Character 结构体扩展（~10 行）
  ├── 1B: analytics/ 模块新建（~400 行）
  ├── 1C: API handlers 新增（~450 行）
  ├── 1D: orchestrator 对话持久化切换（~110 行）
  └── 1E: tool_result WebSocket 事件（~35 行）

Phase 2: Python 推理扩展（~160 行）
  ├── 2A: PersonaAgent 工具定义（~40 行）
  ├── 2B: 工具处理器（~100 行）
  └── 2C: 系统指令更新（~20 行）

Phase 3: 管理端 Vue 前端（~1800 行）
  ├── 3A: 基础设施（路由 + API 层 + 导航）（~220 行）
  ├── 3B: 景点管理页（~300 行）
  ├── 3C: 路线管理页（~400 行）
  ├── 3D: 数据大屏页（~350 行）
  ├── 3E: 对话日志页（~250 行）
  └── 3F: 感受度报告页（~200 行）

Phase 4: 游客端小程序
  ├── 4A: 项目骨架 + 首页
  ├── 4B: 对话页（核心）
  ├── 4C: 景点/路线详情页
  └── 4D: LiveKit 音频集成

Phase 5: 数据准备 + 联调
  ├── 5A: 景区知识库灌入
  ├── 5B: FAQ 问答对准备
  └── 5C: 服务器端到端联调

Phase 6: 提交物准备
  ├── 6A: 部署手册
  ├── 6B: 设计文档
  ├── 6C: 方案 PPT
  └── 6D: 演示视频
```

### 阶段依赖关系

```
Phase 1 (Go) ─────────────────────┐
  ↓                                │
Phase 2 (Python) ──────────────────┤→ Phase 5 (联调) → Phase 6 (提交物)
  ↓                                │
Phase 3 (管理端 Vue) ──────────────┘
  ↓
Phase 4 (小程序) ← 可与 Phase 3 并行
```

Phase 1 和 Phase 2 可以并行（改动不同语言、不同模块）。Phase 3 依赖 Phase 1 的 API 存在。Phase 4 依赖 Phase 1 的 WebSocket 事件定义。Phase 5 依赖全部完成。

---

## 第 3 节：Phase 1 — Go 后端扩展

### 1A: Character 结构体扩展

**文件**：`server/internal/character/store.go`

**改动**：在 `Character` 结构体中新增 2 个字段（`omitempty` 保证向后兼容）：

```go
ScenicCategory    string   `json:"scenic_category,omitempty"`
RecommendedRoutes []string `json:"recommended_routes,omitempty"`
```

**工作量**：~10 行，5 分钟。

### 1B: analytics/ 模块新建

**新建文件**：

| 文件 | 职责 | 估算行数 |
|------|------|---------|
| `server/internal/analytics/store.go` | SQLite store：OpenStore + migrate + CRUD | ~300 行 |
| `server/internal/analytics/types.go` | 数据类型定义 | ~80 行 |
| `server/internal/analytics/store_test.go` | 单元测试 | ~100 行 |

**参照**：`server/internal/agenttask/store.go` 的模式。

**核心内容**：

`store.go`：
- `OpenStore(dbPath string) (*Store, error)` — 打开 cyberverse.db
- `migrate()` — CREATE TABLE IF NOT EXISTS（sessions, conversation_logs, attractions, routes, route_steps）
- `CreateSession()` / `UpdateSession()` / `GetSession()` / `ListSessions()`
- `AppendConversationLog()` / `GetConversationLogs()`
- CRUD for attractions / routes / route_steps
- 聚合查询：`GetDashboard()` / `GetSentimentTrend()`

`types.go`：
- `Session`, `ConversationLog`, `Attraction`, `Route`, `RouteStep` 结构体
- `DashboardData`, `SentimentTrend` 聚合结果类型
- 输入结构体：`CreateAttractionInput`, `CreateRouteInput` 等

**工作量**：~400 行，约 3-4 小时。

### 1C: API handlers 新增

**修改文件**：`server/internal/api/router.go`（~35 行新增路由注册）

**新建文件**：

| 文件 | 职责 | 估算行数 |
|------|------|---------|
| `server/internal/api/attractions.go` | 景点 CRUD handlers | ~120 行 |
| `server/internal/api/routes_mgmt.go` | 路线 CRUD handlers | ~130 行 |
| `server/internal/api/analytics.go` | 会话/大屏/情感 handlers | ~120 行 |
| `server/internal/api/reports.go` | 报告 handlers | ~80 行 |

**关键设计**：

- `Router` 结构体新增 `analyticsStore *analytics.Store` 字段
- 在 `NewRouter()` 中初始化 analytics store（同目录下的 cyberverse.db）
- `handleCharacterRoutes` — 内部组合 character.RecommendedRoutes + routes 表查询 + tags 过滤
- `handleDashboard` — 直接调用 analytics store 的聚合查询
- `handleGenerateReport` — 创建 SubAgent 任务（复用 agenttask service）

**工作量**：~485 行（handlers 450 + router 35），约 4 小时。

### 1D: orchestrator 对话持久化切换（核心改动）

**修改文件**：`server/internal/orchestrator/orchestrator.go`

**策略**：双写过渡期 — 同时写 SQLite 和旧文件，读取优先从 SQLite。

**5 个改动点**：

| # | 改动点 | 位置 | 改动内容 | 风险 |
|---|--------|------|---------|------|
| 1 | 会话开始 | `SetupSession()` ~line 1476 | INSERT INTO sessions | 低 |
| 2 | 每轮对话写入 | `persistSessionConversation()` ~line 4180 | INSERT INTO conversation_logs（同时保留旧的文件写入） | 中 |
| 3 | LLM 上下文加载 | `HydrateVoiceDialogContext()` ~line 1565 | SELECT FROM conversation_logs（同时保留旧的文件读取作为 fallback） | 中 |
| 4 | 会话结束 | `TeardownSession()` ~line 3797 | UPDATE sessions SET ended_at, duration_s, turn_count | 低 |
| 5 | 情感标注 | `TeardownSession()` 尾部 | 新 goroutine：读 user 消息 → 调 LLM → UPDATE sentiment | 中 |

**具体实现**：

改动点 1（会话开始）：
```go
// In SetupSession(), after session is created:
if o.analyticsStore != nil {
    o.analyticsStore.CreateSession(ctx, &analytics.Session{
        ID:          session.ID,
        CharacterID: session.CharacterID,
        StartedAt:   session.CreatedAt,
        InputMode:   string(session.Mode),
    })
}
```

改动点 2（每轮对话写入 — 双写）：
```go
// In persistSessionConversation(), add SQLite write BEFORE existing file write:
if o.analyticsStore != nil {
    for _, msg := range messages {
        o.analyticsStore.AppendConversationLog(ctx, &analytics.ConversationLog{
            SessionID:   sessionID,
            CharacterID: characterID,
            TurnSeq:     turnSeq,
            Role:        msg["role"].(string),
            Content:     msg["content"].(string),
            ContentType: "text",
            CreatedAt:   time.Now().UTC().Format(time.RFC3339),
        })
    }
}
// ... existing file write code unchanged ...
```

改动点 3（LLM 上下文加载 — SQLite 优先，文件 fallback）：
```go
// In HydrateVoiceDialogContext():
var messages []map[string]any
if o.analyticsStore != nil {
    logs, err := o.analyticsStore.GetConversationLogs(ctx, sessionID, doubaoDialogContextLoadLimit)
    if err == nil && len(logs) > 0 {
        messages = convertLogsToMessages(logs)
    }
}
if len(messages) == 0 {
    // Fallback to existing file-based loading
    messages, _, _, err = o.charStore.LoadRecentMessages(...)
}
```

改动点 4（会话结束）：
```go
// In TeardownSession():
if o.analyticsStore != nil {
    o.analyticsStore.UpdateSession(ctx, session.ID, &analytics.SessionUpdate{
        EndedAt:   time.Now().UTC().Format(time.RFC3339),
        DurationS: int(time.Since(session.CreatedAt).Seconds()),
        TurnCount: session.TurnSeq,
    })
}
```

改动点 5（情感标注 — 异步）：
```go
// In TeardownSession(), after session update:
if o.analyticsStore != nil {
    go func() {
        sentiment := o.analyzeSentiment(ctx, sessionID)
        if sentiment != "" {
            o.analyticsStore.UpdateSessionSentiment(ctx, sessionID, sentiment)
        }
    }()
}
```

**Orchestrator 结构体新增字段**：
```go
analyticsStore *analytics.Store  // nil = disabled (backward compatible)
```

**工作量**：~110 行，约 3-4 小时（含仔细审查）。

### 1E: tool_result WebSocket 事件

**修改文件**：`server/internal/orchestrator/orchestrator.go`

**改动**：在 PersonaAgent 工具调用结果返回时，通过 wsHub 推送 `tool_result` 事件。

**位置**：在处理 voice_llm pipeline 的工具调用结果处（约 line 3400 附近），新增：

```go
// After tool result is received from Python:
o.broadcastJSON(sessionID, map[string]any{
    "type": "tool_result",
    "data": map[string]any{
        "tool": toolName,
        "data": toolResultData,
    },
})
```

**工作量**：~35 行，约 1 小时。

### Phase 1 总工作量

| 子阶段 | 估算行数 | 估算时间 |
|--------|---------|---------|
| 1A: Character 扩展 | ~10 行 | 5 分钟 |
| 1B: analytics/ 模块 | ~400 行 | 3-4 小时 |
| 1C: API handlers | ~485 行 | 4 小时 |
| 1D: orchestrator 切换 | ~110 行 | 3-4 小时 |
| 1E: tool_result 事件 | ~35 行 | 1 小时 |
| **合计** | **~1040 行** | **~12 小时** |

---

## 第 4 节：Phase 2 — Python 推理扩展

### 2A: PersonaAgent 工具定义

**修改文件**：`inference/plugins/voice_llm/persona_agent.py`

**改动**：在 `PERSONA_TOOL_DEFINITIONS` 列表中新增 2 个工具定义。

```python
# After existing 4 tool definitions:

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

**工作量**：~40 行，30 分钟。

### 2B: 工具处理器

**修改文件**：`inference/plugins/voice_llm/persona_agent.py`

**改动 1**：在 `_execute_tool()` 方法中新增 2 个分发分支：

```python
async def _execute_tool(self, call: ToolCall, session_config: VoiceLLMSessionConfig) -> SupervisorToolResult:
    name = call.name.strip()
    if name == "retrieve_character_knowledge":
        return await self._retrieve_character_knowledge(call, session_config)
    if name == "get_routes":
        return await self._get_routes(call, session_config)
    if name == "recognize_scenic_spot":
        return await self._recognize_scenic_spot(call, session_config)
    if self.supervisor is None:
        raise RuntimeError("persona supervisor is not initialized")
    return await self.supervisor.handle_tool_call(call, session_config.session_id)
```

**改动 2**：新增 `_get_routes()` 方法：

```python
async def _get_routes(self, call: ToolCall, session_config: VoiceLLMSessionConfig) -> SupervisorToolResult:
    """Fetch recommended routes for the current character from Go server."""
    import aiohttp
    
    tags = call.arguments.get("tags", [])
    character_id = session_config.character_id
    
    params = {"character_id": character_id}
    if tags:
        params["tags"] = ",".join(tags)
    
    go_server_url = os.environ.get("GO_SERVER_URL", "http://localhost:8080")
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{go_server_url}/api/v1/character-routes", params=params) as resp:
            if resp.status == 200:
                data = await resp.json()
                return SupervisorToolResult(content=json.dumps(data, ensure_ascii=False))
            return SupervisorToolResult(content=f"Error: failed to fetch routes (status {resp.status})")
```

**改动 3**：新增 `_recognize_scenic_spot()` 方法：

```python
async def _recognize_scenic_spot(self, call: ToolCall, session_config: VoiceLLMSessionConfig) -> SupervisorToolResult:
    """Recognize a scenic spot from a photo using Qwen-VL API."""
    import aiohttp
    
    image_base64 = call.arguments.get("image_base64", "")
    if not image_base64:
        return SupervisorToolResult(content="Error: no image provided")
    
    # Call Qwen-VL API via DashScope
    dashscope_key = os.environ.get("DASHSCOPE_API_KEY", "")
    base_url = os.environ.get("DASHSCOPE_BASE_URL", "https://dashscope-intl.aliyuncs.com/compatible-mode/v1")
    
    async with aiohttp.ClientSession() as session:
        payload = {
            "model": "qwen-vl-max",
            "messages": [{
                "role": "user",
                "content": [
                    {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{image_base64}"}},
                    {"type": "text", "text": "这张照片中的景点是什么？请介绍该景点的历史、特色和游览建议。用中文回答。"}
                ]
            }]
        }
        headers = {"Authorization": f"Bearer {dashscope_key}", "Content-Type": "application/json"}
        async with session.post(f"{base_url}/chat/completions", json=payload, headers=headers) as resp:
            if resp.status == 200:
                result = await resp.json()
                description = result["choices"][0]["message"]["content"]
                
                # Try to match known attractions
                go_server_url = os.environ.get("GO_SERVER_URL", "http://localhost:8080")
                async with session.get(f"{go_server_url}/api/v1/attractions") as attr_resp:
                    attractions = await attr_resp.json() if attr_resp.status == 200 else []
                
                return SupervisorToolResult(content=json.dumps({
                    "description": description,
                    "attractions": attractions
                }, ensure_ascii=False))
            
            return SupervisorToolResult(content=f"Error: Qwen-VL API failed (status {resp.status})")
```

**工作量**：~100 行，约 1.5 小时。

### 2C: 系统指令更新

**修改文件**：`inference/plugins/voice_llm/persona_agent.py`

**改动**：在 `PERSONA_AGENT_INSTRUCTIONS` 中增加景区导游场景约束。

```python
PERSONA_AGENT_INSTRUCTIONS = """...existing instructions...

## 景区导览场景补充规则：
1. 你是景区 AI 导游，专注于景区相关问题的回答。
2. 当游客询问路线推荐时，使用 get_routes 工具获取推荐路线，基于游客的偏好（如亲子、文化、休闲）进行推荐。
3. 当游客发送照片时，使用 recognize_scenic_spot 工具识别景点并介绍。
4. 只基于知识库中的信息回答景区相关问题。如果知识库中没有相关信息，诚实告知游客"抱歉，我暂时没有这方面的信息"。
5. 不要编造景区信息（如票价、开放时间等），必须基于知识库中的准确数据回答。
6. 回答要简洁友好，适合游客快速获取信息。"""
```

**工作量**：~20 行，15 分钟。

### Phase 2 总工作量

| 子阶段 | 估算行数 | 估算时间 |
|--------|---------|---------|
| 2A: 工具定义 | ~40 行 | 30 分钟 |
| 2B: 工具处理器 | ~100 行 | 1.5 小时 |
| 2C: 系统指令 | ~20 行 | 15 分钟 |
| **合计** | **~160 行** | **~2 小时** |

---

## 第 5 节：Phase 3 — 管理端 Vue 前端

### 3A: 基础设施

**修改文件**：

| 文件 | 改动 | 估算行数 |
|------|------|---------|
| `frontend/src/router/index.ts` | 新增 5 条路由 | ~20 行 |
| `frontend/src/services/api.ts` | 新增 API 调用函数 | ~170 行 |
| `frontend/src/components/AppHeader.vue` | 新增 5 个导航菜单项 | ~30 行 |
| `frontend/src/pages/CharacterEditPage.vue` | 新增 scenic_category + recommended_routes | ~80 行 |
| `frontend/package.json` | 新增 echarts, vue-echarts 依赖 | ~2 行 |

**工作量**：~302 行，约 2 小时。

### 3B: 景点管理页

**新建文件**：`frontend/src/pages/AttractionManagePage.vue`

**功能**：
- 景点列表表格（名称、分类、位置、简介截断、标签芯片、操作按钮）
- 新增/编辑弹窗表单（名称、简介、分类下拉、位置、图片上传、标签输入）
- 删除确认弹窗
- 分类筛选

**参照**：现有 `CharacterListPage.vue` 的表格模式 + `CharacterEditPage.vue` 的弹窗表单模式。

**工作量**：~300 行，约 2.5 小时。

### 3C: 路线管理页

**新建文件**：`frontend/src/pages/RouteManagePage.vue`

**功能**：
- 路线列表表格（名称、时长、难度、途经景点、标签、操作按钮）
- 新增/编辑弹窗表单（名称、描述、时长、难度、标签）
- 途经景点编辑器（从 attractions 表选择，上下移动排序，设置停留时间/讲解提示）
- 删除确认弹窗

**工作量**：~400 行，约 3.5 小时。

### 3D: 数据大屏页

**新建文件**：`frontend/src/pages/DashboardPage.vue`

**功能**：
- 今日服务人次（大字体数字卡片）
- 本周服务趋势（ECharts 折线图）
- 热门问答 TOP 10（ECharts 横向柱状图）
- 情感分布（ECharts 饼图）
- 活跃时段分布（ECharts 柱状图）

**技术**：使用 `echarts` + `vue-echarts`，按需引入模块减少包体积。

**工作量**：~350 行，约 3 小时。

### 3E: 对话日志页

**新建文件**：`frontend/src/pages/SessionLogPage.vue`

**功能**：
- 会话列表表格（时间、角色、时长、轮次、情感标签）
- 筛选条件（角色下拉、日期范围、情感标签）
- 点击展开详情（逐轮对话记录，带时间戳）

**工作量**：~250 行，约 2 小时。

### 3F: 感受度报告页

**新建文件**：`frontend/src/pages/ReportPage.vue`

**功能**：
- "生成报告"按钮（触发 SubAgent 任务）
- 报告列表（历史报告，状态：生成中/已完成/失败）
- 报告详情面板（LLM 分析结果：游客关注点、情感趋势、知识盲区、改进建议）
- WebSocket task_event 监听任务进度

**工作量**：~200 行，约 1.5 小时。

### Phase 3 总工作量

| 子阶段 | 估算行数 | 估算时间 |
|--------|---------|---------|
| 3A: 基础设施 | ~302 行 | 2 小时 |
| 3B: 景点管理页 | ~300 行 | 2.5 小时 |
| 3C: 路线管理页 | ~400 行 | 3.5 小时 |
| 3D: 数据大屏页 | ~350 行 | 3 小时 |
| 3E: 对话日志页 | ~250 行 | 2 小时 |
| 3F: 感受度报告页 | ~200 行 | 1.5 小时 |
| **合计** | **~1802 行** | **~14.5 小时** |

---

## 第 6 节：Phase 4 — 游客端小程序

### 4A: 项目骨架 + 首页

**工作内容**：
- 创建小程序项目结构
- 首页：角色列表展示（调用 `/api/v1/characters`）
- 全局状态管理（Behavior 或 mobx-miniprogram）
- WebSocket 连接管理封装

**文件结构**：
```
miniprogram/
├── app.js / app.json / app.wxss
├── pages/
│   ├── index/          ← 首页（角色选择）
│   ├── conversation/   ← 对话页
│   ├── attraction/     ← 景点详情页
│   └── route/          ← 路线详情页
├── components/
│   ├── chat-bubble/    ← 对话气泡
│   ├── route-card/     ← 路线卡片
│   ├── attraction-card/← 景点卡片
│   └── wave-animation/ ← 语音波形动画
├── services/
│   ├── api.js          ← HTTP 请求封装
│   ├── websocket.js    ← WebSocket 管理
│   └── livekit.js      ← LiveKit 集成
└── utils/
    └── state.js        ← 全局状态
```

**工作量**：约 4 小时。

### 4B: 对话页（核心）

**工作内容**：
- 对话页状态机（disconnected → connecting → connected → listening ↔ speaking）
- 麦克风录音（`wx.getRecorderManager()`，PCM 16kHz）
- 文本输入框 + 发送
- 对话气泡渲染（用户消息 + 数字人消息）
- 拍照识景（`wx.chooseMedia()` → base64 → WebSocket 发送）
- 工具调用卡片渲染（`tool_result` 事件 → 路线卡片 / 景点卡片）
- 打断机制（speaking 状态下按麦克风 → interrupt 事件）

**工作量**：约 8 小时。

### 4C: 景点/路线详情页

**工作内容**：
- 景点详情页（名称、简介全文、分类、标签、关联路线）
- 路线详情页（名称、描述、时长、难度、途经景点有序列表）
- 途经景点可点击跳转景点详情页
- 页面栈管理（`onShow`/`onHide` 保持 WebSocket 连接）

**工作量**：约 3 小时。

### 4D: LiveKit 音频集成

**工作内容**：
- 评估 LiveKit 小程序 SDK 可行性
- 如可行：集成 LiveKit 小程序组件，连接 Room，接收/发送音频
- 如不可行：降级方案 — WebSocket 音频流（Go 端需新增音频转发 handler）

**风险**：这是 Phase 4 中唯一的中等风险点。

**工作量**：约 4-8 小时（取决于方案）。

### Phase 4 总工作量

| 子阶段 | 估算时间 |
|--------|---------|
| 4A: 骨架 + 首页 | 4 小时 |
| 4B: 对话页 | 8 小时 |
| 4C: 详情页 | 3 小时 |
| 4D: LiveKit 集成 | 4-8 小时 |
| **合计** | **~19-23 小时** |

---

## 第 7 节：Phase 5 — 数据准备 + 联调

### 5A: 景区知识库灌入

**工作内容**：
- 准备景区介绍文档（TXT/PDF/FAQ）
- 通过管理端知识库管理页面上传
- 验证 RAG 检索效果

**工作量**：约 2 小时。

### 5B: FAQ 问答对准备

**工作内容**：
- 准备 50-100 条高频问答对
- 格式：`Q: 门票多少钱？\nA: 成人票80元，学生票40元，1.2米以下儿童免费。`
- 灌入知识库，验证准确率

**工作量**：约 3 小时。

### 5C: 服务器端到端联调

**工作内容**：
- 部署到 Linux 服务器（Docker 或直接运行）
- 验证 Go 服务编译 + 启动
- 验证 Python 推理服务启动
- 验证管理端 Vue 前端
- 验证小程序 → WebSocket → Go → Python → LiveKit 全链路
- 验证景点/路线 CRUD
- 验证数据大屏数据
- 验证感受度报告生成

**工作量**：约 4-6 小时。

### Phase 5 总工作量

| 子阶段 | 估算时间 |
|--------|---------|
| 5A: 知识库灌入 | 2 小时 |
| 5B: FAQ 准备 | 3 小时 |
| 5C: 联调 | 4-6 小时 |
| **合计** | **~9-11 小时** |

---

## 第 8 节：Phase 6 — 提交物准备

### 6A: 部署手册

**内容**：环境要求、安装步骤、配置说明、启动命令。

**工作量**：约 2 小时。

### 6B: 设计文档

**内容**：需求分析、架构设计、模块设计、数据库设计、API 设计。

**来源**：从 Spec 1/2/3 提炼整合。

**工作量**：约 3 小时。

### 6C: 方案 PPT

**内容**：项目介绍 + 技术方案 + 演示亮点。

**工作量**：约 2 小时。

### 6D: 演示视频

**内容**：≤ 7 分钟录屏，覆盖：
- 语音对话 + 知识库问答
- 拍照识景
- 路线推荐 + 卡片下钻
- 管理端景点/路线管理
- 数据大屏 + 对话日志 + 感受度报告

**工作量**：约 2 小时（含录制 + 剪辑）。

### Phase 6 总工作量

| 子阶段 | 估算时间 |
|--------|---------|
| 6A: 部署手册 | 2 小时 |
| 6B: 设计文档 | 3 小时 |
| 6C: 方案 PPT | 2 小时 |
| 6D: 演示视频 | 2 小时 |
| **合计** | **~9 小时** |

---

## 第 9 节：总工作量估算与时间线

### 工作量汇总

| Phase | 内容 | 代码行数 | 估算时间 |
|-------|------|---------|---------|
| Phase 1 | Go 后端扩展 | ~1040 行 | ~12 小时 |
| Phase 2 | Python 推理扩展 | ~160 行 | ~2 小时 |
| Phase 3 | 管理端 Vue 前端 | ~1802 行 | ~14.5 小时 |
| Phase 4 | 游客端小程序 | — | ~19-23 小时 |
| Phase 5 | 数据准备 + 联调 | — | ~9-11 小时 |
| Phase 6 | 提交物准备 | — | ~9 小时 |
| **合计** | | **~3002 行** | **~65-72 小时** |

### 建议时间线（4-5 周）

```
Week 1: Phase 1（Go 后端）+ Phase 2（Python 推理）
  ├── Day 1-2: 1A + 1B（analytics 模块）
  ├── Day 3-4: 1C（API handlers）
  ├── Day 5: 1D + 1E（orchestrator + tool_result）
  └── Day 5: Phase 2（Python，与 Go 并行）

Week 2: Phase 3（管理端 Vue）
  ├── Day 1: 3A（基础设施）
  ├── Day 2: 3B + 3C（景点/路线管理）
  ├── Day 3: 3D（数据大屏）
  └── Day 4: 3E + 3F（对话日志 + 报告）

Week 3: Phase 4（小程序）
  ├── Day 1: 4A（骨架 + 首页）
  ├── Day 2-3: 4B（对话页）
  ├── Day 4: 4C（详情页）
  └── Day 5: 4D（LiveKit 集成）

Week 4: Phase 5（联调）+ Phase 6（提交物）
  ├── Day 1-2: 5A + 5B（数据准备）
  ├── Day 3: 5C（联调）
  └── Day 4-5: Phase 6（提交物）

Week 5（缓冲）: 问题修复 + 演示准备
```

### 关键里程碑

| 里程碑 | 完成标志 | 预计时间 |
|--------|---------|---------|
| M1: 后端 API 可用 | Go 服务器启动，景点/路线/分析 API 正常响应 | Week 1 末 |
| M2: 管理端可用 | Vue 前端 5 个新页面全部可用 | Week 2 末 |
| M3: 游客端可用 | 小程序 4 个页面全部可用，语音对话正常 | Week 3 末 |
| M4: 端到端联调通过 | 全链路（小程序 → Go → Python → LiveKit）跑通 | Week 4 Day 3 |
| M5: 提交物就绪 | 源码 + 文档 + PPT + 视频全部完成 | Week 4 末 |

---

## 第 10 节：盲开注意事项

### 10.1 Go 代码盲开检查清单

- [ ] 所有新增 `.go` 文件通过 `go vet ./...`（语法/类型检查）
- [ ] 导入路径正确（`github.com/cyberverse/server/internal/analytics` 等）
- [ ] SQLite driver 导入正确（参照 `agenttask/store.go` 的 `_ "github.com/mattn/go-sqlite3"`）
- [ ] 接口兼容（`Orchestrator` 新增 `analyticsStore` 字段不影响现有构造函数）
- [ ] 错误处理完整（所有 SQLite 操作的 error 都要检查）
- [ ] 并发安全（`analytics.Store` 的 `db.SetMaxOpenConns(1)` 保证 SQLite 单写）

### 10.2 服务器验证优先级

```
高优先级（必须先验证）：
  1. Go 服务器编译通过
  2. cyberverse.db 表创建成功
  3. 景点/路线 CRUD API 正常
  4. orchestrator 对话持久化（双写）正常

中优先级：
  5. 数据大屏聚合查询正常
  6. 感感度报告生成正常
  7. tool_result WebSocket 事件推送正常

低优先级（可最后验证）：
  8. 小程序 LiveKit 音频
  9. 管理端 UI 细节
```

### 10.3 回滚方案

如果 orchestrator 改动导致对话链路异常：
1. 禁用 `analyticsStore`（设为 nil），系统回退到纯文件模式
2. 所有新增功能（景点/路线/分析）仍可用（它们不依赖 orchestrator 改动）
3. 只有对话日志和数据大屏会受影响

---

*Writing Plans 完成。待用户确认后进入实现阶段。*
