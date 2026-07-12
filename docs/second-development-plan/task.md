# 任务清单

> 剩余任务按执行顺序排期。每完成一个任务：① 先写 progress.md ② 再对照本文件更新状态。
>
> 优先级：P0 = 核心链路必做 | P1 = 重要但可降级 | P2 = 锦上添花

---

## Phase 1：Go 后端 — 数据层

### T-01 `Character` 结构体扩展

**优先级**：P0
**依赖**：无
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`server/internal/character/store.go`

---

### T-02 `cyberverse.db` 初始化与迁移

**优先级**：P0
**依赖**：无
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/scenic/db.go`
**改动文件**：`server/cmd/cyberverse-server/main.go`

> 注：实际创建了 6 张表（含 reports），比原计划多 1 张。

---

### T-03 景点 CRUD Store

**优先级**：P0
**依赖**：T-02
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/scenic/attraction.go`

---

### T-04 路线 CRUD Store

**优先级**：P0
**依赖**：T-02
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/scenic/route.go`

> 注：DeleteRoute 的 character 引用检查需在 handler 层（T-07）实现，因 characters 表在文件系统而非 cyberverse.db。

---

### T-05 会话 + 对话日志 Store

**优先级**：P0
**依赖**：T-02
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/scenic/session.go`

**sentiment 字段说明**：
- T-05 阶段：所有新建会话的 sentiment 为 `'neutral'`（DB 默认值 + CreateSession 显式写入）
- T-23 阶段：会话结束时 LLM 分析会覆盖为 `'positive'` / `'neutral'` / `'negative'`
- 因此 T-05 的筛选功能在 T-23 实现前即可正常工作（按 'neutral' 筛选）

---

### T-06 Dashboard 聚合查询

**优先级**：P1
**依赖**：T-05
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/scenic/dashboard.go`

> 注：`hot_questions` 按 `normalized_content`（规则预处理后文本）聚合，非原始 content。预处理逻辑在 `session.go` 的 `normalizeQuestion()` 中实现：去标点、去问词前缀、截断 100 字符。LLM 语义聚合作为 T-23 扩展延后。

---

## Phase 2：Go 后端 — API 层

### T-07 景点/路线 API Handler

**优先级**：P0
**依赖**：T-03, T-04
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/api/scenic_handler.go`

---

### T-08 Analytics + Report API Handler

**优先级**：P0
**依赖**：T-05, T-06
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`server/internal/api/analytics_handler.go`, `server/internal/scenic/report.go`

---

### T-09 路由注册 + 依赖注入

**优先级**：P0
**依赖**：T-07, T-08
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`server/internal/api/router.go`, `server/cmd/cyberverse-server/main.go`

> 注：NewRouter 签名变更，在 configPath 和 taskServices 之间新增 scenicDB 参数。

---

## Phase 3：Go 后端 — Orchestrator 改造

### T-10 Orchestrator 会话生命周期钩子

**优先级**：P0
**依赖**：T-05, T-09
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`server/internal/orchestrator/session.go`, `server/internal/orchestrator/orchestrator.go`, `server/cmd/cyberverse-server/main.go`

> 注：双写策略为"先写文件，再写 DB"。DB 写入失败不影响文件主链路。session 记录在首次 persist 时创建，结束时调 EndSession 更新时长/轮次。

---

### T-11 Orchestrator 对话日志双写

**优先级**：P0
**依赖**：T-10
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：同 T-10（对话日志双写逻辑合并到 `persistSessionConversation` + `syncSessionToDB` 中）

> 注：通过 `dbSyncedCount` 追踪已同步消息数，避免重复 INSERT。每次 persist 只写新增消息。

---

### T-12 LLM 上下文注入景点/路线数据

**优先级**：P0
**依赖**：T-03, T-04
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`server/internal/orchestrator/orchestrator.go`

**实现内容**：
- 新增 `buildScenicContext()` 方法：从 cyberverse.db 加载景点列表 + 推荐路线，格式化为结构化摘要
- 注入点 1 — 文本管线：`standardSystemPrompt` 和 `standardSystemPromptWithRAG` 中，在角色设定之后、RAG 检索结果之前注入
- 注入点 2 — 语音管线：`buildVoiceLLMSessionConfig` 中，追加到 `char.SystemPrompt` 之后
- Best-effort 设计：DB 为 nil 或查询失败时返回空字符串，不影响原有流程
- 数据每次会话消息处理时实时读取（非会话开始时缓存），管理端修改后立即生效

