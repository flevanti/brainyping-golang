package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	bisonmigration.RegisterMigration(2, "Bundled another example with sequence 2", anotherUniqueNameFunction222_UP, anotherUniqueNameFunction222_DOWN)
}

func anotherUniqueNameFunction222_UP(db *mongo.Client) error {
	println("Another Bundled example UP!")
	return nil
}

func anotherUniqueNameFunction222_DOWN(db *mongo.Client) error {
	println("Another Bundled example DOWN!")
	return nil
}
