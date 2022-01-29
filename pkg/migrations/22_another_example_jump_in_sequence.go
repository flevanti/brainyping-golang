package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(22, "Bundled example #2", anotherUniqueNameFunction22_UP, anotherUniqueNameFunction22_DOWN)
}

func anotherUniqueNameFunction22_UP(db *mongo.Client) error {
	println("Another Bundled example UP!")
	return nil
}

func anotherUniqueNameFunction22_DOWN(db *mongo.Client) error {
	println("Another Bundled example DOWN!")
	return nil
}
