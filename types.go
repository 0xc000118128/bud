package main

import "github.com/gookit/color"

type Status int

type BudgetItem struct {
	Name       string  `msgpack:"n"`
	Type       string  `msgpack:"t"`
	Amount     float64 `msgpack:"a"`
	Currency   string  `msgpack:"cu"`
	Confidence float64 `msgpack:"co"` // 0.0 to 1.0
}

func (bi BudgetItem) Color() color.Color {
	switch bi.Type {
	case "income":
		return color.FgGreen
	case "expense":
		return color.FgRed
	case "pending_income":
		return color.FgYellow
	case "pending_expense":
		return color.FgMagenta
	default:
		return color.FgWhite
	}
}
