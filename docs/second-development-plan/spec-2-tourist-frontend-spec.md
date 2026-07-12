# Spec 2: 游客端 Spec（微信小程序）

> 基于 Spec 1 系统级业务 Spec，设计游客端微信小程序的页面结构、交互流程和技术方案。

---

## 第 1 节：页面结构

### 1.1 页面清单

小程序共 4 个页面，纯对话驱动，无独立景点列表入口：

```
小程序入口（扫码/搜索）
  │
  ├─ 首页（角色选择）
  │    └── 展示景区导游角色列表 → 点击进入对话
  │
  ├─ 对话页（核心页面）
  │    ├── 语音交互区（麦克风按钮 + 语音波形）
  │    ├── 文本输入区（聊天输入框）
  │    ├── 数字人回复区（语音播放 + 文字显示）
  │    ├── 拍照识景入口（相机按钮）
  │    └── 景点卡片 / 路线卡片（工具调用结果展示，可点击跳转详情）
  │
  ├─ 景点详情页（从对话页卡片或路线详情页跳转）
  │    └── 景点名称、简介全文、缩略图、分类、标签、关联路线
  │
  └─ 路线详情页（从对话页路线卡片跳转）
       ├── 路线名称、描述、预计时长、难度标签
       └── 途经景点列表（有序，每项可点击跳转景点详情页）
```

### 1.2 页面职责

| 页面 | 功能 | Spec 1 对应 |
|------|------|-------------|
| 首页 | 角色列表展示，选择导游进入对话 | F9（角色管理的数据展示） |
| 对话页 | 语音/文本对话、拍照识景、路线/景点卡片展示 | F1-F7 |
| 景点详情页 | 展示景点完整信息（深层），两种路径可达 | F5（拍照识景结果下钻） |
| 路线详情页 | 展示路线途经点、景点列表（可点击下钻） | F4（个性化推荐结果下钻） |

### 1.3 信息展示层级

同一个 Attraction 对象，两种展示深度：

| 层级 | 展示位置 | 字段 | 数据来源 |
|------|---------|------|---------|
| 浅层（卡片） | 对话页内嵌 | name、description（截断）、image_url（缩略图） | Qwen-VL 返回文本 + attractions 表基础字段 |
| 深层（详情页） | 点击卡片跳转 | name、description（全文）、category、coordinates、tags、关联路线 | attractions 表完整字段 + routes 表关联查询 |

到达景点详情页的两条路径：

| 路径 | 场景 |
|------|------|
| 对话页 → 点击景点卡片 → 景点详情页 | 拍照识景 / 数字人介绍某个景点时 |
| 对话页 → 点击路线卡片 → 路线详情页 → 点击途经景点 → 景点详情页 | 看到路线推荐后，想深入了解某个途经景点 |

两条路径汇聚到同一个详情页，数据来源相同（attractions 表），不需要额外实现。

### 1.4 与 CyberVerse 前端的关系

CyberVerse 现有 Vue 前端只支持桌面 Web 端，不兼容移动端。证据：无移动端适配代码、无 `isMobile`/`responsive` 相关逻辑、WebRTC 组件依赖浏览器 API、布局用固定像素宽度。

游客端小程序所有代码需重新实现，复用的是**设计逻辑**（API 接口、WebRTC 连接流程、状态机），不是代码。

---

## 第 2 节：交互流程

### 2.1 核心交互流程

```
游客扫码/搜索打开小程序
  ↓
首页：展示景区导游角色列表（头像 + 名称 + 简介）
  ↓ 点击某个角色
对话页：自动建立 WebSocket 连接，播放欢迎语音
  ↓
游客按住麦克风说话 / 在输入框输入文字
  ↓
数字人语音/文字回复（支持被打断）
  ↓
游客继续对话 / 触发工具调用 / 拍照识景
```

### 2.2 推荐路线流程

```
游客语音："推荐一条亲子路线"
  ↓
PersonaAgent 判断 → 调用 get_routes 工具
  ↓
对话页内嵌路线卡片（浅层：路线名称 + 时长 + 途经景点摘要）
  ↓ 游客点击卡片
路线详情页（深层：完整描述 + 途经景点有序列表 + 每个景点可点击）
  ↓ 游客点击某个景点
景点详情页（深层：景点完整信息）
  ↓ 返回
回到对话页，对话上下文不丢失
```

### 2.3 拍照识景流程

```
对话页底部工具栏 → 相机按钮
  ↓
调用 wx.chooseMedia({ sourceType: ['camera', 'album'] })
  ↓
图片 base64 编码 → 通过 WebSocket 发送给后端
  ↓
PersonaAgent 调用 recognize_scenic_spot → 返回景点介绍
  ↓
对话页内嵌景点卡片（浅层：名称 + 简介摘要 + 缩略图）
  ↓ 游客点击卡片
景点详情页（深层：完整信息）
  ↓ 返回
回到对话页，对话上下文不丢失
```

