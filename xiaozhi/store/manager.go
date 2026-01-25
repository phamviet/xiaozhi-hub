package store

import "github.com/pocketbase/pocketbase/core"

type Manager struct {
	core.App
	DeviceCollection *core.Collection
}

func NewManager(app core.App) *Manager {
	manager := &Manager{App: app}
	collection, _ := app.FindCollectionByNameOrId(DeviceCollectionName)

	if collection != nil {
		manager.DeviceCollection = collection
	}

	return manager
}
