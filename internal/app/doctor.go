package app

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func NewDoctorCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Check firmdiff dependencies and environment",
		GroupID: "utility",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show extra diagnostics")
	return cmd
}

type toolCheck struct {
	Label      string
	Candidates []string
	Required   bool
}

func runDoctor(verbose bool) error {
	fmt.Println("firmdiff doctor")
	fmt.Printf("Platform: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	checks := []toolCheck{
		{Label: "cmake", Candidates: []string{"cmake"}, Required: true},
		{Label: "ninja", Candidates: []string{"ninja"}, Required: false},
		{Label: "make", Candidates: []string{"make"}, Required: false},

		// Optional: docker backend
		{Label: "docker", Candidates: []string{"docker"}, Required: false},
	}

	var missing []string

	for _, c := range checks {
		path, chosen, err := findFirstInPath(c.Candidates)
		if err != nil {
			if c.Required {
				missing = append(missing, c.Label)
				fmt.Printf("✗ %-8s not found (tried: %s)\n", c.Label, strings.Join(c.Candidates, ", "))
			} else {
				fmt.Printf("• %-8s not found (optional)\n", c.Label)
			}
			continue
		}
		fmt.Printf("✓ %-8s %s\n", c.Label, path)
		if verbose && chosen != c.Label {
			fmt.Printf("  (using %q)\n", chosen)
		}
	}

	// Helpful macOS hint (for users who still use readelf/nm/size tools)
	if runtime.GOOS == "darwin" {
		fmt.Println()
		fmt.Println("macOS note: native builds produce Mach-O binaries (even if named *.elf).")
		fmt.Println("If you need ELF artifacts, build with an embedded toolchain or build inside Docker.")
		fmt.Println("If you use GNU binutils tools, install via: brew install binutils")
	}

	fmt.Println()

	if len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s", strings.Join(missing, ", "))
	}

	fmt.Println("All required tools are installed.")
	return nil
}

func findFirstInPath(candidates []string) (path string, chosen string, err error) {
	for _, name := range candidates {
		p, e := exec.LookPath(name)
		if e == nil {
			return p, name, nil
		}
	}
	return "", "", errors.New("not found")
}

// ---- Binary format detection helpers (used by run.go) ----

type BinaryFormat string

const (
	FormatELF     BinaryFormat = "ELF"
	FormatMachO   BinaryFormat = "Mach-O"
	FormatPE      BinaryFormat = "PE"
	FormatUnknown BinaryFormat = "Unknown"
)

func detectFormat(path string) BinaryFormat {
	if f, err := elf.Open(path); err == nil {
		_ = f.Close()
		return FormatELF
	}
	if f, err := macho.Open(path); err == nil {
		_ = f.Close()
		return FormatMachO
	}
	if f, err := pe.Open(path); err == nil {
		_ = f.Close()
		return FormatPE
	}
	return FormatUnknown
}

func printFormat(path, label string) error {
	format := detectFormat(path)
	switch format {
	case FormatELF:
		fmt.Printf("  %s: ELF ✅ (%s)\n", label, path)
		return nil
	case FormatMachO:
		return fmt.Errorf("%s artifact is Mach-O (macOS), not ELF: %s\nTip: build with an embedded toolchain or run builds in Docker", label, path)
	case FormatPE:
		return fmt.Errorf("%s artifact is Windows PE, not ELF: %s", label, path)
	default:
		return fmt.Errorf("%s artifact format unknown (expected ELF): %s", label, path)
	}
}
