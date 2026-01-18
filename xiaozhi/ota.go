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

	"github.com/pocketbase/dbx"
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
		e.App.Logger().Warn("Failed to unmarshal ota request", "body", bodyString, "error", err)
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ota request"})
	}

	// Store request body to ota_requests collection
	collection, err := e.App.FindCollectionByNameOrId("ota_requests")
	if err == nil {
		record := core.NewRecord(collection)
		record.Set("mac_address", deviceID)
		record.Set("board_type", req.Board.Type)
		record.Set("body_json", bodyString)
		if err := e.App.Save(record); err != nil {
			e.App.Logger().Error("Failed to save ota_request", "error", err)
		}
	}

	// Get websocket url and secret from sys_params
	wsURL := ""
	secret := ""

	var params []struct {
		Name  string `db:"name"`
		Value string `db:"value"`
	}
	err = e.App.DB().Select("name", "value").
		From("sys_params").
		Where(dbx.HashExp{"name": []any{"server.websocket", "server.secret"}}).
		All(&params)

	if err == nil {
		for _, p := range params {
			if p.Name == "server.websocket" {
				wsURL = p.Value
			} else if p.Name == "server.secret" {
				secret = p.Value
			}
		}
	}

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

	return e.JSON(http.StatusOK, response)
}
