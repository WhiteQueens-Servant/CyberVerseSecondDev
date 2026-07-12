# 执行日志（服务端验收指南）

> 本文件供服务器端 Claude Code 验收测试使用。每个 Task 包含：改动文件、验证命令、预期结果。
> 服务器端执行时，请按顺序逐条运行验证命令，确认输出与预期一致。

---

## T-01 Character 结构体扩展

**改动文件**：`server/internal/character/store.go`

**变更内容**：`Character` 结构体新增两个字段：
- `ScenicCategory string` — `json:"scenic_category,omitempty"`
- `RecommendedRoutes []string` — `json:"recommended_routes,omitempty"`

**验收步骤**：

```bash
# 1. 静态检查
cd /path/to/CyberVerse-main/server
go vet ./internal/character/...

# 2. 向后兼容性验证（检查旧 character.json 能否正常加载）
# 启动服务器后，调用现有 character API：
curl -s http://localhost:8080/api/v1/characters | python3 -m json.tool
# 预期：返回现有角色列表，新字段不存在或为空（不影响其他字段）

# 3. 新字段写入验证
# 创建一个带新字段的角色（或更新现有角色）：
curl -s -X PUT http://localhost:8080/api/v1/characters/<existing_id> \
  -H 'Content-Type: application/json' \
  -d '{"scenic_category":"导游","recommended_routes":["route1","route2"]}' | python3 -m json.tool
# 预期：返回的 JSON 中包含 scenic_category 和 recommended_routes
# 检查 data/characters/<dir>/character.json 文件确认字段已写入
```

**注意事项**：
- 新字段使用 `omitempty`，旧数据反序列化后为空值，不影响现有功能
- `load()` 函数中 nil 检查已覆盖 Tags/Images，新字段同理（string 零值为 ""，slice 零值为 nil）

---

## T-02 cyberverse.db 初始化与迁移

**新建文件**：`server/internal/scenic/db.go`

**改动文件**：`server/cmd/cyberverse-server/main.go`

**变更内容**：
- `scenic.OpenDB("data/cyberverse.db")` 在 main.go 中初始化
- 6 张表：`attractions`, `routes`, `route_steps`, `sessions`, `conversation_logs`, `reports`
- 退出时自动关闭连接

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功，无错误

# 2. 启动服务器
# 确保 data/ 目录存在
mkdir -p data
./cyberverse-server -config ../../cyberverse_config.yaml
# 预期日志：
# "Scenic database initialized: db=data/cyberverse.db"

# 3. 数据库文件验证
ls -la data/cyberverse.db
# 预期：文件存在，大小 > 0

# 4. 表结构验证
sqlite3 data/cyberverse.db ".tables"
# 预期输出：attractions  conversation_logs  reports  route_steps  routes  sessions

# 5. 重复启动幂等性验证
# 停止服务器后重新启动
# 预期：无报错，日志正常输出 "Scenic database initialized"

# 6. 现有 tasks.db 不受影响
sqlite3 data/tasks/tasks.db ".tables"
# 预期：tasks.db 仍然独立可用
```

**注意事项**：
- 使用 `modernc.org/sqlite` 驱动（纯 Go，无需 CGO）
- `db.SetMaxOpenConns(1)` 避免并发写冲突
- WAL 模式提升读性能
- `CREATE TABLE IF NOT EXISTS` 确保幂等
- `sessions.sentiment` 默认值为 `'neutral'`（非空字符串），确保 T-05 的筛选功能在 T-24 实现前即可正常工作
- 实际创建 6 张表：attractions, routes, route_steps, sessions, conversation_logs, reports

---

## T-03 景点 CRUD Store

**新建文件**：`server/internal/scenic/attraction.go`

**验收步骤**：

```bash
# 1. 静态检查
cd /path/to/CyberVerse-main/server
go vet ./internal/scenic/...

# 2. 单元测试
go test ./internal/scenic/ -run TestAttraction -v
# 预期：所有测试通过

# 3. API 集成验证（需先完成 T-07 路由注册）
# 创建景点
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"莲花湖","description":"景区核心湖泊，湖面如镜","category":"natural","location":"景区中心","tags":["湖泊","免费"]}' | python3 -m json.tool
# 预期：返回 201 + 完整景点对象（含 id, created_at）

# 查询列表
curl -s http://localhost:8080/api/v1/attractions | python3 -m json.tool
# 预期：返回包含刚创建景点的数组

# 更新景点
curl -s -X PUT http://localhost:8080/api/v1/attractions/<id> \
  -H 'Content-Type: application/json' \
  -d '{"name":"莲花湖（已更新）","description":"更新后的描述","category":"natural"}' | python3 -m json.tool
# 预期：返回更新后的对象

# 删除景点
curl -s -X DELETE http://localhost:8080/api/v1/attractions/<id>
# 预期：204 No Content

# 4. 删除约束验证
# 创建景点 A，创建路线引用景点 A，尝试删除景点 A
# 预期：返回 409 Conflict，提示景点被路线引用
```

---

## T-04 路线 CRUD Store

**新建文件**：`server/internal/scenic/route.go`

**验收步骤**：

```bash
# 1. 静态检查
cd /path/to/CyberVerse-main/server
go vet ./internal/scenic/...

# 2. 单元测试
go test ./internal/scenic/ -run TestRoute -v
# 预期：所有测试通过

