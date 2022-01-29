package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(
		0,
		"Bundled example #0 with anonymous functions - MERGED LATE!",
		func(db *mongo.Client) error {
			println("Another Bundled example UP!")
			return nil
		},
		func(db *mongo.Client) error {
			println("Another Bundled example DOWN!")
			return nil
		})
}
