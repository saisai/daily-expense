package cmd

import (
	"expense-tracker/models"
	"expense-tracker/pkg/logger"
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

		totals, monthlyTotal, err := models.GetMonthlyTotals(month)
		if err != nil {
			panic(err)
		}

		for _, t := range totals {
			fmt.Printf("%s : %.2f\n", t.Date, t.Total)
		}

		logger.Info(fmt.Sprintf("Monthly Total: %.2f\n", monthlyTotal))
	},
}
