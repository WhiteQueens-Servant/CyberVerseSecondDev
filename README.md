<h1 align="center">泉城济南 · 数字人 AI 导游</h1>
<p align="center"><em>实时数字人智能导览系统——基于自研多模态 Agent 框架，支持语音对话、智能问答、路线推荐、拍照识景、数据大屏等完整景区服务能力。</em></p>
<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3-blue.svg" alt="License: GPL v3"/></a>
  <img src="https://img.shields.io/badge/Python-3.10+-blue.svg" alt="Python 3.10+"/>
  <img src="https://img.shields.io/badge/Go-1.25-blue.svg" alt="Go 1.25"/>
  <img src="https://img.shields.io/badge/Node-18+-green.svg" alt="Node 18+"/>
</p>

---

## 项目简介

**泉城济南 · 数字人 AI 导游**是一个面向景区导览场景的实时数字人交互系统。游客可通过语音或文字与 AI 数字人导游实时对话，获取景点介绍、路线推荐、门票查询等服务；管理端提供景点管理、路线管理、数据大屏、感受度报告等运营能力。

系统采用 **Go + Python 双引擎架构**：Go 负责高并发网络通信、会话编排与 API 服务；Python 负责大模型推理、语音合成、数字人视频生成等 AI 任务。两端通过 gRPC 实现跨语言实时通信，前端通过 WebRTC 实现低延迟音视频传输。

### 核心亮点

| 亮点 | 说明 |
|------|------|
| **实时语音对话** | 基于 Qwen-Omni 端到端语音模型，延迟 < 3 秒，支持语音打断 |
| **智能工具调用** | PersonaAgent 自动调用景点查询、路线推荐等工具，无需游客手动操作 |
| **RAG 知识库问答** | 灌入景区文档 + FAQ，数字人可准确回答门票、交通、景点历史等问题 |
| **拍照识景** | 基于 Qwen-VL 多模态模型，拍照即可识别景点并返回介绍 |
| **数据大屏** | 实时展示服务人次、热门问题、情感分布、活跃时段等运营数据 |
| **感受度报告** | SubAgent 异步分析对话日志，生成游客满意度报告 |
| **双端架构** | 游客端（`/scenic`）+ 管理端（`/admin/`），单个 Go 服务统一托管 |

