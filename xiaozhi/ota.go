package xiaozhi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/pocketbase/core"
)

type OTARequest struct {
	Version     int    `json:"version"`
	UUID        string `json:"uuid"`
	Application struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		CompileTime string `json:"compile_time"`
	} `json:"application"`
	Board struct {
		Type string `json:"type"`
		SSID string `json:"ssid"`
		IP   string `json:"ip"`
		MAC  string `json:"mac"`
	} `json:"board"`
	MacAddress string `json:"mac_address"`
}

type OTAResponse struct {
	ServerTime struct {
		Timestamp      int64  `json:"timestamp"`
		TimeZone       string `json:"timeZone"`
		TimezoneOffset int    `json:"timezone_offset"`
	} `json:"server_time"`
	Firmware struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	} `json:"firmware"`
	Websocket struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	} `json:"websocket"`
	Activation struct {
		Code      string `json:"code,omitempty"`
		Challenge string `json:"challenge,omitempty"`
		Message   string `json:"message,omitempty"`
	} `json:"activation,omitempty"`
}

type ActivationRequest struct {
	Payload struct {
		Algorithm    string `json:"algorithm"`
		SerialNumber string `json:"serial_number"`
		Challenge    string `json:"challenge"`
		HMAC         string `json:"hmac"`
	} `json:"Payload"`
}

type ActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DeviceBindResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"msg"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// otaRequest /xiaozhi/ota
func (m *Manager) otaRequest(e *core.RequestEvent) error {
	deviceID := e.Request.Header.Get("device-id")
	if deviceID == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "device-id header is required"})
	}

	clientID := e.Request.Header.Get("client-id")

	// Validate MAC address format
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	if !macRegex.MatchString(deviceID) {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid device-id format"})
	}

	// Read raw body
	bodyBytes, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return err
	}
	bodyString := string(bodyBytes)

	var req OTARequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		e.App.Logger().Warn("unmarshal ota request", "body", bodyString, "error", err)
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	e.App.Logger().Info("ota request body", "body", bodyString, "deviceId", deviceID)
	device, err := m.Store.GetDeviceByMacAddress(deviceID)
	var bindCode, challenge string
	if err != nil {
		bindInfo, err := m.Store.CreateUnboundDevice(deviceID)
		if err != nil || bindInfo == nil {
			e.App.Logger().Error("Error", "err", err, "bindInfo", bindInfo)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		bindCode = bindInfo.BindCode
		challenge = bindInfo.Challenge
	} else {
		// If device exists but is not bound (no user_id), we might want to return the bind code again
		// But for now let's assume we only return it for new devices or if we want to re-bind.
		// Checking if device is bound:
		if device.UserId == "" {
			bindCode = device.BindCode
			challenge = device.Challenge
		}
	}

	// Store request body to ota_requests collection
	if err := m.Store.LogOTARequest(deviceID, req.Board.Type, bodyString); err != nil {
		e.App.Logger().Error("Failed to save ota_request", "error", err)
	}

	// Get secret from sys_params
	secret := ""
	if val, err := m.Store.GetSysParam("server.secret"); err == nil {
		secret = val
	}

	// Construct wsURL from current request host
	scheme := "ws"
	if e.Request.TLS != nil {
		scheme = "wss"
	}
	wsURL := fmt.Sprintf("%s://%s/api/v1", scheme, e.Request.Host)

	now := time.Now()
	timestamp := now.Unix()

	tokenString := ""
	if secret != "" {
		// token content: client-id + |device-id + |current_timestamp
		message := fmt.Sprintf("%s|%s|%d", clientID, deviceID, timestamp)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(message))
		signature := h.Sum(nil)
		signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

		// The final token format is: signature.timestamp
		tokenString = fmt.Sprintf("%s.%d", signatureBase64, timestamp)
	}

	response := OTAResponse{}
	response.ServerTime.Timestamp = now.UnixMilli()
	response.ServerTime.TimeZone = "Asia/Ho_Chi_Minh"
	response.ServerTime.TimezoneOffset = 420

	response.Firmware.Version = "1.0.0"
	response.Firmware.URL = "http://xiaozhi.server.com:8002/xiaozhi/otaMag/download/NOT_ACTIVATED_FIRMWARE_THIS_IS_A_INVALID_URL"

	response.Websocket.URL = wsURL
	response.Websocket.Token = tokenString

	if bindCode != "" {
		response.Activation.Code = bindCode
		response.Activation.Challenge = challenge
		response.Activation.Message = "Device not bound. Please enter the code to bind."
	}

	return e.JSON(http.StatusOK, response)
}