**最终 system prompt 结构**：
```
【全局输出规范】 → 【角色设定】 → 【景区运营数据】→ 【角色素材检索结果(RAG)】
```

**验收标准**：
1. 数字人在对话中能准确说出当前可用的景点列表和路线名称
2. 当无景点/路线数据时，上下文不包含景区数据（优雅降级）
3. 路线数据变更后，下次消息处理立即生效（不依赖重启）
4. scenicDB 为 nil 时完全不影响原有对话流程

---

## Phase 4：Python 推理层

### T-13 PersonaAgent 新增工具定义

**优先级**：P0
**依赖**：无（可与 Go 后端并行）
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`inference/plugins/voice_llm/persona_agent.py`

**实现内容**：
- 新增 3 个工具定义：
  - `get_attractions`：获取景点列表（参数：category 可选）
  - `get_routes`：获取路线列表（参数：无）
  - `get_attraction_detail`：获取景点详情（参数：attraction_id 必填）
- 工具 schema 添加到 `PERSONA_TOOL_DEFINITIONS` 列表
- `PERSONA_AGENT_INSTRUCTIONS` 中补充 3 条工具使用指引

**验收标准**：
1. Python 推理服务启动无报错
2. PersonaAgent 能识别新工具调用意图
3. 工具参数 schema 符合 OpenAI function calling 规范

---

### T-14 PersonaAgent 工具处理器实现

**优先级**：P0
**依赖**：T-13, T-09（需要 Go API 可用）
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`inference/plugins/voice_llm/persona_agent.py`

**实现内容**：
- 新增 `_go_api_base()` 静态方法：读取 `GO_API_BASE` 环境变量，默认 `http://localhost:8080`
- 新增 `_go_api_get()` 异步方法：通过 httpx 调用 Go 后端 HTTP API
- 新增 3 个工具处理器：`_get_attractions`、`_get_routes`、`_get_attraction_detail`
- 更新 `_execute_tool` 分发逻辑：在 supervisor fallback 前匹配 3 个新工具
- 错误处理：API 不可达时返回友好中文提示，不崩溃

**验收标准**：
1. 数字人在对话中被问到"有什么好玩的"时，能调用 `get_attractions` 并返回景点列表
2. 数字人被问到"推荐路线"时，能调用 `get_routes` 并返回路线信息
3. Go 后端不可用时，工具返回"暂时无法获取"的提示

---

## Phase 5：管理端前端（剩余）

### T-15 CharacterEditPage 扩展

**优先级**：P1
**依赖**：T-09（路线 API 可用）
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：
- `server/internal/character/store.go` — Character 结构体 +`ScenicSystemPrompt` 字段
- `server/internal/orchestrator/orchestrator.go` — `characterSystemPrompt` 和语音管线使用 `ScenicSystemPrompt`
- `frontend/src/types/index.ts` — Character 接口 +3 字段
- `frontend/src/i18n/messages.ts` — +13 中英文 i18n 键
- `frontend/src/pages/CharacterEditPage.vue` — +表单字段, +路线加载, +景区配置区模板

**实现内容**：
- `scenic_category` 下拉框（导游 / 讲解员 / 客服）
- `scenic_system_prompt` textarea：景区专属人设提示词，非空时覆盖原 `system_prompt`
- `recommended_routes` 勾选框列表：加载所有路线，显示名称+时长+难度
- orchestrator 中 `characterSystemPrompt()` 和语音管线均检查 `ScenicSystemPrompt`，非空则替换

**验收标准**：
1. 打开角色编辑页 → 可选择 scenic_category ✅
2. 推荐路线区域正确显示所有已有路线 ✅
3. 勾选路线 → 保存 → 重新打开页面 → 勾选状态保持 ✅
4. 不影响现有字段的编辑功能 ✅
5. scenic_system_prompt 非空时，对话使用景区人设而非原角色提示词 ✅

---

## Phase 6：游客端 Web 适配

> **方案变更**：原计划采用微信小程序实现游客端，现改为复用 CyberVerse 现有 Web 前端。
> 原因：SessionPage 已有完整的实时语音/视频对话能力（LiveKit + Direct WebRTC），小程序需重新适配音视频 SDK，
> 开发周期长且不确定性高。复用 Web 端可直接使用已调通的音视频链路，只需新建入口页 + 适配景区 UI。

