package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(3, "Bundled example #3 with anonymous functions",
		func(db *mongo.Client) error {
			println("Another Bundled example UP!")
			return nil
		},

		func(db *mongo.Client) error {
			println("Another Bundled example DOWN!")
			return nil
		})
}
