# 进度记录

## 工作规则

> **每次完成任务**：① 先写 progress.md 记录做了什么 ② 再对照 task.md 确认验收标准是否达成 ③ 及时更新 task.md 状态标记。
>
> **task.md 状态标记**：⬜ 待开始 → 🔄 进行中 → ✅ 已完成 → ❌ 已跳过

---

## 2026-06-17：管理端前端（第一批）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| API 层扩展 | `frontend/src/services/api.ts` | ✅ 完成 |
| 路由扩展 | `frontend/src/router/index.ts` | ✅ 完成 |
| 导航菜单 + i18n | `frontend/src/components/AppHeader.vue` + `frontend/src/i18n/messages.ts` | ✅ 完成 |
| 景点管理页 | `frontend/src/pages/AttractionManagePage.vue` | ✅ 完成 |
| 路线管理页 | `frontend/src/pages/RouteManagePage.vue` | ✅ 完成 |
| 数据大屏页 | `frontend/src/pages/DashboardPage.vue` | ✅ 完成 |
| 对话日志页 | `frontend/src/pages/SessionLogPage.vue` | ✅ 完成 |
| 感受度报告页 | `frontend/src/pages/ReportPage.vue` | ✅ 完成 |
| echarts 依赖安装 | `frontend/package.json` | ✅ 完成 |

### 验证结果

- TypeScript 类型检查：零错误 ✅
- Vite 生产构建：成功 ✅
- 构建产物：8 个 chunk，DashboardPage 560KB（echarts 体积，管理端可接受）

### 新增文件清单

```
frontend/src/pages/AttractionManagePage.vue   (新建)
frontend/src/pages/RouteManagePage.vue        (新建)
frontend/src/pages/DashboardPage.vue          (新建)
frontend/src/pages/SessionLogPage.vue         (新建)
frontend/src/pages/ReportPage.vue             (新建)
```

### 修改文件清单

```
frontend/src/services/api.ts          (+17 API 函数, +12 接口)
frontend/src/router/index.ts          (+5 路由)
frontend/src/components/AppHeader.vue (+导航链接)
frontend/src/i18n/messages.ts         (+中英文 nav 文案)
frontend/package.json                 (+echarts, +vue-echarts)
```

### 备注

- 管理端前端 5 个新页面全部完成，但目前调用的后端 API 尚未实现（Go handler + analytics 模块）
- CharacterEditPage 的 `scenic_category` + `recommended_fields` 改造尚未进行（依赖路线 API 先可用）
- 报告生成的 WebSocket task_event 实时监听尚未接入（依赖 Go 后端 task 广播）

---

## 2026-06-17：Go 后端 — Phase 1 数据层 ✅ 已完成

### 已完成

| Task | 文件 | 变更 |
|------|------|------|
| T-01 Character 结构体扩展 | `server/internal/character/store.go` | +2 字段（ScenicCategory, RecommendedRoutes） |
| T-02 cyberverse.db 初始化 | `server/internal/scenic/db.go` (新建), `server/cmd/cyberverse-server/main.go` | 6 张表迁移 + main.go 初始化 |
| T-03 景点 CRUD Store | `server/internal/scenic/attraction.go` (新建) | List/Get/Create/Update/Delete + 删除约束 |
| T-04 路线 CRUD Store | `server/internal/scenic/route.go` (新建) | List/Get/Create/Update/Delete + 事务 + steps JOIN |
| T-05 会话 + 对话日志 Store | `server/internal/scenic/session.go` (新建) | CreateSession/EndSession/AppendLog/ListSessions/GetSessionDetail |
| T-06 Dashboard 聚合查询 | `server/internal/scenic/dashboard.go` (新建) | GetDashboard 返回 5 个聚合指标 |

### 新增文件

```
server/internal/scenic/db.go           — DB 初始化 + 迁移（6 张表）
server/internal/scenic/attraction.go   — 景点 CRUD Store（164 行）
server/internal/scenic/route.go        — 路线 CRUD Store + 事务（262 行）
server/internal/scenic/session.go      — 会话 + 对话日志（192 行）
server/internal/scenic/dashboard.go    — Dashboard 聚合查询
```

### 修改文件

```
server/internal/character/store.go     — Character 结构体 +2 字段
server/cmd/cyberverse-server/main.go   — scenic DB 初始化 + 优雅关闭
```

### 设计决策记录

