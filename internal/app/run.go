package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// NewRunCmd returns the run command for building and comparing two variants.
func NewRunCmd() *cobra.Command {
	var f runFlags

	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Build variant A, build variant B, compare them, and print results",
		GroupID: "core",
		RunE: func(_ *cobra.Command, _ []string) error {
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
	srcAbs, err := filepath.Abs(f.Src)
	if err != nil {
		return fmt.Errorf("resolve src path: %w", err)
	}
	f.Src = srcAbs

	base := filepath.Join(".firmdiff", "runs", safeName(f.Name))
	outA := filepath.Join(base, "A")
	outB := filepath.Join(base, "B")

	if err := os.MkdirAll(outA, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(outB, 0o755); err != nil {
		return err
	}

	fmt.Printf("firmdiff run: %q\n", f.Name)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  src: %s\n", f.Src)
	fmt.Printf("  out: %s\n\n", base)

	gen, err := detectGenerator(f.Generator)
	if err != nil {
		return err
	}

	f.Generator = gen
	logLine(Info, "Using CMake generator: %s", gen)

	// Build A
	if err := cmakeConfigureBuild(f, outA, "A", f.ACFlags, f.ACXXFlags, f.ALDFlags, f.EnvA); err != nil {
		return err
	}

	// Build B
	if err := cmakeConfigureBuild(f, outB, "B", f.BCFlags, f.BCXXFlags, f.BLDFlags, f.EnvB); err != nil {
		return err
	}

	printSection("BUILD")

	logLine(Info, "A built successfully")
	logLine(Info, "B built successfully")

	artifactA := strings.ReplaceAll(f.ArtifactTemplate, "{out}", outA)
	artifactB := strings.ReplaceAll(f.ArtifactTemplate, "{out}", outB)

	printSection("ARTIFACTS")

	logLine(Info, "A: %s", artifactA)
	logLine(Info, "B: %s", artifactB)

	printSection("BINARY FORMAT")

	if err := printFormat(artifactA, "A"); err != nil {
		logLine(Fail, "%v", err)
		return err
	}

	if err := printFormat(artifactB, "B"); err != nil {
		logLine(Fail, "%v", err)
		return err
	}

	// Analyze ELF A & B
	logLine(Info, "Analyzing ELF artifacts (debug/elf)")
	aRes, err := AnalyzeELF(artifactA, 200) // grab more to improve diff quality
	if err != nil {
		return fmt.Errorf("analyze A: %w", err)
	}
	bRes, err := AnalyzeELF(artifactB, 200)
	if err != nil {
		return fmt.Errorf("analyze B: %w", err)
	}

	// Summary table
	renderSummaryTable(
		f.Name,
		"A", "B",
		aRes.Size.Flash, bRes.Size.Flash,
		aRes.Size.Ram, bRes.Size.Ram,
	)

	// Threshold gates
	dFlash := abs64(bRes.Size.Flash - aRes.Size.Flash)
	dRam := abs64(bRes.Size.Ram - aRes.Size.Ram)

	if f.MaxFlashDelta > 0 && dFlash > int64(f.MaxFlashDelta) {
		logLine(Fail, "FLASH delta %d exceeds max %d", dFlash, f.MaxFlashDelta)
		return fmt.Errorf("%w: flash delta %d > %d", ErrThreshold, dFlash, f.MaxFlashDelta)
	}
	if f.MaxRamDelta > 0 && dRam > int64(f.MaxRamDelta) {
		logLine(Fail, "RAM delta %d exceeds max %d", dRam, f.MaxRamDelta)
		return fmt.Errorf("%w: ram delta %d > %d", ErrThreshold, dRam, f.MaxRamDelta)
	}

	// Symbol deltas (best-effort; may be empty on stripped binaries)
	grown, shrunk := DiffSymbols(aRes.TopSyms, bRes.TopSyms, 12)

	if len(grown) == 0 && len(shrunk) == 0 {
		logLine(Warn, "No symbol deltas found (binary may be stripped, or symbols unavailable).")
		logLine(Info, "Tip: build with -g or CMAKE_BUILD_TYPE=RelWithDebInfo to improve symbol reporting.")
	} else {
		if len(grown) > 0 {
			renderSymbolDeltaTable("Top symbol growth", grown)
		}
		if len(shrunk) > 0 {
			renderSymbolDeltaTable("Top symbol shrink", shrunk)
		}
	}

	logLine(Info, "Run OK")
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
	if workdir != "" {
		cmd.Dir = workdir
	}

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

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func detectGenerator(requested string) (string, error) {

	// If the user explicitly asked for one, use it
	if requested != "" {
		return requested, nil
	}

	// Prefer Ninja (fastest)
	if _, err := exec.LookPath("ninja"); err == nil {
		return "Ninja", nil
	}

	// Fallback to Make
	if _, err := exec.LookPath("make"); err == nil {
		return "Unix Makefiles", nil
	}

	// Windows fallback
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("nmake"); err == nil {
			return "NMake Makefiles", nil
		}
	}

	return "", fmt.Errorf("no supported build tool found (install ninja or make)")
}
