package services

import "github.com/pocketbase/pocketbase/core"

// ServiceContainer holds references to all services
type ServiceContainer struct {
	Device  DeviceService
	Session SessionService
	History HistoryService
}

// NewServiceContainer creates a new service container
func NewServiceContainer(app core.App) *ServiceContainer {
	return &ServiceContainer{
		Device:  NewDeviceService(app),
		Session: NewSessionService(app),
		History: NewHistoryService(app),
	}
}