### T-16 景区入口页 ScenicLandingPage

**优先级**：P0
**依赖**：T-09, T-15
**状态**：✅ 代码完成（服务器端验证待执行）

**新建文件**：`frontend/src/pages/ScenicLandingPage.vue`

**参考模板**：`frontend/src/pages/KanshanLandingPage.vue`（一键进入模式）

**实现内容**：
- 景区品牌展示区：景区名称、简介、背景图
- 导游角色卡片：展示当前导游数字人（头像、名称、类别、简介）
- 景点列表区：从 `GET /api/v1/attractions` 加载，卡片展示（名称 + 分类 + 简介）
- 路线列表区：从 `GET /api/v1/routes` 加载，卡片展示（名称 + 时长 + 难度 + 途经景点）
- "开始对话"按钮：调用 `createSession()` → 跳转 `/session/:id`（参照 KanshanLandingPage 的 `enterKanshanVoiceCall` 模式）
- 动态加载导游角色：通过 API 查询 `scenic_category=guide` 的角色，而非硬编码 ID
- 路由注册：`/scenic` 路径

**验收标准**：
1. 访问 `/scenic` → 看到景区品牌展示 + 景点/路线列表
2. 点击"开始对话" → 自动创建会话 → 跳转到 SessionPage
3. 景点/路线数据从 API 动态加载，无数据时展示空状态
4. 页面在移动端浏览器可正常访问

---

### T-17 SessionPage 景区 UI 适配

**优先级**：P0
**依赖**：T-16
**状态**：⬜ 待开始

**改动文件**：`frontend/src/pages/SessionPage.vue`

**实现内容**：
- 返回按钮目标：从景区入口页进入时，返回 `/scenic` 而非 `/characters`
- 通过 `launchState.returnPath` 或 route query 传递返回路径（现有机制，KanshanLandingPage 已使用）
- 景区信息条：SessionPage 视频区域叠加显示景区名称和当前导游角色名称

> 注：景点/路线浏览不通过侧栏实现，而是通过对话工具调用 + ChatPanel 内嵌卡片渲染（T-14 已实现 get_attractions/get_routes 工具）。

**验收标准**：
1. 从 `/scenic` 进入对话 → 点击返回 → 回到 `/scenic`
2. 对话页视频区域显示景区名称和导游角色名
3. 不影响现有从 `/characters` → `/launch` → `/session` 的入口流程

---

### T-18 ChatPanel 路线/景点卡片渲染

**优先级**：P1
**依赖**：T-14
**状态**：⬜ 待开始

**改动文件**：`frontend/src/components/ChatPanel.vue`

**实现内容**：
- 在 ChatPanel 的 assistant 消息渲染中，检测 LLM 回复中的路线/景点信息
- 匹配模式：路线名称 + 途经景点 + 时长/难度等结构化文字
- 将匹配到的信息渲染为内嵌卡片（路线卡片：名称+时长+途经景点；景点卡片：名称+分类）
- 点击展开/折叠：折叠态显示摘要信息（名称+分类/时长），展开态显示 LLM 回复中的完整详情
- 展开内容来自 LLM 消费 DB + RAG 后整理的完整文字（T-12 注入景区数据 + RAG 注入知识库），前端不做额外数据请求
- 通过 `isExpanded` 属性控制详情区域的显示/隐藏，数据始终在 msg.content 中，展开只是改变渲染密度
- 未匹配到的消息保持原有纯文本渲染
- 不修改后端（Python/Go），仅前端解析 LLM 文字输出

**验收标准**：
1. 数字人推荐路线时，消息中显示路线卡片（名称+时长+途经景点）
2. 数字人列举景点时，消息中显示景点卡片（名称+分类）
3. 点击卡片展开 → 显示完整详情（描述、开放时间、标签等）
4. 再次点击折叠 → 回到摘要视图
5. 普通闲聊消息不受影响，保持纯文本渲染
6. 卡片样式与 ScenicLandingPage 风格一致

---

## Phase 7：数据准备

### T-19 Demo 景点数据

**优先级**：P0
**依赖**：T-09
**状态**：⬜ 待开始

**新建文件**：`data/demo/attractions.json`

