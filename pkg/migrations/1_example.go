package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(1, "Bundled example #1", uniqueNameFunction_UP, uniqueNameFunction_DOWN)
}

func uniqueNameFunction_UP(db *mongo.Client) error {
	println("Bundled example UP!")
	return nil
}

func uniqueNameFunction_DOWN(db *mongo.Client) error {
	println("Bundled example DOWN!")
	return nil
}
