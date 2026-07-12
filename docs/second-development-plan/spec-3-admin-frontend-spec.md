# Spec 3: 管理端 Spec（Vue 前端改造）

> 基于 Spec 1 系统级业务 Spec，在现有 CyberVerse Vue 前端上扩展管理端功能。

---

## 第 1 节：页面结构

### 1.1 现有页面复用情况

| 现有页面 | 功能 | 改动 |
|---------|------|------|
| `CharacterListPage.vue` | 角色列表 | **不改**：角色数量少时筛选无意义 |
| `CharacterEditPage.vue` | 角色编辑 | **小改**：新增 `scenic_category` 下拉框 + `recommended_routes` 多选勾选框 |
| `SettingsPage.vue` | 全局配置 | **不改** |
| `KnowledgeSourceManager.vue` | 知识库文档管理 | **不改**：直接复用 |
| `LandingPage.vue` | 落地页 | **不改** |
| `KanshanLandingPage.vue` | 看山落地页 | **不改** |
| `LaunchConfigPage.vue` | 启动配置 | **不改** |
| `SessionPage.vue` | 对话页 | **不改**：管理端不需要对话功能 |

### 1.2 新增页面

```
管理端（Vue 前端）
├── 现有页面（小改）
│   └── 角色编辑页（CharacterEditPage）← 新增 scenic_category 下拉框 + recommended_routes 多选勾选框
│
└── 新增页面
    ├── 景点管理页（AttractionManagePage）
    │    ├── 景点列表（表格 + 搜索 + 分类筛选）
    │    ├── 新增/编辑景点（表单：名称、简介、分类、坐标、图片、标签）
    │    └── 删除景点（确认弹窗）
    │
    ├── 路线管理页（RouteManagePage）
    │    ├── 路线列表（表格 + 标签筛选）
    │    ├── 新增/编辑路线（表单：名称、描述、时长、难度、标签 + 途经点拖拽排序）
    │    └── 删除路线（确认弹窗）
    │
    ├── 数据大屏页（DashboardPage）
    │    ├── 今日服务人次（数字卡片）
    │    ├── 本周服务趋势（折线图）
    │    ├── 热门问答 TOP 10（横向柱状图）
    │    ├── 情感分布（饼图）
    │    └── 活跃时段分布（柱状图）
    │
    ├── 对话日志页（SessionLogPage）
    │    ├── 会话列表（表格：时间、角色、时长、轮次、情感标签）
    │    └── 会话详情（点击展开：逐轮对话记录）
    │
    └── 感受度报告页（ReportPage）
         ├── 生成报告按钮（触发 SubAgent 任务）
         ├── 报告列表（历史报告）
         └── 报告详情（LLM 分析结果 + 图表）
```

### 1.3 管理端数据依赖顺序

管理端的使用顺序（数据依赖链）：

```
1. 创建景点（景点管理页）→ attractions 表有数据
2. 创建路线（路线管理页，关联景点）→ routes + route_steps 表有数据
3. 编辑角色（角色编辑页，关联路线）→ character.recommended_routes 有数据
4. 上传知识库文档（角色编辑页 → 知识库管理）→ RAG 索引有数据
5. 系统就绪，游客可以开始对话
```

### 1.4 路由扩展

在现有 `frontend/src/router/index.ts` 中新增路由：

```typescript
{ path: '/attractions', name: 'attractions', component: () => import('../pages/AttractionManagePage.vue') },
{ path: '/routes', name: 'routes', component: () => import('../pages/RouteManagePage.vue') },
{ path: '/dashboard', name: 'dashboard', component: () => import('../pages/DashboardPage.vue') },
{ path: '/sessions', name: 'sessions', component: () => import('../pages/SessionLogPage.vue') },
{ path: '/reports', name: 'reports', component: () => import('../pages/ReportPage.vue') },
```

### 1.5 导航菜单

在现有 `AppHeader.vue` 中新增导航入口：

