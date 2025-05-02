package db

import (
	"database/sql"
	constants "expense-tracker/const"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func initDB(db *sql.DB) {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS expenses (
            id SERIAL PRIMARY KEY,
            description TEXT NOT NULL,
            amount NUMERIC(10, 2) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS trash (
            id INTEGER,
            description TEXT,
            amount NUMERIC(10, 2),
            created_at TIMESTAMP
        );
    `)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
}

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

func ConnectPostgres() (*sql.DB, error) {

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

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	initDB(db)
	return db, nil
}
