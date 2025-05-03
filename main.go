package main

import (
	"expense-tracker/cmd"
	"expense-tracker/db"
	"expense-tracker/models"
)

func main() {
	db.Init()
	db.DB.AutoMigrate(&models.Expense{})
	cmd.Execute()
}
