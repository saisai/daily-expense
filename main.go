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
	fmt.Println("start...")
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
		customTime := ""
		if len(os.Args) <= 5 {
			fmt.Println("Usage: add <amount> <description>  [date-time]")
			fmt.Println("Example with date-time: add 100 \"Lunch\" \"2025-04-01 12:30\"")
			customTime = os.Args[4]
			// return
		}
		amount, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Println("Invalid amount.")
			return
		}

		description := os.Args[3]
		fmt.Println("customTime ", customTime)
		addExpese(db, amount, description, customTime)
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
	case "detailed-xlsx":
		exportDetailedMonthlyExpenses(db)
	case "daily-monthly-xlsx":
		exportDailyToMonthlyWithTotals(db)
	case "delete":
		if len(os.Args) == 3 {
			id, _ := strconv.Atoi(os.Args[2])
			deleteExpense(db, id)
		} else if len(os.Args) == 4 && os.Args[2] == "date" {
			deleteByDate(db, os.Args[3])
		} else {
			fmt.Println("Usage: delete <id> OR delete date <YYYY-MM-DD>")
		}
	case "undo":
		undoDelete(db)
	case "report-daily":
		fmt.Println("report-daily")
		if len(os.Args) != 4 {
			fmt.Println("Usage: report-daily <YYYY-MM> <filename.xlsx>")
			return
		}
		month := os.Args[2]    // like "2025-04"
		filename := os.Args[3] // like "april_report.xlsx"
		if err := exportDailyTotalsToExcel(db, month, filename); err != nil {
			log.Fatal("Export failed:", err)
		}
		fmt.Println("Exported to", filename)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("add <amount> <description>   Add a new expense")
	fmt.Println("Optional date-time format: \"YYYY-MM-DD HH:MM\" (24h)")
	fmt.Println("Example: add 100 \"Groceries\"  \"2025-04-01 18:30\"")
	fmt.Println("Example: add 80 \"Snacks\" ")
	fmt.Println("list        List all expenses")
	fmt.Println("total       Show total for today")
	fmt.Println("export      Export expenses to CSV")
	fmt.Println("monthly-report  Export monthly totals to CSV")
	fmt.Println("monthly-xlsx    Export monthly totals to Excel (.xlsx)")
	fmt.Println("detailed-xlsx    Export daily expenses by month to Excel")
	fmt.Println("daily-monthly-xlsx    Export daily montly expenses by month to Excel")
	fmt.Println("delete To delete the expense by ID")
	fmt.Println("undo To restore all expenses from deleting.")
	fmt.Println("report-daily 2025-04 april_expenses.xlsx")
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS trash(
			id INTEGER,
			description TEXT,
			amount REAL,
			created_at DATETIME
		)
	`)
}

func addExpese(db *sql.DB, amount float64, description string, customTime ...string) {
	fmt.Println("addExpense")
	var t time.Time
	var err error
	if len(customTime) > 0 && customTime[0] != "" {
		t, err = time.Parse("2006-01-02 15:04", customTime[0])
		if err != nil {
			fmt.Println("Invalid date format. Use YYYY-MM-DD HH:MM (24hr format)")
			return
		}
	} else {
		t = time.Now()
	}

	_, err = db.Exec("INSERT INTO expenses(amount, description, created_at) VALUES (?, ?, ?)",
		amount, description, t.Format("2006-01-02 15:04:05"))

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
	sheet := "Sheet1"
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

func exportDetailedMonthlyExpenses(db *sql.DB) {
	rows, err := db.Query(`
		SELECT 
			DATE(created_at) as date,
			description,
			amount,
			STRFTIME('%Y-%m', created_at) as month 
		FROM expenses
		ORDER BY month ASC, date ASC 
	`)

	if err != nil {
		log.Fatal("Failed to query expenses:", err)
	}
	defer rows.Close()

	type Expense struct {
		Date        string
		Description string
		Amount      float64
	}

	monthlyData := make(map[string][]Expense)
	monthlyTotals := make(map[string]float64)

	for rows.Next() {
		var e Expense
		var month string
		if err := rows.Scan(&e.Date, &e.Description, &e.Amount, &month); err != nil {
			log.Fatal(err)
		}
		monthlyData[month] = append(monthlyData[month], e)
		monthlyTotals[month] += e.Amount
	}

	f := excelize.NewFile()
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	for month, expenses := range monthlyData {
		sheet := month
		if sheet != "Sheet1" {
			f.NewSheet(sheet)
		}
		f.SetCellValue(sheet, "A1", "Date")
		f.SetCellValue(sheet, "B1", "Description")
		f.SetCellValue(sheet, "C1", "Amount")
		f.SetCellStyle(sheet, "A1", "C1", styleBold)

		rowIndex := 2
		for _, e := range expenses {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIndex), e.Date)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), e.Description)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIndex), e.Amount)
			rowIndex++
		}

		// Add Total row
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), "Total")
		totalFormula := fmt.Sprintf("SUM(C2:C%d)", rowIndex-1)
		f.SetCellFormula(sheet, fmt.Sprintf("C%d", rowIndex), fmt.Sprintf("=%s", totalFormula))
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("C%d", rowIndex), styleBold)
	}

	// Remove default sheet if unused
	if len(monthlyData) > 0 {
		f.DeleteSheet("Sheet1")
	}

	err = f.SaveAs("expenses_report.xlsx")
	if err != nil {
		log.Fatal("Failed to save Excel file:", err)
	}

	fmt.Println("Detailed monthly expense report exported to expenses_report.xlsx")
}

func exportDailyToMonthlyWithTotals(db *sql.DB) {
	rows, err := db.Query(`
		SELECT 
			DATE(created_at) as date,
			description,
			amount,
			STRFTIME('%Y-%m', created_at) as month
		FROM expenses
		ORDER BY month ASC, date ASC, created_at ASC
	`)
	if err != nil {
		log.Fatal("Failed to query expenses:", err)
	}
	defer rows.Close()

	type Expense struct {
		Date        string
		Description string
		Amount      float64
	}

	// Map: month → []Expense
	monthlyData := make(map[string][]Expense)

	for rows.Next() {
		var e Expense
		var month string
		if err := rows.Scan(&e.Date, &e.Description, &e.Amount, &month); err != nil {
			log.Fatal(err)
		}
		monthlyData[month] = append(monthlyData[month], e)
	}

	f := excelize.NewFile()
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	for month, expenses := range monthlyData {
		sheet := month
		if sheet != "Sheet1" {
			f.NewSheet(sheet)
		}

		f.SetCellValue(sheet, "A1", "Date")
		f.SetCellValue(sheet, "B1", "Description")
		f.SetCellValue(sheet, "C1", "Amount")
		f.SetCellStyle(sheet, "A1", "C1", styleBold)

		rowIndex := 2
		var day string
		var dailyStartRow int
		var monthTotal float64

		for i, e := range expenses {
			// On new day, remember start row
			if day != e.Date {
				if dailyStartRow != 0 && i != 0 {
					// Insert daily total row
					dailyTotalCell := fmt.Sprintf("C%d", rowIndex)
					formula := fmt.Sprintf("SUM(C%d:C%d)", dailyStartRow, rowIndex-1)
					f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", day))
					f.SetCellFormula(sheet, dailyTotalCell, formula)
					f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), dailyTotalCell, styleBold)
					rowIndex++
				}
				day = e.Date
				dailyStartRow = rowIndex
			}

			// Write expense row
			f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIndex), e.Date)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), e.Description)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIndex), e.Amount)
			monthTotal += e.Amount
			rowIndex++
		}

		// Final daily total row (last day's total)
		if dailyStartRow != 0 {
			dailyTotalCell := fmt.Sprintf("C%d", rowIndex)
			formula := fmt.Sprintf("SUM(C%d:C%d)", dailyStartRow, rowIndex-1)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", day))
			f.SetCellFormula(sheet, dailyTotalCell, formula)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), dailyTotalCell, styleBold)
			rowIndex++
		}

		// Final month total
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", month))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIndex), monthTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("C%d", rowIndex), styleBold)
	}

	// Remove default "Sheet1" if unused
	if len(monthlyData) > 0 {
		f.DeleteSheet("Sheet1")
	}

	err = f.SaveAs("daily_monthly_report.xlsx")
	if err != nil {
		log.Fatal("Failed to save Excel file:", err)
	}

	fmt.Println("Report exported to daily_monthly_report.xlsx")
}

func exportDailyToMonthlyWithTotalsOld(db *sql.DB) {
	rows, err := db.Query(`
		SELECT 
			DATE(created_at) as date,
			description,
			amount,
			STRFTIME('%Y-%m', created_at) as month
		FROM expenses 
		ORDER BY month ASC, date ASC, created_at ASC
	`)

	if err != nil {
		log.Fatal("Failed to query expenses:", err)
	}
	defer rows.Close()

	type Expense struct {
		Date        string
		Description string
		Amount      float64
	}

	// Map: month -> []Expense
	monthlyData := make(map[string][]Expense)

	for rows.Next() {
		var e Expense
		var month string
		if err := rows.Scan(&e.Date, &e.Description, &e.Amount, &month); err != nil {
			log.Fatal(err)
		}
		monthlyData[month] = append(monthlyData[month], e)
	}

	f := excelize.NewFile()
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	for month, expenses := range monthlyData {
		sheet := month
		if sheet != "Sheet1" {
			f.NewSheet(sheet)
		}

		f.SetCellValue(sheet, "A1", "Date")
		f.SetCellValue(sheet, "B1", "Description")
		f.SetCellValue(sheet, "C1", "Amount")
		f.SetCellStyle(sheet, "A1", "C1", styleBold)

		rowIndex := 2
		var day string
		var dailyStartRow int
		var monthTotal float64

		for i, e := range expenses {
			// On new day, remember start row
			if day != e.Date {
				if dailyStartRow != 0 && i != 0 {
					// Insert daily total row
					dailyTotalCell := fmt.Sprintf("C%d", rowIndex)
					formula := fmt.Sprintf("SUM(C%d:C%d)", dailyStartRow, rowIndex-1)
					f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", day))
					f.SetCellFormula(sheet, dailyTotalCell, formula)
					f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), dailyTotalCell, styleBold)
					rowIndex++
				}
				day = e.Date
				dailyStartRow = rowIndex
			}

			// Write expense row
			f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIndex), e.Date)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), e.Description)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIndex), e.Amount)
			monthTotal += e.Amount
			rowIndex++
		}

		// Final daily total row (Last day's total)
		if dailyStartRow != 0 {
			dailyTotalCell := fmt.Sprintf("C%d", rowIndex)
			formula := fmt.Sprintf("SUM(C%d:C%d)", dailyStartRow, rowIndex-1)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", day))
			f.SetCellValue(sheet, dailyTotalCell, formula)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), dailyTotalCell, styleBold)
			rowIndex++
		}

		// Final month total
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("Total (%s)", month))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIndex), monthTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIndex), fmt.Sprintf("C%d", rowIndex), styleBold)
	}

	// Remove default "Sheet1" if unused
	if len(monthlyData) > 0 {
		f.DeleteSheet("Sheet1")
	}

	err = f.SaveAs("daily_monthly_report.xlsx")
	if err != nil {
		log.Fatal("Failed to save Excel file:", err)
	}

	fmt.Println("Report exported to daily_monthly_report.xlsx")

}

func deleteExpenseOld(db *sql.DB, id int) {
	result, err := db.Exec("DELETE FROM expenses WHERE id = ?", id)
	if err != nil {
		log.Fatal("Failed to delete expense:", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Printf("No expense found with ID %d\n", id)
	} else {
		fmt.Printf("Deleted expense with ID %d\n", id)
	}
}

func deleteExpense(db *sql.DB, id int) {
	var exp struct {
		Description string
		Amount      float64
		CreatedAt   string
	}

	row := db.QueryRow("SELECT description, amount, created_at FROM expenses WHERE id=?", id)
	err := row.Scan(&exp.Description, &exp.Amount, &exp.CreatedAt)
	if err == sql.ErrNoRows {
		fmt.Printf("No expense found with ID %d\n", id)
		return
	} else if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Delete this expense?\n")
	fmt.Printf("-> %s | %s | %.2f\n", exp.CreatedAt, exp.Description, exp.Amount)
	fmt.Printf("Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Cancelled")
		return
	}

	// Backup to trash
	_, err = db.Exec(`
		INSERT INTO trash(id, description, amount, created_at)
		SELECT id, description, amount, created_at FROM expenses WHERE id = ?`, id)
	if err != nil {
		log.Fatal("Failed to backup to trash: ", err)
	}

	// Now delete
	_, err = db.Exec("DELETE FROM expenses WHERE id = ?", id)
	if err != nil {
		log.Fatal("Failed to delete:", err)
	}

	fmt.Printf("Deleted expense with ID %d. You can undo it with: undo\n", id)
}

func undoDelete(db *sql.DB) {
	row := db.QueryRow("SELECT id, description, amount, created_at FROM trash ORDER BY ROWID DESC LIMIT 1")

	var id int
	var desc, createdAt string
	var amount float64

	err := row.Scan(&id, &desc, &amount, &createdAt)
	if err == sql.ErrNoRows {
		fmt.Println("No deleted expense to undo.")
		return
	} else if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO expenses (id, description, amount, created_at)
		VALUES(?, ?, ?, ?)`, id, desc, amount, createdAt)
	if err != nil {
		log.Fatal("Failed to restore expense:", err)
	}

	_, err = db.Exec("DELETE FROM trash WHERE id=? AND created_at = ?", id, createdAt)
	if err != nil {
		log.Fatal("Failed to remove trash:", err)
	}

	fmt.Printf("Restored deleted expense: %s | %.2f\n", desc, amount)
}