# 3. API 集成验证（需先完成 T-07 路由注册）
# 先创建几个景点
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"莲花湖","description":"核心湖泊","category":"natural","tags":["湖泊"]}'
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"古戏台","description":"百年戏台","category":"historical","tags":["历史"]}'
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"萌宠乐园","description":"亲子互动区","category":"family","tags":["亲子"]}'

# 创建路线（使用上述景点 ID）
curl -s -X POST http://localhost:8080/api/v1/routes \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"亲子休闲线",
    "description":"适合带小朋友的家庭路线",
    "duration":"约2小时",
    "difficulty":"easy",
    "tags":["亲子","休闲"],
    "steps":[
      {"attraction_id":"<莲花湖id>","order":0,"duration_minutes":30,"highlight":"湖面倒影拍照"},
      {"attraction_id":"<古戏台id>","order":1,"duration_minutes":20,"highlight":"了解戏曲文化"},
      {"attraction_id":"<萌宠乐园id>","order":2,"duration_minutes":40,"highlight":"与小动物互动"}
    ]
  }' | python3 -m json.tool
# 预期：返回 201 + 完整路线对象（steps 中包含 attraction_name）

# 查询路线详情
curl -s http://localhost:8080/api/v1/routes/<id> | python3 -m json.tool
# 预期：steps 列表按 order 排序，每个 step 有 attraction_name

# 更新路线（修改途经点）
curl -s -X PUT http://localhost:8080/api/v1/routes/<id> \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"亲子休闲线（优化版）",
    "description":"优化后的路线",
    "duration":"约2.5小时",
    "difficulty":"easy",
    "steps":[
      {"attraction_id":"<莲花湖id>","order":0,"duration_minutes":40,"highlight":"增加停留时间"},
      {"attraction_id":"<萌宠乐园id>","order":1,"duration_minutes":50,"highlight":"新增互动项目"}
    ]
  }' | python3 -m json.tool
# 预期：steps 被替换为新的 2 个步骤（旧的古戏台步骤被移除）

# 4. 事务回滚验证
# 尝试创建包含不存在的 attraction_id 的路线
curl -s -X POST http://localhost:8080/api/v1/routes \
  -H 'Content-Type: application/json' \
  -d '{"name":"测试","description":"测试","duration":"1小时","difficulty":"easy","steps":[{"attraction_id":"nonexistent","order":0,"duration_minutes":30}]}'
# 预期：返回错误，routes 表和 route_steps 表均无新增数据
```

**注意事项**：
- `DeleteRoute` 的 character 引用检查：characters 表在文件系统（非 cyberverse.db），因此引用检查已从 Store 层移除，由 API handler 层（T-07）通过 `character.Store.List()` 实现
- `ErrRouteReferenced` 哨兵错误保留在 route.go 中，供 T-07 handler 使用
- `loadSteps` 使用 LEFT JOIN attractions，确保即使景点被删除，路线步骤仍可显示（attraction_name 为空）

---

## T-05 会话 + 对话日志 Store

**新建文件**：`server/internal/scenic/session.go`

**变更内容**：
- `SessionRecord` 结构体（ID, CharacterID, StartedAt, EndedAt, DurationS, TurnCount, Sentiment）
- `ConversationLog` 结构体（ID, SessionID, Role, Content, Timestamp）
- `SessionLogDetail` 组合结构体（Session + Messages）
- CRUD 方法：CreateSession, EndSession, AppendLog, ListSessions, GetSession, GetSessionDetail
- `CreateSession` 显式写入 `sentiment = 'neutral'`
- `EndSession` 自动计算 duration_s 和 turn_count
- `ListSessions` 支持按 character_id, date_from, date_to, sentiment 筛选

**验收步骤**：

```bash
# 1. 静态检查
cd /path/to/CyberVerse-main/server
go vet ./internal/scenic/...

# 2. 单元测试
go test ./internal/scenic/ -run TestSession -v
# 预期：所有测试通过

# 3. API 集成验证（需先完成 T-08 路由注册）
# 创建会话
curl -s -X POST http://localhost:8080/api/v1/analytics/sessions \
  -H 'Content-Type: application/json' \
  -d '{"character_id":"<char_id>"}' | python3 -m json.tool
# 预期：返回会话对象，sentiment = "neutral"

# 查询会话列表（按情感筛选）
curl -s "http://localhost:8080/api/v1/analytics/sessions?sentiment=neutral" | python3 -m json.tool
# 预期：返回所有刚创建的会话（默认 sentiment=neutral）

# 查询会话列表（按日期筛选）
curl -s "http://localhost:8080/api/v1/analytics/sessions?date_from=2026-06-17&date_to=2026-06-17" | python3 -m json.tool
# 预期：返回今天创建的会话

# 查询会话详情（含对话记录）
curl -s "http://localhost:8080/api/v1/analytics/sessions/<session_id>" | python3 -m json.tool
# 预期：返回 session 元数据 + messages 数组

# 4. sentiment 默认值验证
# 直接查数据库
sqlite3 data/cyberverse.db "SELECT id, sentiment FROM sessions LIMIT 5;"
# 预期：所有会话的 sentiment 为 'neutral'（非空字符串）

