package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/gookit/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack/v5"
)

func setupCli() error {
	root := cobra.Command{
		Use:   "bud",
		Short: "tool to manage your money and brain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		DisableFlagParsing:         false,
		TraverseChildren:           false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 2,
	}

	root.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	root.AddCommand(
		buildLs().Command,
		buildEdit().Command,
		buildAdd().Command,
		buildFlush().Command,
		buildDelete().Command,
	)

	return root.Execute()
}

type LsCmd struct {
	*cobra.Command
}

func buildLs() LsCmd {
	c := LsCmd{
		Command: &cobra.Command{
			Use:               "ls",
			Short:             "List all entries",
			SilenceErrors:     true,
			SilenceUsage:      true,
			ValidArgsFunction: cobra.NoFileCompletions,
		},
	}

	c.RunE = c.Main()
	return c
}

func (c *LsCmd) Main() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		items, err := loadData()
		if err != nil {
			return fmt.Errorf("failed to load entries: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("no entries found")
			return nil
		}

		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})

		var totalIncome, totalExpense float64

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{
			color.Style{color.OpBold}.Sprint("IDX"),
			color.Style{color.OpBold}.Sprint("NAME"),
			color.Style{color.OpBold}.Sprint("TYPE"),
			color.Style{color.OpBold}.Sprint("AMOUNT"),
			color.Style{color.OpBold}.Sprint("CURRENCY"),
			color.Style{color.OpBold}.Sprint("CONFIDENCE"),
		})

		for i, item := range items {
			printer := item.Color()
			t.AppendRow(table.Row{
				printer.Sprint(i),
				printer.Sprint(item.Name),
				printer.Sprint(item.Type),
				printer.Sprintf("%.2f", item.Amount),
				printer.Sprint(item.Currency),
				printer.Sprintf("%.2f", item.Confidence),
			})

			switch item.Type {
			case "income":
				totalIncome += item.Amount
			case "expense":
				totalExpense += item.Amount
			}
		}

		t.SetStyle(table.Style{Box: table.StyleBoxDefault})

		t.Render()

		green := color.FgGreen
		red := color.FgRed

		fmt.Printf("\nTotal Income: %s%.2f PEN\n", green.Sprintf("+"), totalIncome)
		fmt.Printf("Total Expense: %s%.2f PEN\n", red.Sprintf("-"), totalExpense)

		net := totalIncome - totalExpense
		if net > 0 {
			fmt.Printf("Net Balance: %s%.2f PEN\n", green.Sprintf("+"), net)
		} else if net < 0 {
			fmt.Printf("Net Balance: %s%.2f PEN\n", red.Sprintf("-"), -net)
		} else {
			fmt.Printf("Net Balance: %.2f PEN\n", net)
		}

		return nil
	}
}

type AddCmd struct {
	*cobra.Command
}

func buildAdd() AddCmd {
	c := AddCmd{
		Command: &cobra.Command{
			Use:               "add",
			Short:             "Create a new entry",
			Args:              cobra.NoArgs,
			SilenceErrors:     true,
			SilenceUsage:      true,
			ValidArgsFunction: cobra.NoFileCompletions,
		},
	}

	c.RunE = c.Main()
	return c
}

func (c *AddCmd) Main() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		items, err := loadData()
		if err != nil {
			return fmt.Errorf("got error loading data: %w", err)
		}

		item := &BudgetItem{}

		if err := editItem(item); err != nil {
			return err
		}

		items = append(items, *item)

		if err := dumpData(items); err != nil {
			return fmt.Errorf("got error saving data: %w", err)
		}

		return nil
	}
}

type EditCmd struct {
	*cobra.Command
}

func buildEdit() EditCmd {
	c := EditCmd{
		Command: &cobra.Command{
			Use:               "edit [index]",
			Short:             "Edit an entry by its index",
			Args:              cobra.ExactArgs(1),
			SilenceErrors:     true,
			SilenceUsage:      true,
			ValidArgsFunction: cobra.NoFileCompletions,
		},
	}

	c.RunE = c.Main()
	return c
}

func (c *EditCmd) Main() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid index: %v", args[0])
		}

		items, err := loadData()
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		if index < 0 || index >= len(items) {
			return fmt.Errorf("index out of range (%d, total %d)", index, len(items))
		}

		if err := editItem(&items[index]); err != nil {
			return fmt.Errorf("edit aborted: %w", err)
		}

		if err := dumpData(items); err != nil {
			return fmt.Errorf("failed to save data: %w", err)
		}

		return nil
	}
}

type FlushCmd struct {
	*cobra.Command
}

func buildFlush() FlushCmd {
	c := FlushCmd{
		Command: &cobra.Command{
			Use:           "flush",
			Short:         "Remove all entries",
			Args:          cobra.NoArgs,
			SilenceErrors: true,
			SilenceUsage:  true,
		},
	}

	c.RunE = c.Main()
	return c
}

func (c *FlushCmd) Main() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		dir, err := userDataDir()
		if err != nil {
			return err
		}
		filePath := filepath.Join(dir, "bud.msg")

		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to flush data: %w", err)
		}

		fmt.Println("all entries removed")
		return nil
	}
}

type DeleteCmd struct {
	*cobra.Command
}

func buildDelete() DeleteCmd {
	c := DeleteCmd{
		Command: &cobra.Command{
			Use:           "delete [index]",
			Short:         "Delete an entry by its index",
			Args:          cobra.ExactArgs(1),
			SilenceErrors: true,
			SilenceUsage:  true,
		},
	}

	c.RunE = c.Main()
	return c
}

func (c *DeleteCmd) Main() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid index: %v", args[0])
		}

		items, err := loadData()
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		if index < 0 || index >= len(items) {
			return fmt.Errorf("index out of range (%d, total %d)", index, len(items))
		}

		// Remove the item
		items = append(items[:index], items[index+1:]...)

		if err := dumpData(items); err != nil {
			return fmt.Errorf("failed to save data: %w", err)
		}

		fmt.Printf("entry %d deleted\n", index)
		return nil
	}
}

func loadData() ([]BudgetItem, error) {
	dir, err := userDataDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, "bud.msg")

	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(filePath)
			if err != nil {
				return nil, err
			}

			f.Close()

			return loadData()

		}

		return nil, err
	}

	var items []BudgetItem

	if err := msgpack.Unmarshal(b, &items); err != nil {
		if err.Error() == "EOF" {
			return items, nil
		}

		return nil, err
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

func dumpData(items []BudgetItem) error {
	dir, err := userDataDir()
	if err != nil {
		return err
	}

	b, err := msgpack.Marshal(items)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "bud.msg"), b, 0o644)
}

func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}

	// fall back to something sane on all platforms
	return os.UserConfigDir()
}
