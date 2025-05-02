package models

import (
	"database/sql"
	"encoding/json"
	redisdb "expense-tracker/pkg/redis"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type Expense struct {
	ID          int
	Description string
	Amount      float64
	CreatedAt   time.Time
}

func AddExpense(db *sql.DB, description string, amount float64, dateStr string) error {
	var createdAt time.Time
	var err error

	if dateStr == "" {
		createdAt = time.Now()
	} else {
		// Try full datetime first: "2006-01-02 15:04"
		createdAt, err = time.Parse("2006-01-02 15:04", dateStr)
		if err != nil {
			// If fail, try date only: "2006-01-02"
			createdAt, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format. Use 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM'")
			}
			// Set default time to noon if only date is given
			createdAt = createdAt.Add(12 * time.Hour)
		}
	}

	expense := Expense{
		Description: description,
		Amount:      amount,
		CreatedAt:   createdAt,
	}

	_, _ = db.Exec(`
		INSERT INTO expenses (description, amount, created_at)
		VALUES ($1, $2, $3)
	`, expense.Description, expense.Amount, expense.CreatedAt)

	year, month := createdAt.Year(), createdAt.Month()
	fmt.Println("created at ", year, fmt.Sprintf("%02d", int(month)))

	yearMonth := fmt.Sprintf("%s-%s", strconv.Itoa(year), string(fmt.Sprintf("%02d", int(month))))
	result, err := ListExpensesByMonth(db, yearMonth)
	if err != nil {
		return err
	}

	err = redisdb.InitRedis()
	if err != nil {
		return err
	}
	// fmt.Println(result)
	// err = redisdb.Rdb.RPush(redisdb.Ctx, yearMonth, json.Marshal(result)).Err()

	err = pushExpensesToRedis(yearMonth, result)
	if err != nil {
		fmt.Printf("Redis push error for key %s: %v\n", yearMonth, err)
		return err
	}

	// result, err = getExpensesFromRedis(yearMonth)
	// if err != nil {
	// 	fmt.Printf("Redis push error for key %s: %v\n", yearMonth, err)
	// 	return err
	// }
	// for _, d := range result {
	// 	fmt.Println(d)
	// }

	return err
}

func pushExpensesToRedis(key string, expenses []Expense) error {
	serializedExpenses := make([]string, len(expenses))
	for i, expense := range expenses {
		jsonData, err := json.Marshal(expense)
		if err != nil {
			return fmt.Errorf("failed to marshal expense to JSON: %w", err)
		}
		serializedExpenses[i] = string(jsonData)
	}

	err := redisdb.Rdb.RPush(redisdb.Ctx, key, serializedExpenses).Err()
	if err != nil {
		return fmt.Errorf("failed to push to Redis: %w", err)
	}
	return nil
}

func GetExpensesFromRedis(key string) ([]Expense, error) {

	stringExpenses, err := redisdb.Rdb.LRange(redisdb.Ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get data from Redis: %w", err)
	}

	expenses := make([]Expense, len(stringExpenses))
	for i, s := range stringExpenses {
		var expense Expense
		err := json.Unmarshal([]byte(s), &expense)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON to Expense: %w", err)
		}
		expenses[i] = expense
	}
	return expenses, nil
}

// func AddExpense(db *sql.DB, expense *Expense) error {
// 	_, err := db.Exec(`
// 		INSERT INTO expenses (description, amount, created_at)
// 		VALUES ($1, $2, $3)
// 	`, expense.Description, expense.Amount, expense.CreatedAt)
// 	return err
// }

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