```
现有菜单：角色列表 | 设置
新增菜单：景点管理 | 路线管理 | 数据大屏 | 对话日志 | 感受度报告
```

### 1.6 `scenic_category` 字段说明

`scenic_category` 是角色的分类标签，影响两个层面：

| 层面 | 影响 |
|------|------|
| 管理端展示 | 角色编辑页显示分类下拉框，便于管理员组织角色 |
| system_prompt | 不同分类有不同的行为约束（导游侧重路线推荐，讲解侧重景点知识，客服侧重服务信息） |
| 技术层面 | 无影响——不影响工具调用、RAG 检索、会话管理 |

---

## 第 2 节：各页面功能详述

### 2.1 景点管理页（AttractionManagePage）

**页面布局**：表格 + 弹窗表单（复用 CyberVerse 现有的弹窗编辑模式）

**景点列表表格**：

| 列 | 字段 | 说明 |
|----|------|------|
| 名称 | `name` | 景点名称 |
| 分类 | `category` | 自然风光 / 历史古迹 / 亲子游乐 |
| 所在位置 | `location` | 景点位置描述 |
| 简介 | `description` | 截断显示前 50 字 |
| 标签 | `tags` | 标签芯片展示 |
| 操作 | — | 编辑 / 删除 |

**新增/编辑表单字段**：

| 字段 | 控件类型 | 必填 | 说明 |
|------|---------|------|------|
| 名称 | 文本输入 | 是 | — |
| 简介 | 多行文本 | 是 | — |
| 分类 | 下拉选择 | 是 | 自然风光 / 历史古迹 / 亲子游乐 |
| 所在位置 | 文本输入 | 否 | "景区东部，莲花湖畔"（自然语言描述） |
| 图片 | 图片上传 | 否 | 管理端展示用 |
| 标签 | 标签输入 | 否 | 多标签，逗号分隔 |

**调用 API**：

| 操作 | API |
|------|-----|
| 加载列表 | `GET /api/v1/attractions` |
| 创建 | `POST /api/v1/attractions` |
| 更新 | `PUT /api/v1/attractions/:id` |
| 删除 | `DELETE /api/v1/attractions/:id` |

**删除约束**：删除景点时，如果该景点被某条路线的 `route_steps` 引用，提示管理员先从路线中移除该景点，或级联删除路线中的途经点。

### 2.2 路线管理页（RouteManagePage）

**路线列表表格**：

| 列 | 字段 | 说明 |
|----|------|------|
| 名称 | `name` | 路线名称 |
| 时长 | `duration` | "约2小时" |
| 难度 | `difficulty` | easy / moderate |
| 途经景点 | `steps` | 显示景点名称列表 |
| 标签 | `tags` | 标签芯片 |
| 操作 | — | 编辑 / 删除 |

**新增/编辑表单字段**：

| 字段 | 控件类型 | 必填 | 说明 |
|------|---------|------|------|
| 名称 | 文本输入 | 是 | — |
| 描述 | 多行文本 | 是 | — |
| 时长 | 文本输入 | 是 | "约2小时" |
| 难度 | 下拉选择 | 是 | easy / moderate |
| 标签 | 标签输入 | 否 | 多标签 |
| 途经景点 | 有序列表 | 是 | 从已有景点中选择，支持拖拽排序 |

**途经景点编辑器**：
- 点击"添加景点"按钮 → 弹出景点选择弹窗（从 attractions 表加载）
- 每个途经点显示：景点名称 + 建议停留时间（分钟）+ 重点讲解提示（可选文本）
- 支持拖拽调整顺序
- 支持删除途经点

**调用 API**：

| 操作 | API |
|------|-----|
| 加载列表 | `GET /api/v1/routes` |
| 加载景点选项 | `GET /api/v1/attractions`（编辑表单中选择途经景点用） |
| 创建 | `POST /api/v1/routes` |
| 更新 | `PUT /api/v1/routes/:id` |
| 删除 | `DELETE /api/v1/routes/:id` |

