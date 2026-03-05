package app

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print version info",
		GroupID: "utility",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("firmdiff %s\n", Version)
			fmt.Printf("commit: %s\n", GitCommit)
			fmt.Printf("built:  %s\n", BuildDate)
			fmt.Printf("go: %s\n", runtime.Version())
		},
	}
}
