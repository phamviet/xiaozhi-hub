package xiaozhi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// reportChat /xiaozhi/agent/chat-history/report
func (m *Manager) reportChat(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, map[string]bool{"ok": true})
}
