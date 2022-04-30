package config

import (
	"github.com/joho/godotenv"

	"log"
	"os"
)

const BasePath = "/v1"

//TODO: Make a struct (on its own file) and and instance for basic important setting
//TODO: Make a function to refresh each setting from the struct above
// dinamically get the variables, so
func Config(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Cannot load given key: " + key)
	}

	return os.Getenv(key)
}
