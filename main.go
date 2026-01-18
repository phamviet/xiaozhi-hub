package main

import (
	"log"
	"os"

	"github.com/phamviet/xiaozhi-hub/internal/hub"
	_ "github.com/phamviet/xiaozhi-hub/migrations"
	"github.com/phamviet/xiaozhi-hub/xiaozhi"
	"github.com/phamviet/xiaozhi-hub/xiaozhi/seeds"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"
)

func main() {
	baseApp := initializeApp()
	h := hub.NewHub(baseApp, []hub.Plugin{xiaozhi.NewManager()})

	if err := h.StartHub(); err != nil {
		log.Fatal(err)
	}
}

func initializeApp() *pocketbase.PocketBase {
	isDev := os.Getenv("ENV") == "dev"

	baseApp := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: "./pb_data",
		DefaultDev:     isDev,
	})
	baseApp.RootCmd.Version = "0.0.1"
	baseApp.RootCmd.Use = "pb"
	baseApp.RootCmd.Short = ""

	// Enable auto creation of migration files when making collection changes in the Admin UI
	migratecmd.MustRegister(baseApp, baseApp.RootCmd, migratecmd.Config{
		Automigrate: false,
		Dir:         "./migrations",
	})

	baseApp.RootCmd.AddCommand(&cobra.Command{
		Use:   "seeds",
		Short: "Seed the database with initial data",
		Run: func(cmd *cobra.Command, args []string) {
			if err := seeds.Seed(baseApp); err != nil {
				log.Fatal(err)
			}
			log.Println("Database seeded successfully")
		},
	})

	return baseApp
}
