package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type runFlags struct {
	Src       string
	Name      string
	Generator string
	Target    string

	ArtifactTemplate string

	ACFlags   string
	ACXXFlags string
	ALDFlags  string
	EnvA      []string

	BCFlags   string
	BCXXFlags string
	BLDFlags  string
	EnvB      []string

	CMakeDefs []string

	Keep bool

	MaxFlashDelta int
	MaxRamDelta   int
}

func NewRunCmd() *cobra.Command {
	var f runFlags

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Build variant A, build variant B, compare them, and print results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f)
		},
	}

	cmd.Flags().StringVar(&f.Src, "src", ".", "source directory (CMake -S)")
	cmd.Flags().StringVar(&f.Name, "name", "firmdiff-run", "run name used in output folder/report")
	cmd.Flags().StringVar(&f.Generator, "generator", "Ninja", "CMake generator (e.g. Ninja, Unix Makefiles)")
	cmd.Flags().StringVar(&f.Target, "target", "", "optional build target (cmake --build --target)")

	cmd.Flags().StringVar(&f.ArtifactTemplate, "artifact", "{out}/app.elf", "artifact path template (supports {out})")

	cmd.Flags().StringVar(&f.ACFlags, "a-cflags", "", "C flags for A (passed via -DCMAKE_C_FLAGS=...)")
	cmd.Flags().StringVar(&f.ACXXFlags, "a-cxxflags", "", "C++ flags for A")
	cmd.Flags().StringVar(&f.ALDFlags, "a-ldflags", "", "linker flags for A (passed via -DCMAKE_EXE_LINKER_FLAGS=...)")
	cmd.Flags().StringArrayVar(&f.EnvA, "env-a", nil, "extra env for A e.g. CC=/path/gcc (repeatable)")

	cmd.Flags().StringVar(&f.BCFlags, "b-cflags", "", "C flags for B (passed via -DCMAKE_C_FLAGS=...)")
	cmd.Flags().StringVar(&f.BCXXFlags, "b-cxxflags", "", "C++ flags for B")
	cmd.Flags().StringVar(&f.BLDFlags, "b-ldflags", "", "linker flags for B (passed via -DCMAKE_EXE_LINKER_FLAGS=...)")
	cmd.Flags().StringArrayVar(&f.EnvB, "env-b", nil, "extra env for B e.g. CC=/path/gcc (repeatable)")

	cmd.Flags().StringArrayVar(&f.CMakeDefs, "cmake-def", nil, "CMake -D definitions KEY=VALUE (repeatable)")

	cmd.Flags().BoolVar(&f.Keep, "keep", true, "keep run working directory (default keeps everything in MVP)")
	cmd.Flags().IntVar(&f.MaxFlashDelta, "max-flash-delta", 0, "fail (exit 2) if abs flash delta exceeds this many bytes (0 disables)")
	cmd.Flags().IntVar(&f.MaxRamDelta, "max-ram-delta", 0, "fail (exit 2) if abs RAM delta exceeds this many bytes (0 disables)")

	return cmd
}

func run(f runFlags) error {
	// Create a working dir
	base := filepath.Join(".firmdiff", "runs", safeName(f.Name))
	outA := filepath.Join(base, "A")
	outB := filepath.Join(base, "B")

	if err := os.MkdirAll(outA, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(outB, 0o755); err != nil {
		return err
	}

	fmt.Printf("firmdiff: run %q\n", f.Name)
	fmt.Printf("  src: %s\n", f.Src)
	fmt.Printf("  out: %s\n\n", base)

	// Build A
	if err := cmakeConfigureBuild(f, outA, "A", f.ACFlags, f.ACXXFlags, f.ALDFlags, f.EnvA); err != nil {
		return err
	}

	// Build B
	if err := cmakeConfigureBuild(f, outB, "B", f.BCFlags, f.BCXXFlags, f.BLDFlags, f.EnvB); err != nil {
		return err
	}

	artifactA := strings.ReplaceAll(f.ArtifactTemplate, "{out}", outA)
	artifactB := strings.ReplaceAll(f.ArtifactTemplate, "{out}", outB)

	fmt.Println("Artifacts:")
	fmt.Printf("  A: %s\n", artifactA)
	fmt.Printf("  B: %s\n", artifactB)

	// Format check (useful immediately)
	fmt.Println()
	fmt.Println("Binary format:")
	if err := printFormat(artifactA, "A"); err != nil {
		return err
	}
	if err := printFormat(artifactB, "B"); err != nil {
		return err
	}

	// TODO:
	// - parse ELF sizes/symbols (debug/elf)
	// - compare & report
	// - implement thresholds for flash/ram deltas

	fmt.Println()
	fmt.Println("Next: wire in ELF analysis + diff report.")
	return nil
}

func cmakeConfigureBuild(f runFlags, outDir, label, cflags, cxxflags, ldflags string, extraEnv []string) error {
	fmt.Printf("[%s] configure\n", label)

	args := []string{"-S", f.Src, "-B", outDir, "-G", f.Generator}
	for _, def := range f.CMakeDefs {
		args = append(args, "-D", def)
	}
	if cflags != "" {
		args = append(args, "-D", "CMAKE_C_FLAGS="+cflags)
	}
	if cxxflags != "" {
		args = append(args, "-D", "CMAKE_CXX_FLAGS="+cxxflags)
	}
	if ldflags != "" {
		args = append(args, "-D", "CMAKE_EXE_LINKER_FLAGS="+ldflags)
	}

	if err := runCmd("cmake", args, outDir, extraEnv); err != nil {
		return fmt.Errorf("[%s] cmake configure failed: %w", label, err)
	}

	fmt.Printf("[%s] build\n", label)
	buildArgs := []string{"--build", outDir}
	if f.Target != "" {
		buildArgs = append(buildArgs, "--target", f.Target)
	}
	if err := runCmd("cmake", buildArgs, outDir, extraEnv); err != nil {
		return fmt.Errorf("[%s] cmake build failed: %w", label, err)
	}
	return nil
}

func runCmd(bin string, args []string, workdir string, extraEnv []string) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workdir

	// Merge environment
	env := os.Environ()
	env = append(env, extraEnv...)
	cmd.Env = env

	return cmd.Run()
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "firmdiff-run"
	}
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}
