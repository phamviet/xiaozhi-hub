package xiaozhi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// serverBaseConfig /xiaozhi/ota
func (m *Manager) otaConfig(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, map[string]bool{"ok": true})
}
