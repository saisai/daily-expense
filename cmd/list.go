package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all expense details for a month",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: list <YYYY-MM>")
			return
		}

		month := args[0]

		database, err := db.ConnectPostgres()
		if err != nil {
			fmt.Println("DB Connection failed:", err)
			os.Exit(1)
		}
		defer database.Close()

		expenses, err := models.ListExpensesByMonth(database, month)
		if err != nil {
			fmt.Println("Error listing expenses:", err)
			os.Exit(1)
		}

		fmt.Println("ID\t\tDate\t\tDescription\t\tAmount")
		fmt.Println("-------------------------------------------------------------")
		for _, e := range expenses {
			fmt.Printf("%d\t\t%s\t%-20s\t%.2s\n", e.ID, e.CreatedAt.Format("2006-01-02"), e.Description, e.Amount)
		}
	},
}