# 5. EndSession 验证
# 结束会话后再次查询
curl -s "http://localhost:8080/api/v1/analytics/sessions/<session_id>" | python3 -m json.tool
# 预期：ended_at 非空，duration_s > 0，turn_count = user 消息数量
```

**注意事项**：
- `CreateSession` 同时在 DB 默认值和代码层显式写入 `sentiment = 'neutral'`，双保险
- `EndSession` 的 turn_count 统计 role='user' 的消息数（一轮 = 一条用户消息）
- `ListSessions` 的 date_to 参数自动补全为当天 23:59:59，确保包含当天所有会话
- `GetSessionDetail` 的 messages 按 timestamp 升序排列（对话时间顺序）
- `AppendLog` 对 user 消息自动做规则预处理（去标点、去问词、截断），存入 `normalized_content` 列，原始 `content` 不变。assistant 消息的 `normalized_content` 为空字符串
- `conversation_logs` 表新增 `normalized_content TEXT NOT NULL DEFAULT ''` 列（Phase 2 review 时加入）

---

## T-06 Dashboard 聚合查询

**新建文件**：`server/internal/scenic/dashboard.go`

**变更内容**：
- `DashboardData` 结构体（TodaySessions, WeekSessions, HotQuestions, SentimentDistribution, HourlyDistribution）
- `GetDashboard()` 方法返回 5 个聚合指标
- 无数据时返回零值和空数组（非 null），避免前端 NPE
- week_sessions 固定返回 7 天（含 count=0 的天），hourly_distribution 固定返回 24 小时

**验收步骤**：

```bash
# 1. 静态检查
cd /path/to/CyberVerse-main/server
go vet ./internal/scenic/...

# 2. API 集成验证（需先完成 T-08 路由注册）
curl -s http://localhost:8080/api/v1/analytics/dashboard | python3 -m json.tool
# 预期返回结构：
# {
#   "today_sessions": 3,
#   "week_sessions": [{"date":"2026-06-11","count":0}, ..., {"date":"2026-06-17","count":3}],
#   "hot_questions": [{"question":"门票多少钱","count":5}, ...],
#   "sentiment_distribution": {"positive":1, "neutral":2, "negative":0},
#   "hourly_distribution": [{"hour":0,"count":0}, ..., {"hour":9,"count":3}, ...]
# }

# 3. 无数据验证（清空数据库或新库）
# 预期：today_sessions=0, week_sessions 全 0, hot_questions=[], sentiment 全 0, hourly 全 0

# 4. hot_questions 边界验证
# 当 user 消息不足 10 条时，返回实际数量（不报错）
```

**注意事项**：
- `hot_questions` 按 `normalized_content` 聚合（规则预处理后的问题文本），而非原始 `content`
- `normalized_content` 由 `AppendLog` 的 `normalizeQuestion()` 生成：去标点、去问词前缀、截断 100 字符
- `sentiment_distribution` 中 sentiment 不在 positive/negative 中的值（含 'neutral' 和空字符串）归入 Neutral
- `week_sessions` 从 7 天前到今天，按日期升序排列

---

## T-07 景点/路线 API Handler

**新建文件**：`server/internal/api/scenic_handler.go`

**变更内容**：
- 景点 CRUD handler（5 个端点）
- 路线 CRUD handler（5 个端点）
- `handleDeleteRoute` 在删除前遍历 `character.Store.List()` 检查 `RecommendedRoutes` 引用

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/

# 2. 景点 CRUD 全流程
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' \
  -d '{"name":"莲花湖","description":"景区核心湖泊","category":"natural","tags":["湖泊"]}' | python3 -m json.tool
# 预期：201 + 完整对象

curl -s http://localhost:8080/api/v1/attractions | python3 -m json.tool
# 预期：200 + JSON 数组

curl -s -X DELETE http://localhost:8080/api/v1/attractions/<id>
# 预期：204

# 3. 路线 CRUD 全流程
curl -s -X POST http://localhost:8080/api/v1/routes \
  -H 'Content-Type: application/json' \
  -d '{"name":"亲子线","description":"家庭路线","duration":"2小时","difficulty":"easy","steps":[{"attraction_id":"<id>","order":0,"duration_minutes":30}]}' | python3 -m json.tool
# 预期：201 + 完整对象（steps 含 attraction_name）

# 4. 路线删除引用检查
# 先将路线 ID 加入某角色的 recommended_routes，再尝试删除
# 预期：409 "route is referenced by character: xxx"

# 5. 错误处理
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H 'Content-Type: application/json' -d '{}'
# 预期：400 "name and description are required"
```

---

## T-08 Analytics + Report API Handler

**新建文件**：`server/internal/api/analytics_handler.go`, `server/internal/scenic/report.go`

**变更内容**：
- Dashboard API handler
- Sessions 列表 + 详情 handler（支持筛选参数）
- Report CRUD handler（List/Get/Create/Delete）
- `report.go`：Report 结构体 + CRUD 方法 + UpdateReportContent/UpdateReportStatus

**验收步骤**：

