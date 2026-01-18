package xiaozhi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

type DeviceBindRequest struct {
	Code    string `json:"code"`
	AgentID string `json:"agentId"`
}

func (m *Manager) deviceBind(e *core.RequestEvent) error {
	var req DeviceBindRequest
	if err := e.BindBody(&req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Code == "" || req.AgentID == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "code and agentId are required"})
	}

	// Validate code
	info, ok := m.BindingManager.VerifyCode(req.Code)
	if !ok {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired device code"})
	}

	// Check if agent exists
	agent, err := m.getAgentByID(req.AgentID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}

	// Check if device already exists
	existingDevice, _ := m.getDeviceByMacAddress(info.MacAddress)
	if existingDevice != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Device already bound"})
	}

	// Create new device
	collection, err := e.App.FindCollectionByNameOrId("ai_device")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "ai_device collection not found"})
	}

	record := core.NewRecord(collection)
	record.Set("mac_address", info.MacAddress)
	record.Set("agent", req.AgentID)
	record.Set("user", agent.UserID)

	if err := e.App.Save(record); err != nil {
		e.App.Logger().Error("Failed to save new device binding", "error", err, "mac", info.MacAddress)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save device binding"})
	}

	// Remove from binding manager
	m.BindingManager.RemoveBinding(info.MacAddress, info.ClientId)

	return e.JSON(http.StatusOK, map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": map[string]string{
			"macAddress": info.MacAddress,
			"agentId":    req.AgentID,
		},
	})
}
