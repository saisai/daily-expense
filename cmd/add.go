package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new expense",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("args ", len(args))
		if len(args) < 2 {
			fmt.Println("Usage: add <description> <amount>")
			return
		}

		description := args[0]
		amount := args[1]

		database, err := db.ConnectPostgres()
		if err != nil {
			panic(err)
		}
		defer database.Close()

		expense := models.Expense{
			Description: description,
			Amount:      amount,
			CreatedAt:   time.Now(),
		}

		if err := models.AddExpense(database, &expense); err != nil {
			panic(err)
		}

		fmt.Println("Expense added")
	},
}