### 2.4 对话页状态机

```
disconnected → connecting → connected → listening ↔ speaking
                                    ↓
                              idle_timeout → disconnected
```

| 状态 | 游客端表现 | 触发条件 |
|------|-----------|---------|
| disconnected | 显示"连接中..."，禁用麦克风 | 初始状态 / WebSocket 断开 |
| connecting | 显示加载动画 | 点击角色进入对话页 |
| connected | 播放欢迎语音，启用麦克风 | WebSocket 连接成功 |
| listening | 麦克风图标高亮，显示语音波形 | 游客按住说话 / 数字人说完等待输入 |
| speaking | 数字人头像动效，显示回复文字，麦克风禁用 | 数字人正在回复 |
| idle | 显示"点击继续对话" | 空闲超时 |

### 2.5 打断机制

游客在数字人说话时（speaking 状态）按住麦克风 → 立即停止音频播放 → 切换到 listening → 数字人被打断。小程序端只需在 speaking 状态下允许麦克风触发。

### 2.6 不在 Spec 2 阶段定义的内容

以下交互在 writing-plans 阶段定义为端到端验收用例：

- 语音问答的基础对话流程（CyberVerse 已验证）
- 文本输入对话流程
- 混合语音+文本输入
- 后台任务（SubAgent）的触发与结果展示
- 错误处理（网络断开、API 超时、麦克风权限拒绝）

---

---

## 第 3 节：API 调用清单 + 技术选型

### 3.1 小程序调用的后端 API 清单

| 功能 | API | 方法 | 触发时机 |
|------|-----|------|---------|
| 角色列表 | `/api/v1/characters` | GET | 首页加载时 |
| 角色详情 | `/api/v1/characters/:id` | GET | 点击角色进入对话页时 |
| 景点详情 | `/api/v1/attractions/:id` | GET | 点击景点卡片时 |
| 路线详情 | `/api/v1/routes/:id` | GET | 点击路线卡片时 |
| 对话通信 | WebSocket `/ws` | WS | 对话页全程保持连接 |

**小程序不直接调用的 API**（由 PersonaAgent 工具处理器在后端调用）：
- `get_routes` 工具 → Go 端 `character-routes` API
- `recognize_scenic_spot` 工具 → Go 端 `attractions` API + Qwen-VL API

### 3.2 WebSocket 事件清单

| 事件方向 | 事件名 | 数据 | 说明 |
|---------|--------|------|------|
| 小程序 → 服务端 | `text_input` | `{ text: string }` | 文本输入 |
| 小程序 → 服务端 | `interrupt` | 空 | 打断数字人说话 |
| 服务端 → 小程序 | `transcript` | `{ speaker, text, is_final, turn_seq }` | 语音转录文字（实时流式） |
| 服务端 → 小程序 | `llm_token` | `{ accumulated, is_final, turn_seq }` | LLM 文字输出（实时流式） |
| 服务端 → 小程序 | `assistant_message` | `{ text, turn_seq }` | 完整回复消息 |
| 服务端 → 小程序 | `tool_result` | `{ tool: string, data: object }` | 工具调用结果（路线/景点 JSON，前端渲染为卡片） |
| 服务端 → 小程序 | `avatar_status` | `{ status: "idle"/"speaking" }` | 数字人状态变更 |
| 服务端 → 小程序 | `task_event` | `{ task_id, event_type, ... }` | 后台任务进度 |

**卡片渲染**：`tool_result` 事件中的 `tool` 字段决定渲染哪种卡片：
- `tool: "get_routes"` → 路线卡片组件
- `tool: "recognize_scenic_spot"` → 景点卡片组件

### 3.3 音频传输方案

**采用 LiveKit 小程序 SDK**，不改动 CyberVerse 现有媒体管道。

```
微信小程序（原生）
  │
  ├── 文字/控制消息 ──→ WebSocket ──→ Go 服务 :8080
  │
  └── 音频流 ──→ LiveKit 小程序组件 ──→ LiveKit 服务器（外部）
                                               ↑
                                          Go 服务的 Bot 也连接同一个 LiveKit 服务器
```

**完整音频推理链路**：

```
小程序录音 → LiveKit Room → Go Bot（订阅用户音频）
  → Orchestrator → gRPC → Python 推理（Qwen-Omni）
  → Orchestrator → Go Bot（发布音频）→ LiveKit Room → 小程序播放
```

**不需要改 Go 端**——Bot 和 Orchestrator 已经实现了完整的音频收发链路，小程序只是 LiveKit Room 中的另一个参与者。

**需要的基础设施**：
- LiveKit 服务器实例（自托管 `livekit-server` 或 LiveKit Cloud）
- `cyberverse_config.yaml` 中配置 `livekit.url`、`livekit.api_key`、`livekit.api_secret`
- 小程序中使用 LiveKit 小程序组件

### 3.4 数字人回复的三条数据流并行