---

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    游客端 Web / 管理端 Web                 │
│              (Vue 3 + Vite + ECharts)                    │
│   /scenic → 景区入口页        /admin/* → 管理后台          │
└────────────────────┬────────────────────────────────────┘
                     │  REST API + WebSocket
                     ▼
┌─────────────────────────────────────────────────────────┐
│              Go API / 编排服务器 (:8080)                   │
│                                                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ 景点/路线 │ │ 会话分析  │ │ 会话编排  │ │ WebSocket │  │
│  │ CRUD API │ │ Dashboard│ │ 管道驱动  │ │   Hub     │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └───────────┘  │
│       │            │            │                        │
│       └────── scenicverse.db (SQLite) ────────────────── │
└────────────────────┬────────────────────────────────────┘
                     │  gRPC (Protocol Buffers)
                     ▼
┌─────────────────────────────────────────────────────────┐
│           Python 推理服务 (:50051)                        │
│                                                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │PersonaAgt│ │ Qwen-Omni│ │ Qwen-VL  │ │  Avatar   │  │
│  │ 工具+RAG │ │ 语音对话  │ │ 拍照识景  │ │ FlashHead │  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 为什么选择 Go + Python 双引擎？

| 维度 | Go | Python |
|------|-----|--------|
| **擅长领域** | 高并发网络、WebSocket、WebRTC、API 服务 | AI/ML 模型推理、PyTorch、Transformers |
| **本项目职责** | 会话编排、REST API、前端托管、RTP 传输 | LLM/TTS/ASR/Avatar 推理、工具调用、RAG |
| **通信方式** | gRPC 客户端 | gRPC 服务端 |

Go 的 goroutine 天然适合管理数百个并发 WebSocket 连接和 WebRTC 会话；Python 生态拥有完整的 AI 推理工具链。两端通过 gRPC Protocol Buffers 实现类型安全的跨语言通信，延迟开销 < 1ms。

### 三个进程，各司其职

| 进程 | 端口 | 技术栈 | 职责 |
|------|------|--------|------|
| **Go API Server** | 8080 | Go 1.25 + SQLite | REST API、WebSocket、会话编排、前端静态托管 |
| **Python Inference** | 50051 | Python 3.10 + gRPC | LLM/TTS/ASR/Avatar 推理、PersonaAgent 工具调用 |
| **Frontend** | 5173 (dev) | Vue 3 + Vite + TypeScript | 游客端景区入口 + 管理端后台 |

> 生产环境下，Go 服务器直接托管 `frontend/dist/` 静态文件，无需单独部署 Nginx 或 Vite。

---

## 功能全景

### 游客端能力

| 功能 | 实现方式 | 说明 |
|------|---------|------|
| 实时语音对话 | Qwen-Omni 端到端 | 支持语音打断、连续对话 |
| 文字对话 | 混合输入 | 语音和文字可在同一轮对话中混用 |
| 景区知识问答 | RAG 检索 + LLM | 灌入景区文档，准确回答门票、交通等问题 |
| 路线推荐 | PersonaAgent 工具调用 | 数字人自动调用 `get_routes` 推荐路线 |
| 景点查询 | PersonaAgent 工具调用 | 数字人自动调用 `get_attractions` 列出景点 |
| 拍照识景 | Qwen-VL 多模态 | 拍照识别景点，返回详细介绍 |
| 后台任务 | SubAgent 异步执行 | 长任务不阻塞对话 |

### 管理端能力

| 功能 | 页面路径 | 说明 |
|------|---------|------|
| 数据大屏 | `/admin/dashboard` | 今日服务人次、热门问题、情感分布、活跃时段 |
| 角色管理 | `/admin/characters` | 创建/编辑导游数字人，配置景区人设 |
| 景点管理 | `/admin/attractions` | 景点 CRUD，支持分类筛选 |
| 路线管理 | `/admin/routes` | 路线 CRUD，关联途经景点 |
| 对话日志 | `/admin/sessions` | 历史会话记录，支持按角色/时间筛选 |
| 分析报告 | `/admin/reports` | SubAgent 生成的游客感受度报告 |
| 系统设置 | `/admin/settings` | LLM/TTS/Avatar 等全局配置 |

---

## 快速开始

### 环境要求

- Node 18+
- Go 1.25（需要 `protoc-gen-go`、`protoc-gen-go-grpc`）
- Python 3.10+（推荐 Conda 管理）
- FFmpeg
- libopus-dev、libopusfile-dev、libsoxr-dev、pkg-config

> 纯语音模式不需要本地 GPU。数字人视频需要 CUDA 12.8+ GPU。

```bash
# 验证环境
node --version
go version
python --version
ffmpeg -version
```

### Step 1：克隆项目

```bash
git clone <your-repo-url>
cd CyberVerseSecondDev-master
```

### Step 2：创建 Python 环境

```bash
conda create -n cyberverse python=3.10
conda activate cyberverse
```

### Step 3：配置环境变量

```bash
cp infra/.env.example .env
```

编辑 `.env`，填入 API Key：

```env
# 阿里云通义千问系列
DASHSCOPE_API_KEY=your_dashscope_api_key

# 或火山引擎豆包系列
DOUBAO_ACCESS_TOKEN=your_doubao_access_token
DOUBAO_APP_ID=your_doubao_app_id
```

### Step 4：创建本地配置

```bash
cp infra/cyberverse_config.example.yaml cyberverse_config.yaml
```

纯语音模式（无需 GPU）：

```yaml
inference:
  avatar:
    enabled: false
```

### Step 5：安装依赖

```bash
make setup
pip install -e ".[all]"
```

### Step 6：导入 Demo 数据

```bash
# 导入济南景区 Demo 数据（10 个景点 + 4 条路线）
sqlite3 data/scenicverse.db < docs/second-development-plan/seed_jinan_demo.sql
```

### Step 7：启动服务（3 个终端）

```bash
# 终端 1 — Python 推理服务
conda activate cyberverse
make inference

# 终端 2 — Go API 服务器
make server

# 终端 3 — 前端开发服务器（仅开发时需要）
make frontend
```

### Step 8：验证

```bash
# 健康检查
curl -s http://localhost:8080/api/v1/health
```

| 页面 | 地址 | 说明 |
|------|------|------|
| 游客端 | http://localhost:5173/scenic | 景区入口页 |
| 管理端 | http://localhost:5173/admin/dashboard | 数据大屏 |
| API | http://localhost:8080/api/v1/health | 健康检查 |

---

## 数字人视频（可选）

如需启用实时数字人视频能力，需要 GPU 环境。

### 额外要求

- CUDA 12.8+ GPU
- PyTorch 2.8 (CUDA 12.8)
- FFmpeg with libvpx

```bash
pip3 install torch==2.8.0 torchvision==0.23.0 torchaudio==2.8.0 --index-url https://download.pytorch.org/whl/cu128
```

### 下载模型权重

#### FlashHead

```bash
hf download Soul-AILab/SoulX-FlashHead-1_3B --local-dir ./checkpoints/SoulX-FlashHead-1_3B
hf download facebook/wav2vec2-base-960h --local-dir ./checkpoints/wav2vec2-base-960h
```

#### LiveAct

```bash
hf download Soul-AILab/LiveAct --local-dir ./checkpoints/LiveAct
hf download TencentGameMate/chinese-wav2vec2-base --local-dir ./checkpoints/chinese-wav2vec2-base
```

### 配置 Avatar

在 `config.yaml` 中设置：

```yaml
inference:
  avatar:
    enabled: true
    default: "flash_head"   # 或 "live_act"
```

### 硬件基准

| 模型 | 质量 | GPU | 分辨率 | FPS | 实时? |
|------|------|-----|--------|-----|-------|
| FlashHead 1.3B | Pro | RTX 5090 ×2 | 512×512 | 25+ | ✅ |
| FlashHead 1.3B | Lite | RTX 4090 | 512×512 | 25+ | ✅ |
| LiveAct 18B | — | RTX PRO 6000 ×2 | 320×480 | 20 | ✅ |

---

## 项目结构

```
├── frontend/                     # Vue 3 前端
│   └── src/
│       ├── pages/
│       │   ├── ScenicLandingPage.vue    # 游客端 - 景区入口页
│       │   ├── SessionPage.vue          # 游客端 - 对话页
│       │   ├── DashboardPage.vue        # 管理端 - 数据大屏
│       │   ├── AttractionManagePage.vue # 管理端 - 景点管理
│       │   ├── RouteManagePage.vue      # 管理端 - 路线管理
│       │   ├── SessionLogPage.vue       # 管理端 - 对话日志
│       │   ├── ReportPage.vue           # 管理端 - 分析报告
│       │   └── ...
│       ├── components/
│       │   ├── ScenicCard.vue           # 景点/路线卡片组件
│       │   ├── ChatPanel.vue            # 聊天面板（双主题）
│       │   └── ...
│       ├── layouts/
│       │   └── AdminLayout.vue          # 管理端布局（侧边栏）
│       ├── router/index.ts              # 路由（/scenic + /admin/*）
│       └── i18n/messages.ts             # 中英文国际化
│
├── server/                       # Go API 服务器
│   ├── cmd/cyberverse-server/main.go    # 入口
│   └── internal/
│       ├── api/
│       │   ├── router.go                # 路由注册（+17 个新端点）
│       │   ├── scenic_handler.go        # 景点/路线 API handler
│       │   └── analytics_handler.go     # 会话分析/报告 API handler
│       ├── scenic/                      # 景区数据层
│       │   ├── db.go                    # db 初始化 + 迁移
│       │   ├── attraction.go            # 景点 CRUD Store
│       │   ├── route.go                 # 路线 CRUD Store
│       │   ├── session.go               # 会话 + 对话日志 Store
│       │   ├── dashboard.go             # Dashboard 聚合查询
│       │   └── report.go                # 报告 Store
│       ├── orchestrator/                # 会话编排引擎
│       │   ├── orchestrator.go          # 管道驱动 + DB 双写 + 景区上下文注入
│       │   └── session.go               # 会话状态机
│       └── character/
│           └── store.go                 # 角色数据（含景区扩展字段）
│
├── inference/                    # Python 推理服务
│   ├── server.py                        # 入口 + 插件注册
│   ├── plugins/
│   │   ├── voice_llm/
│   │   │   └── persona_agent.py         # PersonaAgent（工具调用 + RAG）
│   │   ├── avatar/                      # 数字人视频插件
│   │   ├── llm/                         # LLM 插件
│   │   ├── tts/                         # TTS 插件
│   │   └── asr/                         # ASR 插件
│   ├── services/                        # gRPC 服务层
│   └── rag/                             # RAG 检索引擎
│
├── proto/                        # gRPC Protocol Buffers 定义
├── data/
│   ├── scenicverse.db                    # 统一数据库（会话+景点+路线）
│   ├── characters/<id>/                 # 角色数据 + 知识库
│   └── tasks/                           # SubAgent 任务数据
│
├── docs/second-development-plan/       # 设计文档
│   ├── spec-1-system-level-business-spec.md  # 系统级业务 Spec
│   ├── task.md                               # 任务清单
│   ├── progress.md                           # 进度记录
│   ├── seed_jinan_demo.sql                   # 济南景区 Demo 数据
│   └── server-deployment-guide.md            # 部署指南
│
├── cyberverse_config.yaml               # 运行时配置
├── Makefile                             # 构建命令
└── CLAUDE.md                            # 开发指南
```

---

## 数据库设计

项目使用 `scenicverse.db`（SQLite）统一管理景区业务数据：

```sql
-- 景点
CREATE TABLE attractions (
    id TEXT PRIMARY KEY, name TEXT, description TEXT,
    category TEXT, location TEXT, tags TEXT, ...
);

-- 路线
CREATE TABLE routes (
    id TEXT PRIMARY KEY, name TEXT, description TEXT,
    duration TEXT, difficulty TEXT, ...
);

-- 路线途经点
CREATE TABLE route_steps (
    route_id TEXT, attraction_id TEXT, order_num INTEGER,
    highlight TEXT, stay_minutes INTEGER
);

-- 对话会话
CREATE TABLE sessions (
    id TEXT PRIMARY KEY, character_id TEXT,
    started_at TEXT, ended_at TEXT, duration_s INTEGER,
    turn_count INTEGER, sentiment TEXT
);

-- 对话日志
CREATE TABLE conversation_logs (
    session_id TEXT, turn_seq INTEGER,
    role TEXT, content TEXT, content_type TEXT, created_at TEXT
);

-- 分析报告
CREATE TABLE reports (
    id TEXT PRIMARY KEY, character_id TEXT,
    status TEXT, content_json TEXT, ...
);
```

---

## API 总览

### 基础 API

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| CRUD | `/api/v1/characters` | 角色管理 |
| CRUD | `/api/v1/characters/:id/knowledge` | 知识库管理 |
| GET/POST | `/api/v1/settings` | 全局配置 |
| WS | `/ws` | WebSocket 实时通信 |

### 景区业务 API

| 方法 | 路径 | 功能 |
|------|------|------|
| CRUD | `/api/v1/attractions` | 景点管理 |
| CRUD | `/api/v1/routes` | 路线管理 |
| GET | `/api/v1/character-routes` | 角色推荐路线（供工具调用） |
| GET | `/api/v1/analytics/dashboard` | 数据大屏聚合 |
| GET | `/api/v1/analytics/sessions` | 会话列表 |
| GET | `/api/v1/analytics/sessions/:id` | 会话详情 |
| GET | `/api/v1/analytics/sentiment` | 情感趋势 |
| POST | `/api/v1/reports/generate` | 触发报告生成 |
| GET | `/api/v1/reports` | 报告列表 |

---

## 核心流程

### 游客语音问答全链路

```
游客语音输入
    │
    ▼
[WebRTC] 音频传输 → Go 会话编排引擎
    │
    ▼
[gRPC] 音频流 → Python Qwen-Omni（端到端语音理解）
    │
    ▼
PersonaAgent（LLM 推理 + 工具调用决策）
    ├─ 调用 get_attractions → Go API → SQLite → 返回景点列表
    ├─ 调用 get_routes     → Go API → SQLite → 返回路线推荐
    ├─ 调用 RAG 检索       → Chroma 向量库 → 返回知识片段
    └─ 生成自然语言回答
    │
    ▼
[gRPC] 文本 → Python Qwen-TTS → 音频流
    │
    ▼
[WebRTC] 音频推送 → 游客浏览器播放
```

### 管理端运营闭环

```
管理员配置景点/路线/角色 → 数据写入 scenicverse.db
    │
    ▼
游客对话 → 编排引擎注入景区上下文 → PersonaAgent 工具调用
    │
    ▼
对话日志双写（文件 + DB）→ 会话结束时情感标注
    │
    ▼
数据大屏聚合展示 ← scenicverse.db 查询
    │
    ▼
SubAgent 生成感受度报告 → 报告页面查看
```

---

## 开发命令

```bash
make setup              # 安装依赖 + 生成 proto + npm install
make proto              # 重新生成 gRPC 桩代码
make inference          # 启动 Python 推理服务
make server             # 启动 Go API 服务器
make frontend           # 启动 Vite 开发服务器
make build              # 构建 Go + 前端
make test               # 运行测试
make clean              # 清理生成文件
```

---

## 技术栈

| 层级 | 技术 |
|------|------|
| **前端** | Vue 3 + Vite + TypeScript + ECharts |
| **后端** | Go 1.25（net/http + goroutine 并发） |
| **推理服务** | Python 3.10 + asyncio + gRPC |
| **通信协议** | gRPC (Protocol Buffers) + WebSocket + WebRTC |
| **数据库** | SQLite（scenicverse.db） |
| **向量检索** | Chroma + LangChain Embeddings |
| **语音模型** | Qwen-Omni（端到端语音） |
| **视觉模型** | Qwen-VL（多模态识别） |
| **数字人模型** | SoulX-FlashHead / SoulX-LiveAct |
| **任务编排** | LangGraph + SQLite Checkpointer |

---

## 指标

| 指标 | 要求 | 实现 |
|------|------|------|
| 多模态大模型 | ≥ 1 个 | Qwen-Omni（语音）+ Qwen-VL（视觉） |
| 问答准确率 | ≥ 90% | FAQ + RAG + prompt 约束 |
| 语音问答延迟 | ≤ 5 秒 | Qwen-Omni 端到端 ~2-3 秒 |
| 口型同步 | 专家评估 | FlashHead / LiveAct 实时驱动 |

---

## License

GNU General Public License v3.0 — 详见 [LICENSE](LICENSE)。

## 致谢

- [SoulX-FlashHead](https://github.com/Soul-AILab/SoulX-FlashHead) — 数字人基座模型（Soul AI Lab）
- [SoulX-LiveAct](https://github.com/Soul-AILab/SoulX-LiveAct) — 数字人基座模型（Soul AI Lab）
- [Pion](https://github.com/pion/webrtc) — Go WebRTC 实现
