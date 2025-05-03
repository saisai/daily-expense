package cmd

import (
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

		models.DeleteExpense(id)

		fmt.Println("Expense deleted!")
	},
}
