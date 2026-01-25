package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const DeviceCollectionName = "ai_device"

func (m *Manager) GetDeviceByMacAddress(macAddress string) (*types.Device, error) {
	var row types.Device
	err := m.App.RecordQuery(DeviceCollectionName).Where(dbx.NewExp("mac_address = {:p}", dbx.Params{"p": macAddress})).One(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (m *Manager) GetDeviceById(id string) (*types.Device, error) {
	var row types.Device
	err := m.App.RecordQuery(DeviceCollectionName).Where(dbx.NewExp("id = {:p}", dbx.Params{"p": id})).One(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (m *Manager) GetDeviceByBindCode(bindCode string) (*types.Device, error) {
	var row types.Device
	err := m.App.RecordQuery(DeviceCollectionName).Where(dbx.NewExp("bind_code = {:p}", dbx.Params{"p": bindCode})).One(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

type BindInfo struct {
	BindCode  string `json:"code"`
	Challenge string `json:"challenge"`
}

func (m *Manager) CreateUnboundDevice(macAddress string) (*BindInfo, error) {
	record := core.NewRecord(m.DeviceCollection)
	bindCode := fmt.Sprintf("%06d", mathrand.Intn(1000000))

	// Generate a random 32-byte challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, err
	}
	challenge := hex.EncodeToString(challengeBytes)

	record.Set("mac_address", macAddress)
	record.Set("bind_code", bindCode)
	record.Set("challenge", challenge)

	err := m.App.Save(record)
	if err != nil {
		return nil, err
	}

	return &BindInfo{BindCode: bindCode, Challenge: challenge}, nil
}

func (m *Manager) ActivateDevice(deviceID, serialNumber string) error {
	record, err := m.App.FindRecordById(DeviceCollectionName, deviceID)
	if err != nil {
		return err
	}

	record.Set("bind_code", "")
	record.Set("challenge", "")
	record.Set("status", types.DeviceBound)
	record.Set("serial_number", serialNumber)

	return m.App.Save(record)
}
