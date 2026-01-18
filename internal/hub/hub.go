package hub

import (
	"context"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Hub struct {
	core.App
	appURL  string
	plugins []Plugin
}

func NewHub(app core.App, plugins []Plugin) *Hub {
	hub := &Hub{
		plugins: plugins,
	}
	hub.App = app

	hub.appURL, _ = os.LookupEnv("APP_URL")

	return hub
}

func (h *Hub) StartHub() error {
	for _, p := range h.plugins {
		h.App.Logger().Info("Registering plugin...", "name", p.Name())
		if err := p.Initialize(h); err != nil {
			return err
		}
	}

	h.App.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		ctx := context.Background()
		if err := h.bootstrap(ctx); err != nil {
			return err
		}

		return e.Next()
	})

	h.App.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := h.preStart(e); err != nil {
			return err
		}

		if err := h.startServer(e); err != nil {
			return err
		}

		return e.Next()
	})

	h.App.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		return e.Next()
	})

	if pb, ok := h.App.(*pocketbase.PocketBase); ok {
		err := pb.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

// preStart sets up initial configuration (collections, settings, etc.)
func (h *Hub) preStart(e *core.ServeEvent) error {
	if err := h.registerAuthRoutes(e); err != nil {
		return err
	}

	// set general settings
	settings := e.App.Settings()
	saveChanges := false

	// Enable batch requests
	if !settings.Batch.Enabled {
		settings.Batch.Enabled = true
		saveChanges = true
	}

	// set URL if BASE_URL env is set
	if h.appURL != "" && h.appURL != settings.Meta.AppURL {
		settings.Meta.AppURL = h.appURL
		saveChanges = true
	}

	if saveChanges {
		if err := e.App.Save(settings); err != nil {
			return err
		}
	}

	return nil
}

// custom api routes
func (h *Hub) registerAuthRoutes(se *core.ServeEvent) error {
	// auth optional routes
	apiNoAuth := se.Router.Group("/api")
	// check if first time setup on login page
	apiNoAuth.GET("/first-run", func(e *core.RequestEvent) error {
		total, err := e.App.CountRecords("users")
		return e.JSON(http.StatusOK, map[string]bool{"firstRun": err == nil && total == 0})
	})

	return nil
}
