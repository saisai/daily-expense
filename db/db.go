package db

import (
	constants "expense-tracker/const"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {

	if constants.DEVELOPMENT_MODE == 0 {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

	} else {
		envFilePath, err := getEnvFilePath()
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}

		// Load the .env file using godotenv
		err = godotenv.Load(envFilePath)
		if err != nil {
			fmt.Printf("Error loading .env file at %s: %v", envFilePath, err)
		}
	}

	var err error
	dsn := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
}
