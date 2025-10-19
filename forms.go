package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

func editItem(item *BudgetItem) error {
	var amountStr string
	var confidenceStr string

	if item.Amount != 0 {
		amountStr = fmt.Sprintf("%f", item.Amount)
	}

	if item.Confidence != 0 {
		confidenceStr = fmt.Sprintf("%f", item.Confidence)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Transaction type").
				Description("Choose what kind of entry this is.").
				Options(
					huh.NewOption("Expense", "expense"),
					huh.NewOption("Income", "income"),
					huh.NewOption("Pending Expense", "pending_expense"),
					huh.NewOption("Pending Income", "pending_income"),
				).
				Value(&item.Type),

			huh.NewInput().
				Title("Name").
				Description("Label this entry (e.g. groceries, salary, refund)").
				Value(&item.Name),

			huh.NewInput().
				Title("Amount").
				Description("Enter the amount as a number").
				Value(&amountStr),

			huh.NewSelect[string]().
				Title("Currency").
				Options(huh.NewOptions("PEN", "USD")...).
				Value(&item.Currency),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Println("cancelled, nothing saved")
		return err
	}

	if val, err := strconv.ParseFloat(amountStr, 64); err == nil {
		item.Amount = val
	}

	if strings.HasPrefix(item.Type, "pending_") {
		confForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Confidence").
					Description("How likely is this to happen? (0.0 to 1.0)").
					Value(&confidenceStr),
			),
		)

		if err := confForm.Run(); err != nil {
			return err
		}

		if val, err := strconv.ParseFloat(confidenceStr, 64); err == nil {
			item.Confidence = val
		}
	} else {
		item.Confidence = 1.0
	}

	return nil
}