func deleteByDate(db *sql.DB, date string) {
	rows, err := db.Query("SELECT id, description, amount, created_at FROM expenses WHERE DATE(created_at) = ?", date)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var toDelete []struct {
		ID          int
		Description string
		Amount      float64
		CreatedAt   string
	}

	for rows.Next() {
		var e struct {
			ID          int
			Description string
			Amount      float64
			CreatedAt   string
		}
		err := rows.Scan(&e.ID, &e.Description, &e.Amount, &e.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		toDelete = append(toDelete, e)
	}

	if len(toDelete) == 0 {
		fmt.Println("No expenses found for that date.")
		return
	}

	fmt.Printf("Delete all %d expenses from %s\n", len(toDelete), date)
	fmt.Print("Type 'yes' to confirm: ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	tx, _ := db.Begin()

	for _, e := range toDelete {
		_, err := tx.Exec(`
			INSERT INTO trash (id, description, amount, created_at)
			VALUES(?, ?, ?, ?)`, e.ID, e.Description, e.Amount, e.CreatedAt)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		_, err = tx.Exec(`DELETE FROM expenses WHERE id = ?`, e.ID)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}
	tx.Commit()

	fmt.Printf("Deleted %d expenses from %s. You can undo the last one with: undo\n", len(toDelete), date)
}

func getDailyTotals(db *sql.DB, month string) ([]struct {
	Date  string
	Total float64
}, float64, error) {
	rows, err := db.Query(`
		SELECT DATE(created_at) as day, SUM(amount) as total
		FROM expenses
		WHERE strftime('%Y-%m', created_at) = ?
		GROUP BY day
		ORDER BY day ASC
	`, month)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var data []struct {
		Date  string
		Total float64
	}
	var monthlyTotal float64

	for rows.Next() {
		var date string
		var total float64
		if err := rows.Scan(&date, &total); err != nil {
			return nil, 0, err
		}
		data = append(data, struct {
			Date  string
			Total float64
		}{date, total})
		monthlyTotal += total
	}

	return data, monthlyTotal, nil
}

func exportDailyTotalsToExcel(db *sql.DB, month string, filename string) error {
	data, monthlyTotal, err := getDailyTotals(db, month)
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.NewSheet(sheet)

	// f := excelize.NewFile()
	// styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	// f.SetActiveSheet(index)

	// Header
	f.SetCellValue(sheet, "A1", "Date")
	f.SetCellValue(sheet, "B1", "Total")

	// Bold header style
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "B1", headerStyle)

	// Daily rows
	for i, row := range data {
		cellDate := fmt.Sprintf("A%d", i+2)
		cellTotal := fmt.Sprintf("B%d", i+2)
		f.SetCellValue(sheet, cellDate, row.Date)
		f.SetCellValue(sheet, cellTotal, row.Total)
	}

	// Monthly total
	totalRow := len(data) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "Monthly Total")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", totalRow), monthlyTotal)

	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("B%d", totalRow), boldStyle)

	return f.SaveAs(filename)
}
