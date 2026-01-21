package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_333196930")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_2153001328",
			"hidden": false,
			"id": "relation154121870",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "device",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_333196930")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation154121870")

		return app.Save(collection)
	})
}