```bash
# 1. Dashboard
curl -s http://localhost:8080/api/v1/analytics/dashboard | python3 -m json.tool
# 预期：200 + 5 个聚合指标

# 2. Sessions 列表（带筛选）
curl -s "http://localhost:8080/api/v1/analytics/sessions?sentiment=neutral" | python3 -m json.tool
# 预期：200 + JSON 数组

# 3. Sessions 详情
curl -s http://localhost:8080/api/v1/analytics/sessions/<id> | python3 -m json.tool
# 预期：200 + {session: {...}, messages: [...]}

# 4. 报告生成
curl -s -X POST http://localhost:8080/api/v1/reports/generate \
  -H 'Content-Type: application/json' \
  -d '{"date_from":"2026-06-01","date_to":"2026-06-17"}' | python3 -m json.tool
# 预期：201 + report 对象（status=queued）

# 5. 报告列表
curl -s http://localhost:8080/api/v1/reports | python3 -m json.tool
# 预期：200 + JSON 数组

# 6. 报告删除
curl -s -X DELETE http://localhost:8080/api/v1/reports/<id>
# 预期：204
```

---

## T-09 路由注册 + 依赖注入

**改动文件**：`server/internal/api/router.go`, `server/cmd/cyberverse-server/main.go`

**变更内容**：
- `Router` 结构体新增 `scenicDB *scenic.DB` 字段
- `NewRouter` 签名新增 `scenicDB` 参数（在 configPath 和 taskServices 之间）
- `registerRoutes` 新增 17 个端点（attractions 5 + routes 5 + analytics 3 + reports 4）
- main.go 传入 `scenicDB` 实例

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 2. 启动服务器
./cyberverse-server -config ../../cyberverse_config.yaml
# 预期日志包含：
# "Scenic database initialized: db=data/cyberverse.db"

# 3. 新端点可用性
curl -s http://localhost:8080/api/v1/attractions
curl -s http://localhost:8080/api/v1/routes
curl -s http://localhost:8080/api/v1/analytics/dashboard
curl -s http://localhost:8080/api/v1/reports
# 预期：全部返回 200

# 4. 现有端点不受影响
curl -s http://localhost:8080/api/v1/health
curl -s http://localhost:8080/api/v1/characters
# 预期：正常返回

# 5. CORS 验证
curl -s -I -X OPTIONS http://localhost:8080/api/v1/attractions \
  -H "Origin: http://localhost:5173"
# 预期：Access-Control-Allow-Origin: *
```

**注意事项**：
- NewRouter 签名变更可能影响其他调用方（当前仅 main.go 一处调用）

---

## T-10 Orchestrator 会话生命周期钩子 + T-11 对话日志双写

**改动文件**：
- `server/internal/orchestrator/session.go` — +`dbSyncedCount`, +`dbSessionID`, +`DBSessionID()` 方法
- `server/internal/orchestrator/orchestrator.go` — +`scenicDB` 字段, +`SetScenicDB()`, +`syncSessionToDB()`
- `server/cmd/cyberverse-server/main.go` — `orch.SetScenicDB(scenicDB)`, OnSessionEnd 调用 `EndSession`

**核心设计**：
- 双写顺序：先写文件（主链路），再写 DB（管理端），DB 写入失败不影响主链路
- `syncSessionToDB` 在每次 `persistSessionConversation` 后调用
- 通过 `dbSyncedCount` 追踪已同步消息数，只写新增消息（避免重复 INSERT）
- DB session 记录在首次 sync 时创建（`CreateSession`），session ID 存入 `session.dbSessionID`
- 会话结束时 main.go 的 `OnSessionEnd` 回调调用 `scenicDB.EndSession` 更新时长/轮次

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 2. 发起一次对话
# 通过管理端或 curl 创建 session 并发送消息
# 对话结束后检查数据库：

# 3. sessions 表验证
sqlite3 data/cyberverse.db "SELECT id, character_id, duration_s, turn_count, sentiment FROM sessions;"
# 预期：有一条记录，duration_s > 0，turn_count > 0，sentiment = 'neutral'

# 4. conversation_logs 表验证
sqlite3 data/cyberverse.db "SELECT session_id, role, SUBSTR(content, 1, 30) FROM conversation_logs ORDER BY timestamp;"
# 预期：交替出现 role=user 和 role=assistant 记录

# 5. 文件系统不受影响
ls data/characters/<char_dir>/sessions/
# 预期：仍有 session.json 文件

# 6. 回归验证
# 原有对话功能正常：语音对话、文本对话、SubAgent 任务
```

**注意事项**：
- orchestrator 包新增对 `scenic` 包的 import，确保无循环依赖
- `dbSessionID` 是 DB 生成的 UUID，与 orchestrator 的 session ID 不同
- `EndSession` 使用 `dbSessionID`（非 orchestrator session ID）
- 所有 DB 操作均为 best-effort，错误仅 log 不返回

---

## T-12 LLM 上下文注入景点/路线数据

**改动文件**：
- `server/internal/orchestrator/orchestrator.go` — +`buildScenicContext()` 方法，修改 `standardSystemPrompt`、`standardSystemPromptWithRAG`、`buildVoiceLLMSessionConfig`

**核心设计**：
- `buildScenicContext()` 从 cyberverse.db 加载所有景点（名称列表）和所有路线（名称+时长+途经景点），格式化为结构化摘要
- 文本管线注入点：`standardSystemPrompt` 和 `standardSystemPromptWithRAG` 中，角色设定之后、RAG 检索结果之前
- 语音管线注入点：`buildVoiceLLMSessionConfig` 中，追加到 `char.SystemPrompt` 之后
- Best-effort：`scenicDB` 为 nil 或查询失败时返回空字符串，不影响原有流程
- 每次消息处理时实时读取（非会话级缓存），管理端修改后立即生效

