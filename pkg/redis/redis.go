package redisdb

import (
	"context"
	constants "expense-tracker/const"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func getEnvFilePath() (string, error) {
	// Get the path of the running executable
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Get the directory of the executable
	dir := filepath.Dir(exePath)

	// Now you can append the .env file name to this path
	envFilePath := filepath.Join(dir, ".env")
	return envFilePath, nil
}

func InitRedis() error {

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

	Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}
	// log.Println("✅ Redis connected")
	return err
}
