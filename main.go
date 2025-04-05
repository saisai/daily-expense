package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Connect or create SQLite DB
	db, err := sql.Open("sqlite3", "./expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: add <amount> <description>")
			return
		}
		amount, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Println("Invalid amount.")
			return
		}

		description := os.Args[3]
		addExpese(db, amount, description)
	case "list":
		listExpenses(db)
	case "total":
		totalToday(db)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add <amount> <description>   Add a new expense")
	fmt.Println("  list                         List all expenses")
	fmt.Println("  total                        Show total for today")
}

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS expenses(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func addExpese(db *sql.DB, amount float64, description string) {
	now := time.Now()
	_, err := db.Exec("INSERT INTO expenses(amount, description, created_at) VALUES (?, ?, ?)",
		amount, description, now.Format("2006-01-02 15:04:05"))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Expense added successfully.")
}

func listExpenses(db *sql.DB) {
	rows, err := db.Query("SELECT id, amount, description, created_at FROM expenses ORDER BY created_at DESC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("All Expenses:")
	for rows.Next() {
		var id int
		var amount float64
		var desc, createdAt string
		err := rows.Scan(&id, &amount, &desc, &createdAt)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[%d] $%.2f - %s at %s\n", id, amount, desc, createdAt)
	}
}

func totalToday(db *sql.DB) {
	today := time.Now().Format("2006-01-02")
	row := db.QueryRow(`
		SELECT IFNULL(SUM(amount), 0)
		FROM expenses
		WHERE DATE(created_at) = ?
	`, today)

	var total float64
	err := row.Scan(&total)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total spent today (%s): $%.2f\n", today, total)
}
