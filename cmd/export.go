package cmd

import (
	"expense-tracker/db"
	"expense-tracker/models"
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export montly report to Excel",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Usage: export <YYYY-MM> <filename.xlsx>")
			return
		}

		month := args[0]
		filename := args[1]

		database, err := db.ConnectPostgres()
		if err != nil {
			panic(err)
		}
		defer database.Close()

		if err := models.ExportMonthlyReport(database, month, filename); err != nil {
			panic(err)
		}

		fmt.Println("Exported to", filename)
	},
}