**最终 system prompt 结构**：
```
【全局输出规范】→【角色设定】→【景区运营数据】→【角色素材检索结果(RAG)】
```

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 2. 准备测试数据（通过管理端 API）
# 创建景点
curl -s -X POST http://localhost:8080/api/v1/attractions \
  -H "Content-Type: application/json" \
  -d '{"name":"故宫","category":"历史","description":"明清皇家宫殿","open_hours":"8:30-17:00","tags":["世界遗产"]}'
# 创建路线
curl -s -X POST http://localhost:8080/api/v1/routes \
  -H "Content-Type: application/json" \
  -d '{"name":"皇家文化游","description":"经典路线","duration":"约2小时","difficulty":"简单","tags":["亲子"],"steps":[{"attraction_id":"<id>","order":1,"duration_minutes":60,"highlight":"太和殿"}]}'

# 3. 发起对话，验证 LLM 回复包含景区数据
# 发送消息："附近有什么好玩的？"
# 预期：LLM 回复中提及景点名称和路线信息

# 4. 空数据降级验证
# 删除所有景点和路线后再次对话
# 预期：对话正常进行，LLM 回复不包含景区数据段

# 5. 语音管线验证
# 发起语音对话
# 预期：语音回复中也包含景区数据（通过 persona/doubao 的 system prompt）
```

**注意事项**：
- `buildScenicContext` 每次消息处理时调用，对 SQLite 来说查询开销极小（demo 数据量）
- 景区数据注入不依赖 RAG 服务是否可用，两者独立
- 语音管线的 system prompt 注入是追加式的，不影响原有 `char.SystemPrompt` 内容

---

## T-13 PersonaAgent 新增工具定义 + T-14 工具处理器实现

**改动文件**：
- `inference/plugins/voice_llm/persona_agent.py` — +3 工具定义, +3 处理方法, +HTTP 客户端, 更新 `_execute_tool` 分发

**核心设计**：
- 3 个新工具：`get_attractions`（景点列表，可选 category 筛选）、`get_routes`（路线列表）、`get_attraction_detail`（单个景点详情）
- 工具处理器通过 httpx 调用 Go 后端 HTTP API（`/api/v1/attractions`, `/api/v1/routes`）
- Go API 地址通过 `GO_API_BASE` 环境变量配置，默认 `http://localhost:8080`
- 错误处理：API 不可达时返回友好中文提示，不崩溃

**验收步骤**：

```bash
# 1. Python 语法检查
cd /path/to/CyberVerse-main
python -c "from inference.plugins.voice_llm.persona_agent import PERSONA_TOOL_DEFINITIONS; print(len(PERSONA_TOOL_DEFINITIONS))"
# 预期：输出 8（原有 5 个 + 新增 3 个）

# 2. 工具 schema 验证
python -c "
from inference.plugins.voice_llm.persona_agent import PERSONA_TOOL_DEFINITIONS
for t in PERSONA_TOOL_DEFINITIONS:
    print(f'{t.name}: {list(t.parameters.get(\"properties\", {}).keys())}')
"
# 预期：
# create_task: ['description']
# get_task_status: []
# cancel_task: []
# retrieve_character_knowledge: ['query']
# get_attractions: ['category']
# get_routes: []
# get_attraction_detail: ['attraction_id']

# 3. Instructions 包含新工具指引
python -c "
from inference.plugins.voice_llm.persona_agent import PERSONA_AGENT_INSTRUCTIONS
assert 'get_attractions' in PERSONA_AGENT_INSTRUCTIONS
assert 'get_routes' in PERSONA_AGENT_INSTRUCTIONS
assert 'get_attraction_detail' in PERSONA_AGENT_INSTRUCTIONS
print('OK')
"
# 预期：OK

# 4. 端到端验证（需要 Go 后端 + Python 推理服务均运行）
# 发起语音对话，询问"景区有什么好玩的？"
# 预期：数字人调用 get_attractions，返回景点列表
# 询问"推荐一条路线"
# 预期：数字人调用 get_routes，返回路线信息

# 5. Go 后端不可用降级验证
# 停止 Go 后端，发起对话询问景点
# 预期：数字人返回"暂时无法获取景点信息，请稍后再试。"
```

**注意事项**：
- 新工具不影响原有 `create_task`/`get_task_status`/`cancel_task`/`retrieve_character_knowledge` 工具
- `_execute_tool` 中新工具匹配在 supervisor fallback 之前，确保不被误路由
- `httpx` 已是项目依赖（`pyproject.toml`），无需额外安装

---

## T-15 CharacterEditPage 扩展

**改动文件**：
- `server/internal/character/store.go` — Character 结构体 +`ScenicSystemPrompt` 字段
- `server/internal/orchestrator/orchestrator.go` — `characterSystemPrompt` 和语音管线使用 `ScenicSystemPrompt`
- `frontend/src/types/index.ts` — Character 接口 +3 字段
- `frontend/src/i18n/messages.ts` — +13 中英文 i18n 键
- `frontend/src/pages/CharacterEditPage.vue` — +表单字段, +路线加载, +景区配置区模板

