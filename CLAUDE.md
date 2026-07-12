# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Rules

1. All code comments must be written in English.
2. Frontend changes must consider internationalization and include both Chinese and English user-facing text where applicable.
3. When modifying code, follow the minimum-change principle and keep edits narrowly scoped.

## Common Commands

The Makefile is the source of truth. It auto-detects Go 1.25, Node 22 (via nvm), and a `cyberverse` conda env for native libs (opus/soxr/pkg-config) used by LiveKit media SDK.

```bash
make setup              # pip install -e ".[dev,inference]" + generate proto + npm install
make proto              # Regenerate Python + Go gRPC stubs from proto/*.proto
make inference          # Start Python gRPC inference server (reads cyberverse_config.yaml; auto-picks python vs torchrun)
make server             # Start Go API server with -tags livekit (loads .env, listens on :8080 + :50051 client + WebRTC)
make frontend           # Vite dev server on :5173 (CHOKIDAR_USEPOLLING=true)
make build              # build-go + frontend-build
make docker-up          # docker compose up --build (infra/docker-compose.yml)
make clean              # Remove generated proto files and bin/
```

Tests:

```bash
make test               # test-py + test-go
make test-py            # python -m pytest tests/unit -v
make test-go            # cd server && go test ./... -v  (must run via make for PKG_CONFIG_PATH/LD_LIBRARY_PATH)
make test-integration   # pytest tests/integration -m integration -v -s  (requires GPU + checkpoints)

# Single tests
python -m pytest tests/unit/path/to/test_file.py::TestClass::test_method -v
cd server && go test ./internal/orchestrator -run TestNameRegex -v
```

Optional Python extras: `pip install -e ".[all]"` (or pick from `[llm,tts,omni,asr,rag,flash_head,live_act,agent,voice_llm,qwen]`). The `agent` extra (langgraph + sqlite checkpointer) is installed separately via `make setup-agent`.

## High-Level Architecture

Three processes cooperate at runtime: a **Vue frontend** (port 5173 dev), a **Go API/orchestration server** (HTTP :8080, optional WebRTC :8443), and a **Python inference server** (gRPC :50051). They communicate via gRPC defined in `proto/` and a Go-frontend WebSocket hub.

```
frontend (Vue 3 + Vite)
   │  REST + WS
   ▼
server/cmd/cyberverse-server (Go)         <- .env + cyberverse_config.yaml
   |  ├── api/        REST routes (characters, conversations, knowledge, settings, tasks)
   |  ├── ws/         WebSocket hub fanning events to clients
   |  ├── orchestrator/  Session manager — drives one conversation: voice in -> omni/persona -> TTS+avatar out
   |  ├── direct/     P2P WebRTC + embedded TURN (streaming_mode: "direct")
   |  ├── livekit/    LiveKit SFU bot/room manager (streaming_mode: "livekit", build tag `livekit`)
   |  ├── mediapeer/  RTP segmentation/VP8 chunking shared by both streaming modes
   |  ├── agenttask/  SQLite-backed projection store + service for SubAgent task lifecycle
   |  ├── inference/  gRPC client to Python server
   |  ├── character/  Per-character on-disk store (data/characters/<id>/)
   |  ├── rag/        RAG retrieval client glue
   |  └── recording/  Optional video recording
   │  gRPC
   ▼
inference/server.py (Python asyncio)
   ├── core/registry.py   Plugin registry — instantiates plugin_class strings from YAML
   ├── plugins/
   │   ├── avatar/        flash_head, live_act (only one initialized; selected by inference.avatar.default)
   │   ├── voice_llm/     qwen_omni_realtime, doubao_realtime, persona_agent (PersonaAgent wraps a provider)
   │   ├── llm/, tts/, asr/
   │   └── qwen_endpoint.py
   ├── services/          One *_service.py per gRPC service (avatar, llm, tts, asr, rag, voice_llm)
   ├── rag/               Embeddings + vector store (chroma via langchain)
   └── generated/         protoc output (do not edit; regenerate with `make proto`)
```

### Plugin pattern

Every backend (avatar, omni, persona, LLM, TTS, ASR) is a plugin loaded by FQCN from `cyberverse_config.yaml`. The registry resolves `plugin_class: "inference.plugins.foo.bar.BarPlugin"` and constructs it with the YAML config block. Avatar plugins are special: only `inference.avatar.default` is initialized in the inference process to keep one model resident on the GPU. Other categories (`llm`, `tts`, `asr`, `omni`, `persona`, `voice_llm`) initialize all configured providers eagerly — see `_INITIALIZE_ALL_CATEGORIES` in `inference/server.py`.

