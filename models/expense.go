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
	Amount      string
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
