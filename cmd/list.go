package cmd

import (
	"expense-tracker/models"
	redisdb "expense-tracker/pkg/redis"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exportFile string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all expense details for a month, and optionally export to Excel",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: list <YYYY-MM> [--export filename.xlsx]")
			os.Exit(1)
		}

		month := args[0]

		err := redisdb.InitRedis()
		if err != nil {
			fmt.Printf("Redis push error for key %s: %v\n", month, err)
		}

		cacheResult, err := models.GetExpensesFromRedis(month)
		if err != nil {
			fmt.Printf("Redis push error for key %s: %v\n", month, err)
		}
		// for _, d := range result {
		// 	fmt.Println(d)
		// }

		var expenses []models.Expense

		if len(cacheResult) > 0 {
			expenses = cacheResult
			fmt.Println("testing")
		} else {
			expenses, err = models.ListExpensesByMonth(month)
		}

		if err != nil {
			fmt.Println("Error listing expenses:", err)
			os.Exit(1)
		}

		if exportFile != "" {
			if err := models.ExportExpensesDetailToExcel(expenses, month, exportFile); err != nil {
				fmt.Println("Error exporting to Excel:", err)
				os.Exit(1)
			}
			fmt.Println("Exported to", exportFile)
		} else {
			fmt.Println("ID\t\tDate\t\tDescription\t\tAmount")
			fmt.Println("--------------------------------------------------------------")
			total := 0.0
			totalRows := 0
			for _, e := range expenses {
				fmt.Printf("%d\t\t%s\t%-20s\t%.2f\n", e.ID, e.CreatedAt.Format("2006-01-02"), e.Description, e.Amount)
				total += e.Amount
				totalRows++
			}
			fmt.Printf("\nTotal Record: %d\t\t\t%-20s\t%.2f\n", totalRows, "Total", total)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&exportFile, "export", "e", "", "Export the list to an Excel file")
}

// var listCmd = &cobra.Command{
// 	Use:   "list",
// 	Short: "List all expense details for a month",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if len(args) < 1 {
// 			fmt.Println("Usage: list <YYYY-MM>")
// 			return
// 		}

// 		month := args[0]

// 		database, err := db.ConnectPostgres()
// 		if err != nil {
// 			fmt.Println("DB Connection failed:", err)
// 			os.Exit(1)
// 		}
// 		defer database.Close()

// 		expenses, err := models.ListExpensesByMonth(database, month)
// 		if err != nil {
// 			fmt.Println("Error listing expenses:", err)
// 			os.Exit(1)
// 		}

// 		fmt.Println("ID\t\tDate\t\tDescription\t\tAmount")
// 		fmt.Println("-------------------------------------------------------------")
// 		for _, e := range expenses {
// 			fmt.Printf("%d\t\t%s\t%-20s\t%.2s\n", e.ID, e.CreatedAt.Format("2006-01-02"), e.Description, e.Amount)
// 		}
// 	},
// }