现有链路支持"语音+文字+卡片"的完整交互，三条数据流各司其职：

```
数字人回复时，同时走三条通道：

1. LiveKit Room → 音频流 → 小程序播放语音
2. WebSocket → transcript/llm_token 事件 → 对话框显示文字
3. WebSocket → tool_result 事件 → 对话框渲染卡片（如有工具调用）
```

三条流并行，不互相阻塞。语音通过 LiveKit 实时播放，文字和卡片通过 WebSocket 推送到对话框。

### 3.5 小程序技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 小程序框架 | 微信小程序原生 | 只做小程序，不需要 uni-app 跨端 |
| 状态管理 | `Behavior` 或 `mobx-miniprogram` | 对话状态、连接状态管理 |
| 音频录制 | `wx.getRecorderManager()` | PCM 格式，16kHz 采样率 |
| 音频播放 | LiveKit 小程序组件 | 通过 LiveKit Room 接收音频 |
| 网络通信 | `wx.connectSocket()` | WebSocket 长连接（文字+控制） |
| LiveKit | LiveKit 小程序 SDK / 组件 | 音频流传输 |
| 图片选择 | `wx.chooseMedia()` | 拍照/相册选图 |
| UI 组件 | 自定义组件 | 对话气泡、卡片、波形动画 |

### 3.6 小程序与后端的连接拓扑

```
微信小程序
  │
  ├── WebSocket (WSS) ──→ Go 服务 :8080 ──→ 文字消息 + 控制信号
  │
  └── LiveKit ──→ LiveKit 服务器 ──→ Go Bot ──→ Orchestrator ──→ Python 推理 :50051
```

小程序通过 WSS 访问 Go 服务（需要域名 + SSL 证书），通过 LiveKit 连接音频通道。Go 服务内部通过 gRPC 调用 Python 推理服务。小程序不直接接触 Python 推理服务。

---

---

## 第 4 节：技术风险与注意事项

### 4.1 LiveKit 小程序适配风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| LiveKit 官方无小程序 SDK | 需要使用社区方案或自行封装 | 优先评估 `livekit-miniprogram-sdk` 社区包；备选：直接用 `<live-pusher>` + `<live-player>` 组件封装 |
| `<live-pusher>` / `<live-player>` 机型兼容性 | Android 低端机可能出现音频卡顿/无声 | 提前在主流机型测试（iPhone、华为、小米）；准备降级方案（WebSocket 音频流） |
| 小程序后台限制 | 切后台时 LiveKit 连接可能断开 | 监听 `wx.onHide` / `wx.onShow`，切后台时暂停音频，切回时重连 Room |
| LiveKit 服务器部署 | 比赛 demo 需要一个可访问的 LiveKit 实例 | 本地开发用 `livekit-server` Docker 镜像；演示时部署在云服务器或用 LiveKit Cloud 免费版 |

### 4.2 WebSocket 域名与证书

微信小程序要求：
- WebSocket 必须使用 WSS（加密连接）
- 域名必须在小程序后台配置的合法域名列表中
- 不能使用 IP 地址，必须有域名

**缓解措施**：比赛 demo 可以用微信开发者工具的"不校验合法域名"选项跳过限制。正式部署需要域名 + SSL 证书。

### 4.3 麦克风权限

微信小程序录音需要用户授权 `scope.record`。如果用户拒绝授权，需要引导用户手动开启。

### 4.4 音频格式兼容

| 环节 | 格式 | 注意事项 |
|------|------|---------|
| 小程序录音 | PCM 16kHz（wx.getRecorderManager） | 需确认帧大小和采样率与 LiveKit Bot 的 PCMRemoteTrack 兼容 |
| LiveKit 传输 | Opus（WebRTC 标准） | 小程序 `<live-pusher>` 默认使用 Opus 编码 |
| Go Bot 接收 | Opus → PCM 16kHz（已有解码逻辑） | `onTrackSubscribed` 中已实现 |
| Python 推理 | PCM 16kHz | Qwen-Omni 期望 16kHz PCM 输入 |

格式链路已打通，不需要额外转换。但需测试小程序 `<live-pusher>` 输出的 Opus 参数（码率、声道数）是否和 Go Bot 的解码器兼容。

### 4.5 对话上下文保持

小程序页面栈有限（默认 10 层），从对话页跳转到景点详情页再返回，对话页的状态需要保持。

**方案**：对话页使用 `onShow` / `onHide` 生命周期管理 WebSocket 连接——跳转详情页时不断开连接，返回时恢复状态。LiveKit Room 连接同理。

### 4.6 不在 Spec 2 范围内的技术问题

以下问题在 writing-plans 阶段解决：
- LiveKit 小程序 SDK 的具体选型和集成步骤
- 卡片组件的具体 UI 实现（WXML/WXSS）
- 错误处理和重连策略的详细设计
- 性能优化（音频缓冲、渲染优化）

---

*Spec 2 游客端 Spec 完成。*