- `sessions.sentiment` 默认值 `'neutral'`（非空字符串），确保筛选功能在 T-24 实现前即可工作
- `DeleteRoute` 的 character 引用检查从 Store 层移至 handler 层（T-07），因 characters 表在文件系统
- `SessionLogDetail` 为 Session + Messages 的组合结构，减少前端请求次数

### 待服务器端验证

- [ ] `go vet ./server/internal/scenic/...` 无错误
- [ ] `go build -tags livekit ./cmd/cyberverse-server/` 编译成功
- [ ] 启动服务器后 `data/cyberverse.db` 自动创建 + 6 张表存在
- [ ] 景点/路线 CRUD 全流程
- [ ] 会话创建 → 对话 → 结束 → 查询 全流程
- [ ] Dashboard 返回 5 个聚合指标
- [ ] 现有功能不受影响（角色 CRUD、WebSocket 对话、SubAgent 任务）

> 详细验收命令见 `execution_log.md`

---

## 2026-06-17：Go 后端 — Phase 2 API 层 ✅ 已完成

### 已完成

| Task | 文件 | 变更 |
|------|------|------|
| T-07 景点/路线 Handler | `server/internal/api/scenic_handler.go` (新建) | 10 个端点（景点 5 + 路线 5） |
| T-08 Analytics/Report Handler | `server/internal/api/analytics_handler.go` (新建), `server/internal/scenic/report.go` (新建) | 7 个端点（dashboard + sessions 3 + reports 3） |
| T-09 路由注册 + 依赖注入 | `server/internal/api/router.go`, `server/cmd/cyberverse-server/main.go` | Router +scenicDB, +17 路由 |

### 新增文件

```
server/internal/api/scenic_handler.go      — 景点/路线 CRUD handler
server/internal/api/analytics_handler.go   — Dashboard/Sessions/Reports handler
server/internal/scenic/report.go           — Report Store（Create/List/Get/Delete/UpdateContent/UpdateStatus）
```

### 修改文件

```
server/internal/api/router.go              — +scenicDB 字段, +17 路由注册, NewRouter 签名变更
server/cmd/cyberverse-server/main.go       — 传入 scenicDB 到 NewRouter
```

### API 端点总览

```
景点：GET/POST /api/v1/attractions, GET/PUT/DELETE /api/v1/attractions/{id}
路线：GET/POST /api/v1/routes, GET/PUT/DELETE /api/v1/routes/{id}
分析：GET /api/v1/analytics/dashboard, GET /api/v1/analytics/sessions, GET /api/v1/analytics/sessions/{id}
报告：GET /api/v1/reports, POST /api/v1/reports/generate, GET/DELETE /api/v1/reports/{id}
```

### 设计决策

- `handleDeleteRoute` 的 character 引用检查在 handler 层实现（遍历 file-based character store）
- Report 的 `handleGenerateReport` 目前仅创建 queued 状态的记录，LLM 分析集成留待 T-24

### 待服务器端验证

- [ ] `go build -tags livekit ./cmd/cyberverse-server/` 编译成功
- [ ] 17 个新端点全部可用（curl 验证）
- [ ] 现有端点不受影响
- [ ] CORS 对新端点生效

> 详细验收命令见 `execution_log.md`

---

## 2026-06-17：Go 后端 — Phase 3 Orchestrator 改造 ✅ 已完成

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-10 会话生命周期 DB 钩子 | `orchestrator/orchestrator.go`, `orchestrator/session.go`, `main.go` | ✅ 完成 |
| T-11 对话日志双写 | `orchestrator/orchestrator.go` (`syncSessionToDB`) | ✅ 完成 |
| T-12 LLM 上下文注入 | `orchestrator/orchestrator.go` (`buildScenicContext`) | ✅ 完成 |

### 改动文件总览

```
server/internal/orchestrator/session.go       — +dbSyncedCount, +dbSessionID, +DBSessionID()
server/internal/orchestrator/orchestrator.go   — +scenicDB 字段, +SetScenicDB(), +syncSessionToDB(), +buildScenicContext()
server/cmd/cyberverse-server/main.go           — orch.SetScenicDB(scenicDB), OnSessionEnd→EndSession
```

### 核心机制

