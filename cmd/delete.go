package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an expense byy ID",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: delete <id>")
			return
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}

		database, err := db.ConnectPostgres()
		if err != nil {
			panic(err)
		}
		defer database.Close()

		if err := models.DeleteExpense(database, id); err != nil {
			panic(err)
		}

		fmt.Println("Expense deleted!")
	},
}
