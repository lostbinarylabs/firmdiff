package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

// NewRootCmd returns the root command for the CLI application.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "firmdiff",
		Short:         "Firmware build diff tool: build A, build B, compare outputs",
		Long:          "firmdiff builds two CMake configurations, compares artifacts, and reports size/symbol changes.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddGroup(
		&cobra.Group{
			ID:    "core",
			Title: "Core Commands",
		},
		&cobra.Group{
			ID:    "utility",
			Title: "Utility Commands",
		},
	)

	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewExplainCmd())
	cmd.AddCommand(NewDoctorCmd())
	cmd.AddCommand(NewVersionCmd())

	// Hide cobra's auto-generated completion command
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return cmd
}

var ErrThreshold = errors.New("threshold exceeded")

// ExitCode returns the exit code for the given error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrThreshold) {
		return 2
	}
	return 1
}

// Execute runs the root command of the CLI application and handles any errors by printing them to stderr.
func Execute() error {
	if err := NewRootCmd().Execute(); err != nil {
		// cobra already formats usage; we keep errors clean
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	return nil
}
