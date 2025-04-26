package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"fmt"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a monthly report",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: report <YYYY-MM>")
			return
		}

		month := args[0]

		database, err := db.ConnectPostgres()
		if err != nil {
			panic(err)
		}
		defer database.Close()
		totals, monthlyTotal, err := models.GetMonthlyTotals(database, month)
		if err != nil {
			panic(err)
		}

		for _, t := range totals {
			fmt.Printf("%s : %.2f\n", t.Date, t.Total)
		}
		fmt.Printf("Monthly Total: %.2f\n", monthlyTotal)
	},
}
