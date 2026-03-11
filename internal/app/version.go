package app

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewVersionCmd returns the version command for printing firmdiff version info.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print version info",
		GroupID: "utility",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("firmdiff %s\n", Version)
			fmt.Printf("commit: %s\n", GitCommit)
			fmt.Printf("built:  %s\n", BuildDate)
			fmt.Printf("go: %s\n", runtime.Version())
		},
	}
}