**删除约束**：删除路线时，如果某角色的 `recommended_routes` 引用了该路线，提示管理员先从角色中移除该路线引用。

### 2.3 角色编辑页改造（CharacterEditPage）

在现有角色编辑表单中新增两个字段：

| 字段 | 控件类型 | 位置 | 说明 |
|------|---------|------|------|
| 角色分类 | 下拉选择 | persona 配置区域 | 导游 / 讲解员 / 客服 |
| 推荐路线 | 多选勾选框 | 新增区域 | 从 routes 表加载可选路线，勾选该角色可推荐的路线 |

**推荐路线多选控件**：
- 页面加载时调用 `GET /api/v1/routes` 获取所有路线
- 以勾选框列表展示，每项显示路线名称 + 时长 + 难度
- 勾选结果写入 `character.recommended_routes` 字段
- 保存时调用 `PUT /api/v1/characters/:id`

### 2.4 数据大屏页（DashboardPage）

**页面布局**：卡片 + 图表，单页展示所有核心指标。

**数据来源**：`GET /api/v1/analytics/dashboard`

**展示组件**：

| 组件 | 数据 | 图表类型 |
|------|------|---------|
| 今日服务人次 | `today_sessions` | 数字卡片（大字体） |
| 本周服务趋势 | `week_sessions` + 每日明细 | 折线图（X轴: 日期, Y轴: 人次） |
| 热门问答 TOP 10 | `hot_questions` | 横向柱状图（X轴: 次数, Y轴: 问题摘要） |
| 情感分布 | `sentiment_distribution` | 饼图（positive/neutral/negative 占比） |
| 活跃时段分布 | `hourly_distribution` | 柱状图（X轴: 小时, Y轴: 会话数） |

**前端图表库**：ECharts（`npm install echarts`）

**刷新策略**：页面加载时请求一次，不自动轮询（比赛 demo 场景数据量小，手动刷新即可）。

### 2.5 对话日志页（SessionLogPage）

**会话列表表格**：

| 列 | 字段 | 说明 |
|----|------|------|
| 时间 | `started_at` | 会话开始时间 |
| 角色 | `character_id` → 角色名称 | 关联查询角色名 |
| 时长 | `duration_s` | 格式化为 "5分30秒" |
| 轮次 | `turn_count` | 对话轮次数 |
| 情感 | `sentiment` | 彩色标签：绿(positive) / 灰(neutral) / 红(negative) |
| 操作 | — | 查看详情 |

**会话详情（点击展开）**：

调用 `GET /api/v1/analytics/sessions/:id`，展示该 session 的逐轮对话：

```
[用户] 门票多少钱？（09:15:23）
[数字人] 成人票80元，学生票40元，1.2米以下儿童免费。（09:15:25）
[用户] 推荐一条亲子路线（09:16:10）
[数字人] 我推荐您试试亲子休闲线...（09:16:13）[路线卡片]
```

**筛选条件**：
- 角色（下拉选择）
- 日期范围（日期选择器）
- 情感标签（下拉选择）

**调用 API**：

| 操作 | API |
|------|-----|
| 加载列表 | `GET /api/v1/analytics/sessions?character_id=&date_from=&date_to=` |
| 查看详情 | `GET /api/v1/analytics/sessions/:id?format=admin` |

### 2.6 感受度报告页（ReportPage）

**页面布局**：按钮 + 列表 + 详情面板

**生成报告**：
- 点击"生成报告"按钮
- 调用 `POST /api/v1/reports/generate`
- 返回 SubAgent 任务 ID，页面显示"生成中..."状态
- 通过 WebSocket `task_event` 监听任务进度
- 任务完成后刷新报告列表

**报告列表**：

| 列 | 字段 | 说明 |
|----|------|------|
| 生成时间 | `created_at` | — |
| 分析范围 | 日期范围 / 角色 | — |
| 状态 | `status` | 生成中 / 已完成 / 失败 |
| 操作 | — | 查看详情 / 删除 |

**报告详情**：

