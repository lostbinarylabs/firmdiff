package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewExplainCmd returns the explain command for comparing two ELF files.
func NewExplainCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "explain <A.elf> <B.elf>",
		Short:   "Explain why firmware size changed",
		GroupID: "core",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {

			aPath := args[0]
			bPath := args[1]

			return runExplain(aPath, bPath)
		},
	}

	return cmd
}

func runExplain(aPath, bPath string) error {

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

	grown, shrunk := DiffSymbols(aRes.TopSyms, bRes.TopSyms, 10)

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
