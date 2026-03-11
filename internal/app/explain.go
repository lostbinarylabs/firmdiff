package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewExplainCmd returns the explain command for comparing two ELF files.
func NewExplainCmd() *cobra.Command {
	var top int

	cmd := &cobra.Command{
		Use:     "explain <A.elf> <B.elf>",
		Short:   "Explain why firmware size changed",
		GroupID: "core",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExplain(args[0], args[1], top)
		},
	}

	cmd.Flags().IntVar(&top, "top", 10, "number of symbols to show in growth/shrink tables")

	return cmd
}

func runExplain(aPath, bPath string, top int) error {

	aRes, err := AnalyzeELF(aPath, 500)
	if err != nil {
		return err
	}

	bRes, err := AnalyzeELF(bPath, 500)
	if err != nil {
		return err
	}

	deltaFlash := bRes.Size.Flash - aRes.Size.Flash

	fmt.Println()
	fmt.Printf("FLASH delta: %+d bytes\n", deltaFlash)
	fmt.Println()

	grown, shrunk := DiffSymbols(aRes.TopSyms, bRes.TopSyms, top)

	if len(grown) > 0 {
		fmt.Println("Top causes of growth")
		for _, g := range grown {
			fmt.Printf("+%-6d %s\n", g.Delta, g.Name)
		}
		fmt.Println()
	}

	if len(shrunk) > 0 {
		fmt.Println("Top shrink")
		for _, s := range shrunk {
			fmt.Printf("%-7d %s\n", s.Delta, s.Name)
		}
		fmt.Println()
	}

	return nil
}
