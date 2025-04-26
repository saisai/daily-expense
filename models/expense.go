package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type Expense struct {
	ID          int
	Description string
	Amount      float64
	CreatedAt   time.Time
}

func AddExpense(db *sql.DB, expense *Expense) error {
	_, err := db.Exec(`
		INSERT INTO expenses (description, amount, created_at)
		VALUES ($1, $2, $3)
	`, expense.Description, expense.Amount, expense.CreatedAt)
	return err
}

func DeleteExpense(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM expenses WHERE id = $1`, id)
	return err
}

type DailyTotal struct {
	Date  string
	Total float64
}

func GetMonthlyTotals(db *sql.DB, month string) ([]DailyTotal, float64, error) {
	rows, err := db.Query(`
		SELECT DATE(created_at) as day, SUM(amount)
		FROM expenses
		WHERE TO_CHAR(created_at, 'YYYY-MM') = $1
		GROUP BY day
		ORDER BY day ASC
	`, month)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var totals []DailyTotal
	var monthlyTotal float64

	for rows.Next() {
		var day string
		var total float64
		if err := rows.Scan(&day, &total); err != nil {
			return nil, 0, err
		}
		totals = append(totals, DailyTotal{Date: day, Total: total})
		monthlyTotal += total
	}
	return totals, monthlyTotal, nil
}

func ExportMonthlyReport(db *sql.DB, month string, filename string) error {
	totals, monthlyTotal, err := GetMonthlyTotals(db, month)
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Monthly Report"
	index, err := f.NewSheet(sheet)

	if err != nil {
		panic(err)
	}

	f.SetActiveSheet(index)

	f.SetCellValue(sheet, "A1", "Date")
	f.SetCellValue(sheet, "B1", "Total")

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "B1", headerStyle)

	for i, t := range totals {
		f.SetCellValue(sheet, "A"+fmt.Sprint(i+2), t.Date)
		f.SetCellValue(sheet, "B"+fmt.Sprint(i+2), t.Total)
	}

	totalRow := len(totals) + 2
	f.SetCellValue(sheet, "A"+fmt.Sprint(totalRow), "Monthly Total")
	f.SetCellValue(sheet, "B"+fmt.Sprint(totalRow), monthlyTotal)

	totalStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "#FF0000"}})
	f.SetCellStyle(sheet, "A"+fmt.Sprint(totalRow), "B"+fmt.Sprint(totalRow), totalStyle)

	return f.SaveAs(filename)
}

func ListExpensesByMonth(db *sql.DB, month string) ([]Expense, error) {
	query := `
		SELECT id, description, amount, created_at
		FROM expenses
		WHERE TO_CHAR(created_at, 'YYYY-MM') = $1
		ORDER BY created_at ASC
	`

	rows, err := db.Query(query, month)
	if err != nil {
		return nil, fmt.Errorf("query expenses by month: %w", err)
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.Description, &e.Amount, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func ExportExpensesDetailToExcel(expenses []Expense, month, filename string) error {
	f := excelize.NewFile()

	sheet := "Expenses-" + month
	f.NewSheet(sheet)
	index, err := f.NewSheet(sheet)

	if err != nil {
		panic(err)
	}

	f.SetActiveSheet(index)

	// f.SetActiveSheet(f.GetSheetIndex(sheet))

	// Header
	f.SetCellValue(sheet, "A1", "Date")
	f.SetCellValue(sheet, "B1", "Description")
	f.SetCellValue(sheet, "C1", "Amount")

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "C1", headerStyle)

	for i, exp := range expenses {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), exp.CreatedAt.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), exp.Description)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), exp.Amount)
	}

	// ➡️ Add Total
	totalRow := len(expenses) + 2

	f.SetCellValue(sheet, fmt.Sprintf("B%d", totalRow), "Total")
	sumFormula := fmt.Sprintf("SUM(C2:C%d)", totalRow-1)
	f.SetCellFormula(sheet, fmt.Sprintf("C%d", totalRow), fmt.Sprintf("=%s", sumFormula))

	// Style the total row
	totalStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", totalRow), fmt.Sprintf("C%d", totalRow), totalStyle)

	f.DeleteSheet("Sheet1")

	return f.SaveAs(filename)
}
