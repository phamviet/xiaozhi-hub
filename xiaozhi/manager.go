package xiaozhi

import (
	"github.com/phamviet/xiaozhi-hub/internal/hub"
	"github.com/pocketbase/pocketbase/core"
)

var _ hub.Plugin = (*Manager)(nil)

type Manager struct {
	App            core.App
	BindingManager *DeviceBindingManager
}

func NewManager() *Manager {
	return &Manager{
		BindingManager: NewDeviceBindingManager(),
	}
}

func (m *Manager) Name() string {
	return "xiaozhi"
}

func (m *Manager) Initialize(hub *hub.Hub) error {
	m.App = hub.App
	hub.App.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := m.registerAuthRoutes(e); err != nil {
			return err
		}

		return e.Next()
	})

	return nil
}

func (m *Manager) registerAuthRoutes(se *core.ServeEvent) error {
	hubAPI := se.Router.Group("/hub/api")
	xiaozhi := se.Router.Group("/xiaozhi")
	xiaozhi.POST("/ota", m.otaRequest)

	// Ensure ending slash is supported
	xiaozhi.POST("/ota/", m.otaRequest)

	// Auth with manager secret
	apiAuth := xiaozhi.Group("")

	// Server base config
	apiAuth.POST("/config/server-base", m.serverBaseConfig)

	// Get agent models
	apiAuth.POST("/config/agent-models", m.agentModelsConfig)

	// report chat
	apiAuth.POST("/agent/chat-history/report", m.reportChat)

	// Emit chat summary
	apiAuth.POST("/agent/chat-summary/{sessionId}/save", m.summaryChat)

	// Device binding (manually from Hub UI)
	hubAPI.POST("/device/bind", m.deviceBind)

	return nil
}