**核心设计**：
- `scenic_category`：导游(guide) / 讲解员(narrator) / 客服(service)
- `scenic_system_prompt`：景区专属人设提示词，非空时覆盖原 `system_prompt`
- `recommended_routes`：角色关联的路线 ID 列表（checkbox 多选）
- orchestrator 构建 prompt 时：`ScenicSystemPrompt` 非空 → 替换 `SystemPrompt`；否则用原值

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 2. 前端构建验证
cd /path/to/CyberVerse-main/frontend
npm run build
# 预期：构建成功

# 3. 角色编辑页功能验证
# 打开管理端 → 角色管理 → 编辑角色
# 预期：看到"景区配置"区域，包含：
#   - 数字人类别下拉框（导游/讲解员/客服）
#   - 景区人设提示词 textarea
#   - 推荐路线勾选列表

# 4. 数据保存/加载验证
# 选择 scenic_category = 导游
# 填写 scenic_system_prompt = "你是XX景区专属导游..."
# 勾选 2 条路线 → 保存 → 重新打开页面
# 预期：所有字段保持不变

# 5. scenic_system_prompt 生效验证
# 发起对话
# 预期：数字人使用 scenic_system_prompt 的人设回复

# 6. scenic_system_prompt 为空降级验证
# 清空 scenic_system_prompt → 保存 → 发起对话
# 预期：数字人使用原 system_prompt 回复

# 7. 回归验证
# 确认现有角色编辑功能（名称、描述、语音等）不受影响
```

**注意事项**：
- `ScenicSystemPrompt` 使用 `omitempty`，旧 character.json 文件自动兼容（字段为空）
- 路线加载失败不影响页面渲染（仅显示"暂无路线"）
- `recommended_routes` 存储的是路线 ID 列表，不是路线对象

---

## T-16 景区入口页 ScenicLandingPage

**改动文件**：
- `frontend/src/pages/ScenicLandingPage.vue`（新建）— 景区入口页
- `frontend/src/router/index.ts` — +`/scenic` 路由
- `frontend/src/i18n/messages.ts` — +20 条中英文 i18n 键

**核心设计**：
- 动态加载导游角色：`getCharacters()` → 过滤 `scenic_category=guide` → 第一个匹配
- 头像优先级：`active_image`（API 路径）> `avatar_image`（URL）> SVG 占位符
- 景点/路线数据：`Promise.allSettled` 并行加载，任一失败不影响其他
- 会话创建：参照 KanshanLandingPage 的 `createSession()` + `buildSessionLaunchState()` 模式

**验收步骤**：

```bash
# 1. 前端构建验证
cd /path/to/CyberVerse-main/frontend
npm run build
# 预期：构建成功，无 TypeScript 错误

# 2. 路由验证
# 启动前端 dev server，访问 http://localhost:5173/scenic
# 预期：页面正常渲染，显示景区标题、导游卡片、景点/路线列表

# 3. 导游角色加载验证
# 确保有一个 scenic_category=guide 的角色
# 预期：页面显示该角色的头像和名称

# 4. 无导游角色降级
# 删除所有角色的 scenic_category
# 预期：页面显示"暂无可用的导游角色"提示

# 5. 开始对话验证
# 点击"开始对话"
# 预期：创建会话 → 跳转到 /session/{id} → 对话页正常工作

# 6. 返回路径验证
# 在对话页点击返回按钮
# 预期：回到 /scenic（而非 /characters）
```

**注意事项**：
- 景点卡片的 `image_url` 为空时显示灰色占位符，布局不断裂
- 路线的 `attraction_name` 为可选字段，使用 `attraction_id` 作为 fallback
- 页面使用绿色主题（--primary: #16a34a），与管理端蓝色主题区分

---

## T-18 ChatPanel 路线/景点卡片渲染

**改动文件**：
- `frontend/src/components/ScenicCard.vue`（新建）— 可展开/折叠的卡片组件
- `frontend/src/components/ChatPanel.vue` — +ScenicCard import, +解析函数, +模板条件渲染

**核心设计**：
- 解析逻辑：在 assistant 消息的 `msg.content` 中检测路线/景点关键词模式
  - 路线检测：`推荐...路线`、`途经：`、`时长：` 等模式
  - 景点检测：`推荐...景点`、`分类：` 等模式
- 匹配到 → 渲染 ScenicCard（折叠态摘要 + 展开态详情）
- 未匹配到 → 保持原有纯文本渲染
- 展开/折叠：通过 `expanded` ref 控制，纯前端状态切换
- 数据来源：LLM 消费 DB（T-12 注入）+ RAG 后的文字输出，前端不做额外请求

**验收步骤**：

```bash
# 1. 前端构建验证
npm run build
# 预期：构建成功

# 2. 对话中路线卡片渲染
# 发起对话，问"推荐一条路线"
# 预期：数字人回复中包含路线卡片（绿色左边框，显示名称+时长+途经景点）

# 3. 对话中景点卡片渲染
# 问"有什么好玩的景点"
# 预期：数字人回复中包含景点卡片（蓝色左边框，显示名称+分类）

# 4. 点击展开/折叠
# 点击路线卡片
# 预期：展开显示完整详情（描述、途经景点详情等）
# 再次点击
# 预期：折叠回摘要视图

# 5. 普通消息不受影响
# 问"你好"
# 预期：纯文本渲染，无卡片

