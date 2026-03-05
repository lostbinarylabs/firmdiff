package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "firmdiff",
		Short:         "Firmware build diff tool: build A, build B, compare outputs",
		Long:          "firmdiff builds two CMake configurations, compares artifacts, and reports size/symbol changes.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewDoctorCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

func Execute() error {
	if err := NewRootCmd().Execute(); err != nil {
		// cobra already formats usage; we keep errors clean
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	return nil
}