// otaActivateRequest /xiaozhi/ota/activate
func (m *Manager) otaActivateRequest(e *core.RequestEvent) error {
	macAddress := e.Request.Header.Get("device-id")
	if macAddress == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "device-id header is required"})
	}

	// Validate MAC address format
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	if !macRegex.MatchString(macAddress) {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid device-id format"})
	}

	// Read request body
	bodyBytes, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return err
	}

	e.App.Logger().Info("request", "body", string(bodyBytes))
	var req ActivationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate required fields
	if req.Payload.Algorithm != "hmac-sha256" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported algorithm"})
	}

	if req.Payload.Challenge == "" || req.Payload.HMAC == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "missing required fields"})
	}

	// Get device by MAC address
	device, err := m.Store.GetDeviceByMacAddress(macAddress)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "device not found"})
	}

	// Verify challenge matches the one stored for this device
	if device.Challenge != req.Payload.Challenge {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge"})
	}

	// Skip HMAC validation if hmac_key is empty
	if device.HmacKey != "" {
		// Calculate expected HMAC using device's hmac_key
		h := hmac.New(sha256.New, []byte(device.HmacKey))
		h.Write([]byte(req.Payload.Challenge))
		expectedMAC := h.Sum(nil)
		expectedMACHex := fmt.Sprintf("%x", expectedMAC)

		// Compare HMACs (timing-safe comparison would be better for production)
		if !hmac.Equal([]byte(req.Payload.HMAC), []byte(expectedMACHex)) {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid hmac"})
		}
	}

	if device.Status != types.DeviceCodeVerified {
		return e.JSON(http.StatusBadRequest, map[string]string{"message": "Waiting for device code submission"})
	}

	// Update device activation status
	if err := m.Store.ActivateDevice(device.Id, req.Payload.SerialNumber); err != nil {
		e.App.Logger().Error("Failed to activate device", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to activate device"})
	}

	return e.JSON(http.StatusOK, ActivationResponse{
		Success: true,
		Message: "Device activated successfully",
	})
}

type DeviceBindRequest struct {
	Code    string `json:"code"`
	AgentID string `json:"agentId"`
}

// otaBindDeviceRequest /xiaozhi/ota/bind-device
func (m *Manager) otaBindDeviceRequest(e *core.RequestEvent) error {
	authRecord := e.Auth
	if authRecord == nil {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
	}

	var req DeviceBindRequest
	if err := e.BindBody(&req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Code == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "code is required"})
	}

	// Find pending device by bind code in database
	device, err := m.Store.GetDeviceByBindCode(req.Code)
	if err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired device code"})
	}

	// Check if device is already bound
	if device.UserId != "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Device already bound"})
	}

	var agentUserID string
	var agentID string

	if req.AgentID != "" {
		// Use existing agent
		agent, err := m.Store.GetAgentByID(req.AgentID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
		}
		agentUserID = agent.UserID
		agentID = req.AgentID
	} else {
		// Create new agent with device board name
		agentName := device.Board
		if agentName == "" {
			agentName = req.Code
		}

		agent, err := m.Store.CreateNewAgent(authRecord.Id, agentName)
		if err != nil {
			e.App.Logger().Error("Failed to create new agent", "error", err, "userId", authRecord.Id, "agentName", agentName)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new agent"})
		}

		agentUserID = authRecord.Id
		agentID = agent.ID
	}

	if err := m.Store.BindDevice(device.MacAddress, agentID, agentUserID); err != nil {
		e.App.Logger().Error("Failed to save device binding", "error", err, "mac", device.MacAddress)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save device binding"})
	}

	return e.JSON(http.StatusOK, successResponse(true))
}