# 6. 历史消息兼容
# 刷新页面，加载历史对话
# 预期：历史消息中的路线/景点也正确渲染为卡片
```

**注意事项**：
- 解析函数 `parseScenicCards()` 基于关键词匹配，非严格格式解析
- 卡片样式使用暗色主题（与 ChatPanel 的 #1e1e1e 背景一致）
- ScenicCard 的 `type` 字段决定左边框颜色：路线=绿色，景点=蓝色，混合=橙色

---

## Phase 1~6 代码审查修复（4 项）

### Fix #1: RouteManagePage 缺少 duration 校验

**改动文件**：`frontend/src/pages/RouteManagePage.vue`（第 163 行）

**问题**：`handleSubmit` 校验条件缺少 `duration` 字段，允许空 duration 提交。

**修复**：
```typescript
// 修复前
if (!form.value.name || !form.value.description || form.value.steps.length === 0) return
// 修复后
if (!form.value.name || !form.value.description || !form.value.duration || form.value.steps.length === 0) return
```

**验证**：新建路线时留空"时长"字段 → 提交按钮不触发请求。

---

### Fix #2: generateReport 返回类型错误

**改动文件**：`frontend/src/services/api.ts`

**问题**：`generateReport` 返回类型声明为 `{ task_id: string }`，但后端 `/reports/generate` 实际返回完整 `Report` 对象（status=queued）。`reporter.analyze()` 传入 `response.id` 实际为 `undefined`。

**修复**：返回类型从 `{ task_id: string }` 改为 `Report`。

```typescript
// 修复前
export async function generateReport(dateFrom: string, dateTo: string): Promise<{ task_id: string }> {
// 修复后
export async function generateReport(dateFrom: string, dateTo: string): Promise<Report> {
```

**验证**：生成分析报告 → 列表页出现 status=completed 的报告记录（而非报错）。

---

### Fix #3: 半角冒号正则（确认无需修复）

**涉及文件**：`server/internal/scenic/dashboard.go`

**检查结果**：`ROUTE_SPLIT_PATTERN = `[：:]`` 已同时匹配全角冒号（U+FF1A）和半角冒号（U+003A）。无需修改。

---

### Fix #4: SessionLog.character_name 字段注释

**改动文件**：`frontend/src/types/index.ts`

**问题**：`SessionLog` 类型声明了 `character_name?: string`，但后端 `ListSessions` SQL 查询不包含该列（sessions 表无 character_name 列），Go 结构体的 `omitempty` 标签导致该字段始终从 JSON 响应中省略。前端实际通过 `getCharacterName(s.character_id)` fallback 获取角色名。

**修复**：添加注释说明该字段为预留字段，当前后端不填充。

```typescript
// Currently not populated by backend (sessions table has no character_name column).
// The UI falls back to getCharacterName(character_id) for display.
character_name?: string
```

**验证**：SessionLogPage 角色名列仍正常显示（通过 ID 映射）。

---

## Fix #5: Go 服务器前端静态文件托管（SPA Fallback）

**改动文件**：`server/internal/api/router.go`

**问题**：Go 服务器只处理 API 请求（`/api/`、`/ws/`），访问前端路由（如 `/scenic`、`/attractions`）返回 404。原方案依赖 Nginx 托管前端，增加了部署复杂度。

**修复**：在 `Handler()` 中新增 `spaFallback` 中间件，将前端静态文件托管内嵌到 Go 服务器。

```go
func (r *Router) Handler() http.Handler {
    return corsMiddleware(r.spaFallback(r.mux))
}
```

`spaFallback` 逻辑：
- `/api/`、`/ws/` 请求 → 走原有 API 路由
- 其他请求 → 从 `frontend/dist/` 返回对应静态文件
- 文件不存在 → 返回 `frontend/dist/index.html`（Vue SPA 入口，由 Vue Router 在浏览器端渲染对应页面）

若 `frontend/dist/` 不存在（纯 API 模式），自动跳过，不影响原有功能。

**验证**：
1. `npm run build` 后启动 Go 服务器
2. 访问 `http://localhost:8080/scenic` → 景区入口页正常渲染
3. 访问 `http://localhost:8080/attractions` → 景点管理页正常渲染
4. 访问 `http://localhost:8080/api/v1/health` → API 正常响应
5. 删除 `frontend/dist/` 后重启 → API 正常，页面返回 404（预期行为）

---

## T-22 游客端 UI 风格重构（泉城主题）

**改动文件**：
- `frontend/src/components/ChatPanel.vue` — +`theme` prop, +CSS 变量双主题
- `frontend/src/pages/SessionPage.vue` — +`isScenicMode`, +`scenic-mode` CSS class
- `frontend/src/router/index.ts` — `/` 重定向至 `/scenic`

**核心设计**：
- ChatPanel 新增 `theme` prop（`'light' | 'dark'`，默认 `'dark'`），通过 `--chat-*` CSS 变量实现双主题
- SessionPage 通过 `launchState.returnPath === '/scenic'` 判断是否为景区模式
- 景区模式下：ChatPanel 传入 `theme="light"`，侧栏/控制栏/按钮改为浅色泉城风格
- 管理端模式下：完全保持原有深色风格不变
- AppHeader 不需要修改（仅管理端页面使用，保持深色）

**颜色体系**（泉城主题）：
- 主色：`#16a34a`（绿色，与 ScenicLandingPage 一致）
- 背景：`#f8faf8`（极浅绿）
- 边框：`#d4e8d4`（浅绿）
- 用户消息气泡：`#16a34a`（绿色）
- 助手消息气泡：`#ffffff`（白色）+ 浅绿边框
- 输入框：白色背景 + 浅绿边框
- 控制栏：`rgba(255,255,255,0.88)` 半透明白色
- 返回/FPS 按钮：白色毛玻璃

**验收步骤**：

```bash
# 1. 前端构建验证
cd /path/to/CyberVerse-main/frontend
npm run build
# 预期：无新增 TypeScript 错误（预存错误不影响）

# 2. 游客入口验证
# 访问 http://localhost:5173/ → 自动跳转到 /scenic
# 预期：ScenicLandingPage 正常渲染（绿白泉城风格）

# 3. 游客对话验证
# 点击"开始对话" → 进入 SessionPage
# 预期：
#   - 视频区域保持黑色
#   - 侧栏背景为浅色（#f8faf8），边框为浅绿
#   - ChatPanel 消息区域为浅色背景
#   - 用户消息气泡为绿色（#16a34a）
#   - 助手消息气泡为白色
#   - 输入框为白色背景
#   - 控制栏为半透明白色
#   - 返回按钮为白色毛玻璃风格

# 4. 管理端回归验证
# 访问 http://localhost:5173/characters → 编辑角色 → 启动对话
# 预期：SessionPage + ChatPanel 保持原有深色风格

# 5. 双主题切换验证
# 同时打开两个浏览器标签：
#   标签 A：从 /scenic 进入对话 → 浅色
#   标签 B：从 /characters 进入对话 → 深色
# 预期：两个会话的主题互不影响
```

---

## T-23 会话结束情感标注

**改动文件**：
- `server/internal/scenic/session.go` — +`UpdateSentiment()` 方法
- `server/internal/orchestrator/orchestrator.go` — +`AnalyzeSessionSentiment()` 方法, +`normalizeSentiment()`
- `server/cmd/cyberverse-server/main.go` — OnSessionEnd 回调追加 `go orch.AnalyzeSessionSentiment()`

**核心设计**：
- 会话结束时，异步启动 goroutine 调用 LLM 分析对话情感
- 不阻塞会话关闭，best-effort 失败仅 log
- LLM 通过 `inference.GenerateLLMStream()` gRPC 调用 Python 推理服务
- temperature=0.1 确保分类确定性
- 只取最后 20 条消息，避免超长对话溢出 token

**变更详情**：

1. `session.go` 新增 `UpdateSentiment(ctx, sessionID, sentiment)` 方法
   - 校验 sentiment 值必须为 positive/neutral/negative，否则默认 neutral
   - 执行 `UPDATE sessions SET sentiment = ? WHERE id = ?`

2. `orchestrator.go` 新增 `AnalyzeSessionSentiment(dbSessionID, history)` 方法
   - 拼接最后 20 条消息为"角色: 内容"格式
   - 构建情感分析 prompt
   - 调用 `o.inference.GenerateLLMStream()` 获取 LLM 回复
   - 解析回复 → `normalizeSentiment()` → positive/neutral/negative
   - 调用 `o.scenicDB.UpdateSentiment()` 写回 DB

3. `orchestrator.go` 新增 `normalizeSentiment(text)` 函数
   - 包含 "positive" → "positive"
   - 包含 "negative" → "negative"
   - 其他 → "neutral"

4. `main.go` OnSessionEnd 回调追加一行
   - `go orch.AnalyzeSessionSentiment(dbSID, history)`
   - 在 `scenicDB.EndSession()` 成后执行

**验收步骤**：

```bash
# 1. 编译验证
cd /path/to/CyberVerse-main/server
go build -tags livekit ./cmd/cyberverse-server/
# 预期：编译成功

# 2. 正面情感测试
# 发起对话，发送"这个景区太美了！我非常开心！"→ 结束会话
sqlite3 data/cyberverse.db "SELECT id, sentiment FROM sessions ORDER BY started_at DESC LIMIT 1;"
# 预期：sentiment = 'positive'

# 3. 负面情感测试
# 发起对话，发送"服务太差了，再也不来了"→ 结束会话
sqlite3 data/cyberverse.db "SELECT id, sentiment FROM sessions ORDER BY started_at DESC LIMIT 1;"
# 预期：sentiment = 'negative'

# 4. 中性情感测试
# 发起对话，发送"趵突泉在哪里？"→ 结束会话
sqlite3 data/cyberverse.db "SELECT id, sentiment FROM sessions ORDER BY started_at DESC LIMIT 1;"
# 预期：sentiment = 'neutral'

# 5. Dashboard 验证
curl -s http://localhost:8080/api/v1/analytics/dashboard | python3 -m json.tool
# 预期：sentiment_distribution 中 positive/negative/neutral 桶有数据

# 6. 会话关闭不卡顿验证
# 结束会话后，确认会话立即关闭（不等待 LLM 分析完成）

# 7. LLM 不可用降级验证
# 停止 Python 推理服务 → 结束会话
# 预期：会话正常关闭，sentiment 保持 'neutral'（默认值），日志输出错误
```
