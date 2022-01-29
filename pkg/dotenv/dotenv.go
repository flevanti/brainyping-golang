package dotenv

import (
	"github.com/joho/godotenv"
	"log"
)

func init() {
	getDotEnv()
}

func getDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file", err.Error())
	}
}