**实现内容**：
- 8-12 个虚构景区景点，覆盖 3 个分类（自然风光、历史古迹、亲子游乐）
- 每个景点：名称、简介（50-100字）、分类、位置描述、标签
- 提供 seed 脚本或 curl 命令批量导入

**验收标准**：
1. 导入后 `GET /api/v1/attractions` 返回完整数据
2. 3 个分类均有数据覆盖
3. 景点简介内容合理、无乱码

---

### T-20 Demo 路线数据

**优先级**：P0
**依赖**：T-19
**状态**：⬜ 待开始

**新建文件**：`data/demo/routes.json`

**实现内容**：
- 3-4 条路线，每条含 3-5 个途经景点
- 路线类型：亲子休闲线、文化深度线、自然探索线等
- 每个途经点含建议停留时间和重点讲解提示
- 提供 seed 脚本或 curl 命令批量导入

**验收标准**：
1. 导入后 `GET /api/v1/routes` 返回完整数据
2. 每条路线的 steps 按 order 排序，attraction_name 正确关联
3. 路线总时长与各步骤时长大致吻合

---

### T-21 知识库文档准备

**优先级**：P1
**依赖**：无
**状态**：⬜ 待开始

**新建目录**：`data/demo/knowledge/`

**实现内容**：
- 景区介绍文档（Markdown）：景区概况、开放时间、门票价格、交通指南
- 景点详细文档（每个景点一个 Markdown 文件）
- FAQ 文档：常见问答对（20-30 条）
- 通过现有 knowledge API 上传到角色知识库

**验收标准**：
1. 上传后 RAG 索引构建成功（sources.json 中 status=ready）
2. 数字人能回答"门票多少钱"、"怎么去"等基础问题
3. 数字人能回答特定景点的详细介绍

---

## Phase 8：UI 适配 + 情感分析

### T-22 游客端 UI 风格重构（泉城主题）

**优先级**：P0
**依赖**：T-16, T-18
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：
- `frontend/src/pages/SessionPage.vue` — 对话页整体色调
- `frontend/src/components/ChatPanel.vue` — 聊天面板色调
- `frontend/src/components/AppHeader.vue` — 导航栏色调（游客访问时）

**问题现状**：
- `ScenicLandingPage.vue` 已适配泉城风格（绿白配色 `#16a34a` / `#f0fdf4` / `#ffffff`）
- `SessionPage.vue` 仍使用深色主题（`background: #000`，`cv-*` 深灰 token）
- `ChatPanel.vue` 使用硬编码深色（`#1e1e1e`、`#2a2a2a`、`#333`），无 CSS 变量
- 游客从 `/scenic` 点击"开始对话"后，视觉风格从绿色清新突变为深色赛博朋克，体验割裂

**实现内容**：
- SessionPage：视频区域保留黑色（视频播放需要），侧栏和控制栏改为浅色/半透明泉城风格
- ChatPanel：背景改为浅色系，消息气泡配色适配绿色主题，使用 CSS 变量替代硬编码色值
- AppHeader：游客从 `/scenic` 进入时显示浅色导航栏（可通过 route 或 launchState 判断来源）
- 管理端页面（`/attractions`、`/routes`、`/dashboard` 等）保持现有深色风格不变

**验收标准**：
1. 从 `/scenic` 进入对话 → SessionPage 和 ChatPanel 为浅色泉城风格
2. 从 `/characters` → `/launch` → `/session` 进入 → 保持原有深色风格（双主题共存）
3. 对话功能不受影响：文字消息、语音对话、ScenicCard 卡片渲染均正常
4. 移动端浏览器下样式不错乱

---

### T-23 会话结束情感标注

**优先级**：P1
**依赖**：T-10, T-13
**状态**：✅ 代码完成（服务器端验证待执行）

**改动文件**：`server/internal/orchestrator/` 会话结束流程 + `inference/plugins/voice_llm/persona_agent.py`

**实现内容**：
- 会话结束时，收集该会话所有 `conversation_logs`
- 调用 LLM 判定整体情感极性（positive / neutral / negative）
- 将结果写入 `sessions.sentiment` 字段
- 作为 SubAgent 任务异步执行，不阻塞会话关闭

**验收标准**：
1. 会话结束后 `sessions` 表的 `sentiment` 字段被正确填充
2. 正面对话（夸赞、满意）→ positive
3. 负面对话（投诉、不满）→ negative
4. 中性对话（纯问路、查信息）→ neutral
5. 情感分析失败不影响会话正常关闭

