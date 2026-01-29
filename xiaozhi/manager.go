package xiaozhi

import (
	"net/http"
	"strings"

	"github.com/phamviet/xiaozhi-hub/internal/hub"
	"github.com/phamviet/xiaozhi-hub/xiaozhi/store"
	"github.com/pocketbase/pocketbase/core"
)

var _ hub.Plugin = (*Manager)(nil)

type Manager struct {
	App   core.App
	Store *store.Manager
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Name() string {
	return "xiaozhi"
}

func (m *Manager) Initialize(hub *hub.Hub) error {
	m.App = hub.App
	hub.App.OnServe().BindFunc(func(e *core.ServeEvent) error {
		m.Store = store.NewManager(hub.App)
		if err := m.registerAuthRoutes(e); err != nil {
			return err
		}

		return e.Next()
	})

	return nil
}

func (m *Manager) registerAuthRoutes(se *core.ServeEvent) error {
	xiaozhi := se.Router.Group("/xiaozhi")
	xiaozhi.POST("/ota", m.otaRequest)
	xiaozhi.POST("/ota/activate", m.otaActivateRequest)
	xiaozhi.POST("/ota/bind-device", m.otaBindDeviceRequest)

	// Ensure ending slash is supported
	xiaozhi.POST("/ota/", m.otaRequest)
	xiaozhi.POST("/ota/activate/", m.otaActivateRequest)

	// Auth with manager secret
	apiAuth := xiaozhi.Group("")
	apiAuth.BindFunc(m.requireAuth)

	// Server base config
	apiAuth.POST("/config/server-base", m.serverBaseConfig)

	// Get agent models
	apiAuth.POST("/config/agent-models", m.getAgentModels)

	// report chat
	apiAuth.POST("/agent/chat-history/report", m.reportChat)

	// Emit chat summary
	apiAuth.POST("/agent/chat-summary/{sessionId}/save", m.summaryChat)

	return nil
}

func (m *Manager) requireAuth(e *core.RequestEvent) error {
	authHeader := e.Request.Header.Get("Authorization")
	if authHeader == "" {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing Authorization header"})
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Authorization header format"})
	}
	token := parts[1]

	secret, err := m.Store.GetSysParam("server.secret")
	if err != nil {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Server secret not configured"})
	}

	if secret == "" || token != secret {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	return e.Next()
}