**T-10/T-11 — 双写架构**：
```
对话消息到达 → persistSessionConversation()
  ├─ 1. 写文件（原有逻辑，不动）
  └─ 2. syncSessionToDB()（best-effort）
       ├─ 首次：CreateSession → 记录 dbSessionID
       └─ 每次：只写 dbSyncedCount 之后的新增消息
会话结束 → OnSessionEnd → EndSession(dbSessionID)
```

- 先写文件再写 DB，DB 失败不影响主链路
- `dbSessionID`（DB UUID）与 orchestrator session ID 通过 `DBSessionID()` 映射

**T-12 — LLM 上下文注入**：
```
【全局输出规范】→【角色设定】→【景区运营数据】→【角色素材检索结果(RAG)】
```

- `buildScenicContext()` 从 cyberverse.db 读取景点列表 + 路线，格式化为结构化摘要
- 文本管线：注入到角色设定之后、RAG 之前
- 语音管线：追加到 `char.SystemPrompt` 之后
- Best-effort：DB 为 nil 或无数据时返回空，不影响原有流程

### 待服务器端验证

- [ ] `go build -tags livekit ./cmd/cyberverse-server/` 编译成功
- [ ] 会话创建/结束时 DB 同步正常
- [ ] 对话消息双写到 conversation_logs
- [ ] LLM 回复中包含景点/路线信息
- [ ] scenicDB 为 nil 时不影响原有对话

> 详细验收命令见 `execution_log.md`

---

## 2026-06-17：Python 推理层 — Phase 4 ✅ 代码完成（服务器端验证待执行）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-13 PersonaAgent 工具定义 | `inference/plugins/voice_llm/persona_agent.py` | ✅ 完成 |
| T-14 PersonaAgent 工具处理器 | `inference/plugins/voice_llm/persona_agent.py` | ✅ 完成 |

### 改动文件总览

```
inference/plugins/voice_llm/persona_agent.py — +3 工具定义(get_attractions/get_routes/get_attraction_detail), +HTTP 客户端, +3 处理方法, 更新 _execute_tool
```

### 核心机制

**新增工具**：
- `get_attractions`：获取景点列表，可选 category 筛选
- `get_routes`：获取路线列表（含途经景点和时长）
- `get_attraction_detail`：获取单个景点详情

**HTTP 调用链**：
```
PersonaAgent 工具调用 → httpx GET → Go 后端 /api/v1/attractions|routes → 返回 JSON
```

- Go API 地址通过 `GO_API_BASE` 环境变量配置，默认 `http://localhost:8080`
- 错误处理：API 不可达时返回友好中文提示，不崩溃
- `_execute_tool` 中新工具匹配在 supervisor fallback 之前

### 待服务器端验证

- [ ] Python 推理服务启动无报错，工具数量为 8
- [ ] 语音对话中询问景点/路线时，数字人正确调用新工具
- [ ] Go 后端不可用时，工具返回友好提示

> 详细验收命令见 `execution_log.md`

---

## 2026-06-17：管理端前端 — Phase 5 ✅ 代码完成（服务器端验证待执行）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-15 CharacterEditPage 扩展 | `character/store.go`, `orchestrator/orchestrator.go`, `types/index.ts`, `i18n/messages.ts`, `CharacterEditPage.vue` | ✅ 完成 |

### 改动文件总览

```
server/internal/character/store.go         — +ScenicSystemPrompt 字段
server/internal/orchestrator/orchestrator.go — characterSystemPrompt + 语音管线使用 ScenicSystemPrompt
frontend/src/types/index.ts                — Character 接口 +3 字段
frontend/src/i18n/messages.ts              — +13 中英文 i18n 键
frontend/src/pages/CharacterEditPage.vue   — +表单字段, +路线加载, +景区配置区模板
```

### 核心机制

**新增字段**：
- `scenic_category`：导游(guide) / 讲解员(narrator) / 客服(service)
- `scenic_system_prompt`：景区专属人设提示词，非空时覆盖原 `system_prompt`
- `recommended_routes`：角色关联的路线 ID 列表

**scenic_system_prompt 生效链路**：
```
管理员在 CharacterEditPage 编辑 → PUT /api/v1/characters/{id}
→ character.json 保存 scenic_system_prompt
→ orchestrator 构建 prompt 时检查：非空则替换 system_prompt
→ 文本管线 + 语音管线均生效
```

### 待服务器端验证