---

## Phase 9：集成验证

### T-24 后端编译 + 启动验证

**优先级**：P0
**依赖**：T-01 ~ T-12
**状态**：⬜ 待开始

**验收标准**：
1. `make server` 编译成功（服务器端）
2. `make test-go` 全部通过
3. 服务器启动后 health check 通过
4. 所有新 API 端点可用（curl 逐个验证）
5. 现有功能不受影响：角色 CRUD、会话创建、WebSocket 连接

---

### T-25 前端构建 + 页面验证

**优先级**：P0
**依赖**：T-07 ~ T-09, T-15, T-16
**状态**：⬜ 待开始

**验收标准**：
1. `npm run build` 构建成功（无 TypeScript 错误）
2. 5 个新页面路由可正常访问
3. 景点管理页：创建 → 列表 → 编辑 → 删除 全流程
4. 路线管理页：创建（含途经点）→ 列表 → 编辑 → 删除 全流程
5. 数据大屏：图表正常渲染（有 demo 数据时）
6. 对话日志：列表 + 筛选 + 展开详情
7. 报告页：生成 → 列表 → 查看详情

---

### T-26 端到端业务流程验证

**优先级**：P0
**依赖**：T-24, T-25
**状态**：⬜ 待开始

**验收标准**：

**流程 1：管理端配置流程**
1. 创建景点（至少 5 个，覆盖 3 分类）
2. 创建路线（至少 2 条，关联景点）
3. 编辑角色，设置 scenic_category 和 recommended_routes
4. 上传知识库文档

**流程 2：游客对话流程**
1. 访问 `/scenic` → 景区入口页显示景点和路线
2. 点击"开始对话" → 自动创建会话 → 进入 SessionPage
3. 问"有什么好玩的"→ 数字人调用工具返回景点列表
4. 问"推荐一条路线"→ 数字人返回路线推荐（含途经景点）
5. 问"门票多少钱"→ 数字人从知识库检索回答
6. 结束对话 → sessions 表有记录 + conversation_logs 有逐轮日志

**流程 3：数据分析流程**
1. 多次对话后，打开数据大屏 → 图表有数据
2. 打开对话日志 → 可查看历史对话
3. 生成感受度报告 → 报告内容合理

**流程 4：回归验证**
1. 原有角色 CRUD 功能正常
2. 原有 WebSocket 对话功能正常
3. 原有 SubAgent 任务功能正常
4. 原有知识库管理功能正常

---

## 附加任务：管理端门户聚合

### T-27 管理端 AdminLayout 侧边栏 + 路由重构

**优先级**：P0
**依赖**：T-15, T-18
**状态**：✅ 代码完成

**新建文件**：`frontend/src/layouts/AdminLayout.vue`

**改动文件**：
- `frontend/src/router/index.ts` — 所有管理端页面改为 `/admin/` 子路由
- `frontend/src/components/AppHeader.vue` — 导航链接改为 `/admin/` 前缀
- `frontend/src/i18n/messages.ts` — +adminPortal/backToScenic/characters/settings 中英文键
- `frontend/src/components/CharacterCard.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/components/SetupBanner.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/pages/CharacterListPage.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/pages/CharacterEditPage.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/pages/LaunchConfigPage.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/pages/SettingsPage.vue` — 链接改为 `/admin/` 前缀
- `frontend/src/pages/SessionPage.vue` — 默认返回路径改为 `/admin/characters`

**实现内容**：
- AdminLayout：左侧可折叠侧边栏（7 个导航项）+ 顶部 header + router-view 内容区
- 侧边栏导航项：数据大屏、角色管理、景点管理、路线管理、对话日志、分析报告、系统设置
- 底部"返回游客端"链接（→ /scenic）
- 路由结构：`/admin` 作为父路由，所有管理页面为 children
- `/admin` 默认重定向到 `/admin/dashboard`

**验收标准**：
1. 访问 `/admin` → 显示 AdminLayout + 侧边栏 + 默认显示数据大屏
2. 侧边栏点击各导航项 → 内容区切换对应页面
3. 侧边栏折叠/展开功能正常
4. 从侧边栏"返回游客端"→ 跳转到 `/scenic`
5. 所有管理页面内部链接（角色编辑、创建等）正常工作
