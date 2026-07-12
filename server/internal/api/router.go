package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyberverse/server/internal/agenttask"
	"github.com/cyberverse/server/internal/character"
	"github.com/cyberverse/server/internal/config"
	"github.com/cyberverse/server/internal/livekit"
	"github.com/cyberverse/server/internal/orchestrator"
	ragstore "github.com/cyberverse/server/internal/rag"
	"github.com/cyberverse/server/internal/scenic"
	"github.com/cyberverse/server/internal/ws"
)

type Router struct {
	sessionMgr *orchestrator.SessionManager
	orch       *orchestrator.Orchestrator
	wsHub      *ws.Hub
	roomMgr    *livekit.RoomManager
	taskSvc    *agenttask.Service
	cfg        *config.Config
	charStore  *character.Store
	ragStore   *ragstore.Store
	scenicDB   *scenic.DB
	envPath    string
	configPath string
	modelsDir  string
	mux        *http.ServeMux
}

func NewRouter(
	sessionMgr *orchestrator.SessionManager,
	orch *orchestrator.Orchestrator,
	wsHub *ws.Hub,
	roomMgr *livekit.RoomManager,
	cfg *config.Config,
	charStore *character.Store,
	envPath string,
	configPath string,
	scenicDB *scenic.DB,
	taskServices ...*agenttask.Service,
) *Router {
	r := &Router{
		sessionMgr: sessionMgr,
		orch:       orch,
		wsHub:      wsHub,
		roomMgr:    roomMgr,
		cfg:        cfg,
		charStore:  charStore,
		ragStore:   ragstore.NewStore(charStore),
		scenicDB:   scenicDB,
		envPath:    envPath,
		configPath: configPath,
		modelsDir:  filepath.Join(filepath.Dir(configPath), "models"),
		mux:        http.NewServeMux(),
	}
	if len(taskServices) > 0 {
		r.taskSvc = taskServices[0]
	}
	r.registerRoutes()
	return r
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("GET /api/v1/health", r.handleHealth)
	r.mux.HandleFunc("GET /api/v1/components", r.handleListComponents)
	r.mux.HandleFunc("POST /api/v1/sessions", r.handleCreateSession)
	r.mux.HandleFunc("DELETE /api/v1/sessions/{id}", r.handleDeleteSession)
	r.mux.HandleFunc("POST /api/v1/sessions/{id}/message", r.handleSendMessage)
	r.mux.HandleFunc("GET /api/v1/sessions/{id}/tasks", r.handleListSessionTasks)
	r.mux.HandleFunc("GET /api/v1/sessions", r.handleListSessions)
	r.mux.HandleFunc("GET /api/v1/tasks/{task_id}", r.handleGetTask)
	r.mux.HandleFunc("GET /api/v1/tasks/{task_id}/events", r.handleListTaskEvents)
	r.mux.HandleFunc("GET /api/v1/tasks/{task_id}/artifacts/{artifact_id}", r.handleGetTaskArtifact)
	r.mux.HandleFunc("POST /api/v1/internal/tasks/{task_id}/events", r.handleInternalTaskEvent)
	r.mux.HandleFunc("POST /api/v1/internal/tasks/{task_id}/artifacts", r.handleInternalTaskArtifact)
	r.mux.HandleFunc("POST /api/v1/internal/characters/{id}/knowledge/search", r.handleInternalKnowledgeSearch)
	r.mux.HandleFunc("GET /ws/chat/{id}", r.handleWebSocket)

	// Character CRUD
	r.mux.HandleFunc("GET /api/v1/characters", r.handleListCharacters)
	r.mux.HandleFunc("POST /api/v1/characters", r.handleCreateCharacter)
	r.mux.HandleFunc("POST /api/v1/characters/test-voice", r.handleTestCharacterVoice)
	r.mux.HandleFunc("GET /api/v1/characters/{id}", r.handleGetCharacter)
	r.mux.HandleFunc("PUT /api/v1/characters/{id}", r.handleUpdateCharacter)
	r.mux.HandleFunc("DELETE /api/v1/characters/{id}", r.handleDeleteCharacter)
	r.mux.HandleFunc("POST /api/v1/characters/{id}/avatar", r.handleUploadAvatar)
	r.mux.HandleFunc("GET /api/v1/characters/{id}/images", r.handleListImages)
	r.mux.HandleFunc("GET /api/v1/characters/{id}/images/{filename}", r.handleGetCharacterImage)
	r.mux.HandleFunc("GET /api/v1/characters/{id}/knowledge", r.handleListKnowledgeSources)
	r.mux.HandleFunc("POST /api/v1/characters/{id}/knowledge/files", r.handleUploadKnowledgeFiles)
	r.mux.HandleFunc("DELETE /api/v1/characters/{id}/knowledge/{source_id}", r.handleDeleteKnowledgeSource)
	r.mux.HandleFunc("POST /api/v1/characters/{id}/knowledge/{source_id}/reindex", r.handleReindexKnowledgeSource)
	r.mux.HandleFunc("GET /api/v1/characters/{id}/idle-videos/{imgbase}/{variant}/{filename}", r.handleGetIdleVideo)
	r.mux.HandleFunc("GET /api/v1/characters/{id}/idle-videos/{imgbase}/{filename}", r.handleGetIdleVideo)
	r.mux.HandleFunc("DELETE /api/v1/characters/{id}/images/{filename}", r.handleDeleteImage)
	r.mux.HandleFunc("PUT /api/v1/characters/{id}/images/{filename}/activate", r.handleActivateImage)
	r.mux.HandleFunc("GET /api/v1/avatars/{filename}", r.handleGetAvatar)

	// Conversation history
	r.mux.HandleFunc("GET /api/v1/characters/{id}/conversations/messages", r.handleGetConversationMessages)

	// Settings
	r.mux.HandleFunc("GET /api/v1/settings", r.handleGetSettings)
	r.mux.HandleFunc("PUT /api/v1/settings", r.handleUpdateSettings)
	r.mux.HandleFunc("POST /api/v1/settings/test", r.handleTestConnection)

	// Launch config
	r.mux.HandleFunc("GET /api/v1/config/avatar-model", r.handleGetAvatarModelInfo)
	r.mux.HandleFunc("GET /api/v1/config/launch", r.handleGetLaunchConfig)
	r.mux.HandleFunc("PUT /api/v1/config/launch", r.handleUpdateLaunchConfig)

	// Scenic guide: Attractions
	r.mux.HandleFunc("GET /api/v1/attractions", r.handleListAttractions)
	r.mux.HandleFunc("POST /api/v1/attractions", r.handleCreateAttraction)
	r.mux.HandleFunc("GET /api/v1/attractions/{id}", r.handleGetAttraction)
	r.mux.HandleFunc("PUT /api/v1/attractions/{id}", r.handleUpdateAttraction)
	r.mux.HandleFunc("DELETE /api/v1/attractions/{id}", r.handleDeleteAttraction)

	// Scenic guide: Routes
	r.mux.HandleFunc("GET /api/v1/routes", r.handleListRoutes)
	r.mux.HandleFunc("POST /api/v1/routes", r.handleCreateRoute)
	r.mux.HandleFunc("GET /api/v1/routes/{id}", r.handleGetRoute)
	r.mux.HandleFunc("PUT /api/v1/routes/{id}", r.handleUpdateRoute)
	r.mux.HandleFunc("DELETE /api/v1/routes/{id}", r.handleDeleteRoute)

	// Scenic guide: Analytics
	r.mux.HandleFunc("GET /api/v1/analytics/dashboard", r.handleGetDashboard)
	r.mux.HandleFunc("GET /api/v1/analytics/sessions", r.handleListAnalyticsSessions)
	r.mux.HandleFunc("GET /api/v1/analytics/sessions/{id}", r.handleGetAnalyticsSessionDetail)

	// Scenic guide: Reports
	r.mux.HandleFunc("GET /api/v1/reports", r.handleListReports)
	r.mux.HandleFunc("POST /api/v1/reports/generate", r.handleGenerateReport)
	r.mux.HandleFunc("GET /api/v1/reports/{id}", r.handleGetReport)
	r.mux.HandleFunc("DELETE /api/v1/reports/{id}", r.handleDeleteReport)
}

func (r *Router) Handler() http.Handler {
	return corsMiddleware(r.spaFallback(r.mux))
}

// spaFallback wraps the API handler with a static file server for the Vue SPA.
// Non-API requests: serve file from frontend/dist/ if exists, else index.html.
func (r *Router) spaFallback(api http.Handler) http.Handler {
	staticDir := filepath.Join("..", "frontend", "dist")
	indexFile := filepath.Join(staticDir, "index.html")

	// If frontend/dist doesn't exist (e.g. server-only deployment), skip SPA serving.
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return api
	}

	fs := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// API and WebSocket routes go directly to the mux.
		if strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/ws/") {
			api.ServeHTTP(w, req)
			return
		}
		// Try to serve static file.
		path := filepath.Join(staticDir, filepath.FromSlash(req.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, req)
			return
		}
		// SPA fallback: serve index.html for all other routes.
		http.ServeFile(w, req, indexFile)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