- [ ] 角色编辑页可选择 scenic_category、填写 scenic_system_prompt、勾选推荐路线
- [ ] 保存后重新打开，数据保持
- [ ] scenic_system_prompt 非空时，对话使用景区人设
- [ ] scenic_system_prompt 为空时，回退到原 system_prompt

> 详细验收命令见 `execution_log.md`

---

## 2026-06-17：游客端 Web 适配 — Phase 6 ✅ 代码完成（服务器端验证待执行）

### 方案变更

原计划采用微信小程序实现游客端（T-16~T-20 共 5 个任务），现改为复用 CyberVerse 现有 Web 前端（T-16~T-18 共 3 个任务）。

**变更原因**：
- SessionPage 已有完整的实时语音/视频对话能力（LiveKit + Direct WebRTC）
- 小程序需重新适配 LiveKit SDK，开发周期长且不确定性高
- 复用 Web 端可直接使用已调通的音视频链路

**任务重编号**：原 T-21~T-27 → T-19~T-25（总任务数 27→25）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-16 景区入口页 | `frontend/src/pages/ScenicLandingPage.vue`（新建）, `frontend/src/router/index.ts`, `frontend/src/i18n/messages.ts` | ✅ 完成 |
| T-17 SessionPage 适配 | `frontend/src/pages/SessionPage.vue` | ⬜ 待开始 |
| T-18 ChatPanel 卡片渲染 | `frontend/src/components/ScenicCard.vue`（新建）, `frontend/src/components/ChatPanel.vue` | ✅ 完成 |

### T-16 变更详情

**新建文件**：`frontend/src/pages/ScenicLandingPage.vue`
**路由**：`/scenic` → ScenicLandingPage
**功能**：
- 景区品牌展示区（绿色主题，标题+描述）
- 导游角色卡片：动态加载 `scenic_category=guide` 的角色，头像优先级 `active_image > avatar_image > placeholder`
- 景点列表：从 `GET /api/v1/attractions` 加载，卡片展示（名称+分类+标签+图片占位）
- 路线列表：从 `GET /api/v1/routes` 加载，卡片展示（名称+时长+难度+途经景点）
- "开始对话"按钮：`createSession()` → `/session/:id`（参照 KanshanLandingPage 模式）
- 20 条中英文 i18n 键

**设计决策**：
- 景点卡片留出 `image_url` 图片占位，有值时展示图片，无值时灰色占位符
- 导游角色通过 API 动态查询（`scenic_category=guide`），不硬编码 ID

### T-18 变更详情

**新建文件**：`frontend/src/components/ScenicCard.vue`
**改动文件**：`frontend/src/components/ChatPanel.vue`

**功能**：
- ChatPanel 中 assistant 消息增加路线/景点信息检测
- 检测到时渲染 ScenicCard 组件（可展开/折叠）
- 折叠态：摘要信息（名称+时长/分类）
- 展开态：LLM 消费 DB+RAG 后的完整详情
- 未检测到的消息保持原有纯文本渲染

**设计决策**：
- 不改后端（Python/Go），仅前端解析 LLM 文字输出
- 数据全在 `msg.content` 中，展开只是渲染密度切换
- ScenicCard 样式与 ScenicLandingPage 风格一致（暗色主题）

### 待服务器端验证

- [ ] `npm run build` 构建成功
- [ ] `/scenic` 页面正常加载，景点/路线数据展示正确
- [ ] 点击"开始对话" → 创建会话 → 跳转 SessionPage
- [ ] 对话中数字人推荐路线时，ChatPanel 显示路线卡片
- [ ] 点击卡片展开/折叠正常
- [ ] 从 `/scenic` 进入对话 → 返回 → 回到 `/scenic`

> 详细验收命令见 `execution_log.md`

---

## 2026-06-18：游客端 UI 风格重构 — T-22 ✅ 代码完成（服务器端验证待执行）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-22a ChatPanel 浅色主题 | `frontend/src/components/ChatPanel.vue` | ✅ 完成 |
| T-22b SessionPage 浅色主题 | `frontend/src/pages/SessionPage.vue` | ✅ 完成 |
| T-22c AppHeader 双主题 | 不需要修改（管理端保持深色） | ✅ 无需改动 |

### 改动文件总览

```
frontend/src/components/ChatPanel.vue  — +theme prop, +CSS 变量双主题, .theme-light 覆盖
frontend/src/pages/SessionPage.vue     — +isScenicMode, +scenic-mode CSS class, ChatPanel theme 传递
frontend/src/router/index.ts           — '/' 重定向至 '/scenic'
```

