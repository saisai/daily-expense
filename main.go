package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
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
	case "export":
		exportExpenses(db)
	case "monthly-report":
		exportMonthlyReport(db)
	case "monthly-xlsx":
		exportMonthlyReportXLSX(db)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add <amount> <description>   Add a new expense")
	fmt.Println("  list                         List all expenses")
	fmt.Println("  total                        Show total for today")
	fmt.Println("  export                        Export expenses to CSV")
	fmt.Println("  monthly-report                Export monthly totals to CSV")
	fmt.Println("  monthly-xlsx                            Export monthly totals to Excel (.xlsx)")
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

func exportExpenses(db *sql.DB) {
	rows, err := db.Query("SELECT amount, description, created_at FROM expenses ORDER BY created_at DESC;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	file, err := os.Create("expenses.csv")
	if err != nil {
		log.Fatal("Could not create CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	writer.Write([]string{"Amount", "Descriptioin", "Date"})

	// Write each row
	for rows.Next() {
		var amount float64
		var description, createdAt string
		err := rows.Scan(&amount, &description, &createdAt)
		if err != nil {
			log.Fatal(err)
		}

		writer.Write([]string{
			fmt.Sprintf("%.2f", amount),
			description,
			createdAt,
		})
	}
	fmt.Println("Expenses exported to expenses.csv")
}

func exportMonthlyReport(db *sql.DB) {
	rows, err := db.Query(`
		SELECT
			STRFTIME('%Y-%m', created_at) as month,
			IFNULL(SUM(amount), 0) AS total
		FROM expenses
		GROUP BY month
		ORDER BY month DESC
	`)
	if err != nil {
		log.Fatal("Failed to query monthly totals:", err)
	}
	defer rows.Close()

	// Create csv file
	file, err := os.Create("monthly_report.csv")
	if err != nil {
		log.Fatal("Could not create CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	writer.Write([]string{"Month", "Total Amount"})

	// Write each month's data

	for rows.Next() {
		var month string
		var total float64
		err := rows.Scan(&month, &total)
		if err != nil {
			log.Fatal(err)
		}
		writer.Write([]string{month, fmt.Sprintf("%.2f", total)})
	}

	fmt.Println("Monthly report exported to monthly_report.csv")
}

func exportMonthlyReportXLSX(db *sql.DB) {
	rows, err := db.Query(`
		SELECT 
			STRFTIME('%Y-%m', created_at) AS month,
			IFNULL(SUM(amount), 0) AS total
		FROM expenses
		GROUP BY month
		ORDER BY month DESC
	`)
	if err != nil {
		log.Fatal("Failed to query monthly totals:", err)
	}
	defer rows.Close()

	// Create a new Excel file
	f := excelize.NewFile()
	sheet := "Monthly Report"
	f.NewSheet(sheet)

	// Header row
	f.SetCellValue(sheet, "A1", "Month")
	f.SetCellValue(sheet, "B1", "Total Amount")

	rowIndex := 2
	for rows.Next() {
		var month string
		var total float64
		err := rows.Scan(&month, &total)
		if err != nil {
			log.Fatal(err)
		}

		cellMonth := fmt.Sprintf("A%d", rowIndex)
		cellTotal := fmt.Sprintf("B%d", rowIndex)
		fmt.Println("month", month)
		fmt.Println("total", total)
		f.SetCellValue(sheet, cellMonth, month)
		f.SetCellValue(sheet, cellTotal, total)

		rowIndex++
	}

	// Save the file
	err = f.SaveAs("monthly_report.xlsx")
	if err != nil {
		log.Fatal("Failed to save Excel file:", err)
	}

	fmt.Println("Monthly report exported to monthly_report.xlsx")
}
