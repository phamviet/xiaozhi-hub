package services

import (
	"errors"

	"github.com/pocketbase/pocketbase/core"
)

type DeviceService interface {
	ValidateDevice(macAddress string) (string, error)
}

type deviceService struct {
	app core.App
}

func NewDeviceService(app core.App) DeviceService {
	return &deviceService{app: app}
}

// ValidateDevice checks if the device exists and is bound
func (s *deviceService) ValidateDevice(macAddress string) (string, error) {
	// Find the device by MAC address
	record, err := s.app.FindFirstRecordByFilter("ai_device", "mac_address = {:mac}", map[string]interface{}{
		"mac": macAddress,
	})

	if err != nil {
		return "", err
	}

	if record.GetString("status") != "bound" {
		return "", errors.New("device not bound")
	}

	return record.Id, nil
}
