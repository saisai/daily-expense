package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"expense-tracker/pkg/logger"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var dateStr string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new expense",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			fmt.Println("Usage: add <description> <amount>")
			return
		}

		description := args[0]
		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("Invalid amount.")
			return
		}

		database, err := db.ConnectPostgres()
		if err != nil {
			panic(err)
		}
		defer database.Close()

		dateStr = args[2]

		if err := models.AddExpense(database, description, amount, dateStr); err != nil {
			panic(err)
		}

		logger.Info("Expense added")
	},
}

func init() {
	addCmd.Flags().StringVarP(&dateStr, "date", "d", "", "Expense date (format: 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM')")

}
