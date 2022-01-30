package main

var template = `package migrations

import (
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

func up_{{sequence}}(db *mongo.Client) error {
	// Your code here
	return nil
}

func down_{{sequence}}(db *mongo.Client) error {
	//your code here
	return nil
}

//
//this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration({{sequence}}, "f{{name}}", up_{{sequence}}, down_{{sequence}})
}
`