LLM 分析结果展示（具体内容由 SubAgent 生成，管理端只负责渲染）：

| 模块 | 内容 |
|------|------|
| 游客关注点 | 高频关键词 / 热门景点排名 |
| 情感趋势 | 整体满意度 / 正面/负面反馈占比 |
| 知识盲区 | 游客问了但知识库没覆盖的问题 |
| 改进建议 | 基于分析的服务优化建议 |

**调用 API**：

| 操作 | API |
|------|-----|
| 生成报告 | `POST /api/v1/reports/generate` |
| 加载列表 | `GET /api/v1/reports` |
| 查看详情 | `GET /api/v1/reports/:id` |

---

## 第 3 节：前端改动量估算 + 技术风险

### 3.1 新增页面改动量估算

| 页面 | 估算行数 | 主要工作 |
|------|---------|---------|
| `AttractionManagePage.vue` | ~300 行 | 表格 + 弹窗表单 + CRUD API 调用 |
| `RouteManagePage.vue` | ~400 行 | 表格 + 弹窗表单 + 途经景点编辑器（拖拽排序） |
| `DashboardPage.vue` | ~350 行 | ECharts 图表初始化 + 数据绑定（5 个图表组件） |
| `SessionLogPage.vue` | ~250 行 | 表格 + 展开详情 + 筛选条件 |
| `ReportPage.vue` | ~200 行 | 按钮 + 列表 + 详情面板 + 任务状态监听 |

### 3.2 现有页面改动量估算

| 页面 | 估算行数 | 主要工作 |
|------|---------|---------|
| `CharacterEditPage.vue` | ~80 行 | 新增 scenic_category 下拉框 + recommended_routes 多选勾选框 |
| `AppHeader.vue` | ~30 行 | 新增 5 个导航菜单项 |
| `router/index.ts` | ~20 行 | 新增 5 条路由 |

### 3.3 API 调用层扩展

在现有 `frontend/src/services/api.ts` 中新增 API 调用函数：

| 函数 | 对应 API | 估算行数 |
|------|---------|---------|
| `listAttractions` / `getAttraction` / `createAttraction` / `updateAttraction` / `deleteAttraction` | `/api/v1/attractions` CRUD | ~50 行 |
| `listRoutes` / `getRoute` / `createRoute` / `updateRoute` / `deleteRoute` | `/api/v1/routes` CRUD | ~50 行 |
| `getDashboard` / `getSessions` / `getSessionDetail` / `getSentiment` | `/api/v1/analytics/*` | ~40 行 |
| `generateReport` / `listReports` / `getReport` | `/api/v1/reports/*` | ~30 行 |

### 3.4 前端总改动量

| 类型 | 估算行数 |
|------|---------|
| 新增 5 个页面 | ~1500 行 |
| 修改 3 个现有文件 | ~130 行 |
| API 调用层扩展 | ~170 行 |
| **合计** | **~1800 行** |

### 3.5 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| ECharts 包体积 | 管理端是桌面 Web，无影响；但增加 build 产物大小 | 按需引入 ECharts 模块（`echarts/core` + 按需注册图表类型） |
| 途经景点拖拽排序 | 需要拖拽库支持 | 使用 `vuedraggable`（Vue 3 版本）或用上下移动按钮代替拖拽 |
| 会话详情展开性能 | 单个 session 上百轮对话时渲染可能卡顿 | 分页加载对话记录，或虚拟滚动 |
| 报告生成等待体验 | SubAgent 生成报告可能需要 30 秒以上 | WebSocket `task_event` 实时推送进度，页面显示进度条 |

### 3.6 前端新增依赖

| 依赖 | 用途 | 是否必须 |
|------|------|---------|
| `echarts` | 数据大屏图表 | 是 |
| `vue-echarts` | Vue 3 ECharts 封装 | 推荐（简化集成） |
| `vuedraggable` | 路线途经点拖拽排序 | 可选（可用上下移动按钮代替） |

---

*Spec 3 管理端 Spec 完成。*
