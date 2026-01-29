package store

import (
	"github.com/pocketbase/pocketbase/core"
)

const OTARequestsCollectionName = "ota_requests"

func (m *Manager) LogOTARequest(macAddress, boardType, bodyJson string) error {
	collection, err := m.App.FindCollectionByNameOrId(OTARequestsCollectionName)
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("mac_address", macAddress)
	record.Set("board_type", boardType)
	record.Set("body_json", bodyJson)

	return m.App.Save(record)
}
