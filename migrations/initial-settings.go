package migrations

import (
	"os"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/seeds"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	TempAdminEmail = "_@example.com"
)

func init() {
	m.Register(func(app core.App) error {
		// initial settings
		settings := app.Settings()
		settings.Meta.AppName = "Acme"
		settings.Meta.HideControls = true
		settings.Logs.MinLevel = 0
		if err := app.Save(settings); err != nil {
			return err
		}

		// create superuser
		superuserCollection, _ := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		superUser := core.NewRecord(superuserCollection)

		// set email
		email := os.Getenv("USER_EMAIL")
		password := os.Getenv("USER_PASSWORD")
		didProvideUserDetails := email != "" && password != ""

		// set superuser email
		if email == "" {
			email = TempAdminEmail
		}
		superUser.SetEmail(email)

		// set superuser password
		if password != "" {
			superUser.SetPassword(password)
		} else {
			superUser.SetRandomPassword()
		}

		// if user details are provided, we create a regular user as well
		if didProvideUserDetails {
			usersCollection, _ := app.FindCollectionByNameOrId("users")
			user := core.NewRecord(usersCollection)
			user.SetEmail(email)
			user.SetPassword(password)
			user.SetVerified(true)
			err := app.Save(user)
			if err != nil {
				return err
			}
		}

		if err := app.Save(superUser); err != nil {
			return err
		}

		if err := seeds.Seed(app); err != nil {
			app.Logger().Error("Database seeding error", "error", err.Error())
		}

		return nil
	}, nil)
}
