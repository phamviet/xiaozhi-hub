package xiaozhi

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type BindingInfo struct {
	MacAddress string
	ClientId   string
	Code       string
	CreatedAt  time.Time
}

type DeviceBindingManager struct {
	mu       sync.RWMutex
	bindings map[string]*BindingInfo // key: mac_address + "|" + client_id
}

func NewDeviceBindingManager() *DeviceBindingManager {
	return &DeviceBindingManager{
		bindings: make(map[string]*BindingInfo),
	}
}

func (dbm *DeviceBindingManager) GetOrGenerateCode(macAddress, clientId string) string {
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	key := fmt.Sprintf("%s|%s", macAddress, clientId)
	if info, ok := dbm.bindings[key]; ok {
		// Optional: Implement expiration logic here if needed
		return info.Code
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	dbm.bindings[key] = &BindingInfo{
		MacAddress: macAddress,
		ClientId:   clientId,
		Code:       code,
		CreatedAt:  time.Now(),
	}

	return code
}

func (dbm *DeviceBindingManager) VerifyCode(code string) (*BindingInfo, bool) {
	dbm.mu.RLock()
	defer dbm.mu.RUnlock()

	for _, info := range dbm.bindings {
		if info.Code == code {
			// Optional: check expiration
			return info, true
		}
	}

	return nil, false
}

func (dbm *DeviceBindingManager) RemoveBinding(macAddress, clientId string) {
	dbm.mu.Lock()
	defer dbm.mu.Unlock()

	key := fmt.Sprintf("%s|%s", macAddress, clientId)
	delete(dbm.bindings, key)
}