Adding a new provider: implement against the matching base class in `inference/plugins/<category>/base.py`, register it in `cyberverse_config.yaml` under the right category, and (if it exposes new RPCs) update the relevant `proto/*.proto` then run `make proto`.

### PersonaAgent + SubAgent

`persona_agent` is an orchestration plugin that wraps a real omni provider (currently `qwen_omni`) and exposes hidden tool calls for long-running background work. The Go server's `agenttask` package projects task state into SQLite (`data/tasks/tasks.db`) and broadcasts updates via the WS hub; artifacts land in `data/tasks/artifacts/`. PersonaAgent state is checkpointed by langgraph in `data/tasks/langgraph_checkpoints.db`.

### Two streaming modes

`pipeline.streaming_mode` selects between `direct` (Pion-based P2P + embedded TURN on `:8443/TCP`, code in `server/internal/direct/`) and `livekit` (SFU mode, code in `server/internal/livekit/`, gated by Go build tag `livekit`). Both modes share `mediapeer/` for RTP segmentation. The `make server` target always builds with `-tags livekit`.

### Proto-generated code

`inference/generated/*_pb2*.py` and `server/internal/pb/*.go` are checked-in but treated as build artifacts — regenerate via `make proto` (runs `scripts/generate_proto.sh`) after editing any `proto/*.proto`. `make clean` removes both.

### Configuration loading

`cyberverse_config.yaml` (copied from `infra/cyberverse_config.example.yaml`) is the single source of runtime config for both processes. The Go server first loads `.env` from the same directory, then expands `${VAR}` placeholders in YAML. The web UI at `/settings` writes back into the same YAML/data tree, so don't assume YAML values are static across a session.

### Key runtime ports

| Port | Process | Purpose |
|------|---------|---------|
| 5173 | frontend | Vite dev server |
| 8080 | server | REST + WebSocket |
| 50051 | inference | gRPC (server is client) |
| 8443/TCP | server | Embedded TURN for `streaming_mode: direct` (must be reachable from the browser) |

Health check: `curl -s http://localhost:8080/api/v1/health`.

## Key Files

- `cyberverse_config.yaml` — Runtime config (copy from `infra/cyberverse_config.example.yaml`)
- `.env` — API keys and secrets (copy from `infra/.env.example`)
- `proto/*.proto` — gRPC service definitions
- `inference/server.py` — Python inference server entry point
- `server/cmd/cyberverse-server/main.go` — Go API server entry point
- `frontend/src/main.ts` — Vue frontend entry point
- `Makefile` — Build and development commands

## Conventions

- **Avatar inference logs** — when changing avatar pipelines, watch RTP (`elapsed / (frames / fps)`) in `make inference` output. RTP > 1 means the model can't keep up with playback; the README's "QA — Self-Check" section is the troubleshooting flow.
- **Go build tags** — production and `make server` always use `-tags livekit`. Don't add code paths that only compile without the tag.
- **Native deps** — `make server` and `make test-go` inject `PKG_CONFIG_PATH`/`LD_LIBRARY_PATH` from the conda env. Running `go test`/`go run` directly will fail to link opus/soxr; always go through the Makefile.
- **Frontend i18n** — user-facing strings must be added to both Chinese and English locale files under `frontend/src/i18n/`.

## Truth-First Reasoning Rules

Verdict before agreement. Treat every claim, diagnosis, or plan as unverified until checked against code, docs, or logic. When the user makes a technical claim, lead with one of: `Correct` / `Incorrect` / `Partially correct` / `Unknown` / `Bad approach` / `Better approach available`, then explain why and propose the next concrete step. Skip the format when a direct answer is simpler.

Disagree directly when the user is wrong — "No. The issue is X" beats "You're right, but…". Do not implement bad instructions silently; flag the flaw, propose the better approach, and wait. Do not invent facts: say "unknown" when verification is needed. Reject fixes that patch symptoms instead of root causes, and reject rewrites that damage architecture, security, performance, or type safety even if requested. Prefer minimal correct fixes.

Forbidden: agreeing without verification, flattering, hiding disagreement, treating assumptions as fact, pretending certainty when evidence is weak. Tone is calm and firm, not rude. The goal is to prevent incorrect thinking and weak execution, not to argue.

## Additional Guidelines

See `AGENTS.md` for detailed behavioral guidelines on code review, planning, and factual accuracy. Key principles:
- Challenge weak assumptions before implementation
- Inspect actual code paths before accepting explanations
- Prefer minimal correct fixes over large rewrites
- Say "unknown" when verification is needed