### 核心机制

**双主题共存**：
- 游客从 `/scenic` 进入对话 → `launchState.returnPath === '/scenic'` → `isScenicMode = true`
- 管理端从 `/characters` → `/launch` → `/session` 进入 → `isScenicMode = false`
- 同一个 SessionPage + ChatPanel，通过 CSS class 切换主题

**ChatPanel 主题切换**：
- 新增 `theme` prop（`'light' | 'dark'`，默认 `'dark'`）
- 使用 `--chat-*` CSS 变量，`.theme-light` 覆盖为泉城绿色系
- 原有深色值完全保留（硬编码 → CSS 变量默认值）

**SessionPage 泉城风格**：
- 视频区域保持黑色（视频播放需要）
- 侧栏：`#f8faf8` 浅绿背景 + `#d4e8d4` 边框
- 控制栏：`rgba(255,255,255,0.88)` 半透明白色
- 返回按钮/FPS 按钮：白色毛玻璃风格
- 徽标按钮：绿色主色调 `#16a34a`

**颜色体系**：
- 主色：`#16a34a`（绿色，与 ScenicLandingPage 一致）
- 背景：`#f8faf8`（极浅绿）
- 边框：`#d4e8d4`（浅绿）
- 文字：`#1a1a1a`（深灰）

### 待服务器端验证

- [ ] `npm run build` 构建成功（无新增 TS 错误）
- [ ] 从 `/scenic` 进入对话 → SessionPage + ChatPanel 为浅色泉城风格
- [ ] 从 `/characters` → `/launch` → `/session` 进入 → 保持原有深色风格
- [ ] 对话功能不受影响（文字/语音/ScenicCard）
- [ ] 移动端浏览器样式正常

---

## 2026-06-18：会话结束情感标注 — T-23 ✅ 代码完成（服务器端验证待执行）

### 已完成任务

| 任务 | 文件 | 状态 |
|------|------|------|
| T-23 会话结束情感标注 | `scenic/session.go`, `orchestrator/orchestrator.go`, `main.go` | ✅ 完成 |

### 改动文件总览

```
server/internal/scenic/session.go              — +UpdateSentiment() 方法
server/internal/orchestrator/orchestrator.go    — +AnalyzeSessionSentiment() 方法, +normalizeSentiment()
server/cmd/cyberverse-server/main.go            — OnSessionEnd 追加 go orch.AnalyzeSessionSentiment()
```

### 核心机制

**触发链路**：
```
会话结束 → OnSessionEnd 回调
  ├─ PersistSessionConversation()    ← 现有，同步
  ├─ scenicDB.EndSession()           ← 现有，同步
  └─ go orch.AnalyzeSessionSentiment()  ← 新增，异步 goroutine
```

**情感分析流程**：
1. 从 history 中取最后 20 条消息，拼接为"角色: 内容"格式的对话文本
2. 构建 prompt："请分析以下对话的整体情感倾向，只回复一个单词：positive/neutral/negative"
3. 调用 `inference.GenerateLLMStream()`（gRPC → Python LLM）以 temperature=0.1 生成
4. 解析 LLM 回复，提取 positive/neutral/negative
5. 调用 `scenicDB.UpdateSentiment()` 写回 sessions 表

**设计决策**：
- **异步 goroutine**：不阻塞会话关闭，best-effort 失败仅 log
- **不用 SubAgent task**：orchestrator 没有 taskService 引用，且情感分析是单次 LLM 调用，不需要 LangGraph 工具循环
- **temperature=0.1**：低温度确保分类结果确定性
- **只取最后 20 条消息**：避免超长对话超出 token 限制
- **normalizeSentiment**：容错解析，LLM 回复中包含 "positive"/"negative" 即匹配，否则默认 neutral

### 待服务器端验证

- [ ] `go build -tags livekit` 编译成功
- [ ] 对话结束 → sessions 表 sentiment 字段被更新（非全 neutral）
- [ ] 正面对话 → sentiment = positive
- [ ] 负面对话 → sentiment = negative
- [ ] 中性对话 → sentiment = neutral
- [ ] Dashboard 情感分布饼图出现非 neutral 数据
- [ ] LLM 调用失败时不影响会话正常关闭（仅 log）
