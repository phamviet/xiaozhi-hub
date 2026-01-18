package xiaozhi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// summaryChat /xiaozhi/agent/chat-summary/{sessionId}/save
func (m *Manager) summaryChat(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, map[string]bool{"ok": true})
}
