package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(2, "Bundled example #2", anotherUniqueNameFunction_UP, anotherUniqueNameFunction_DOWN)
}

func anotherUniqueNameFunction_UP(db *mongo.Client) error {
	println("Another Bundled example UP!")
	return nil
}

func anotherUniqueNameFunction_DOWN(db *mongo.Client) error {
	println("Another Bundled example DOWN!")
	return nil
}
