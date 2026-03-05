package app

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Level int

const (
	Info Level = iota
	Warn
	Fail
)

func printSection(title string) {
	fmt.Println()
	fmt.Println(title)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func logLine(level Level, format string, args ...any) {
	prefix := "•"
	switch level {
	case Info:
		prefix = "✓"
	case Warn:
		prefix = "⚠"
	case Fail:
		prefix = "✗"
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
}

func renderSummaryTable(runName string, aName, bName string, flashA, flashB, ramA, ramB int64) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Run", "Metric", aName, bName, "Delta"})

	t.AppendRow(table.Row{runName, "FLASH", flashA, flashB, flashB - flashA})
	t.AppendRow(table.Row{runName, "RAM", ramA, ramB, ramB - ramA})

	t.Render()
	fmt.Println()
}

func renderSymbolDeltaTable(title string, rows []SymDelta) {
	fmt.Println(title)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Symbol", "A", "B", "Delta"})

	for _, r := range rows {
		t.AppendRow(table.Row{r.Name, r.A, r.B, r.Delta})
	}
	t.Render()
	fmt.Println()
}
