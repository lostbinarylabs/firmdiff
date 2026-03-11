package app

import (
	"debug/elf"
	"fmt"
	"sort"
	"strings"
)

// SizeInfo represents the size of an ELF file.
type SizeInfo struct {
	Flash int64 // approx: allocated + !write sections (SHT_NOBITS excluded)
	Ram   int64 // approx: allocated + writable + NOBITS (bss)
	Text  int64 // optional breakdown
	Data  int64
	Bss   int64
}

// SymbolInfo represents information about a symbol, including its name and size.
type SymbolInfo struct {
	Name string
	Size int64
}

// AnalyzeResult represents the result of analyzing an ELF file.
type AnalyzeResult struct {
	Size    SizeInfo
	TopSyms []SymbolInfo // largest symbols (best-effort)
}

// AnalyzeELF analyzes an ELF file at the given path and returns its size and top symbols.
func AnalyzeELF(path string, topN int) (AnalyzeResult, error) {
	f, err := elf.Open(path)
	if err != nil {
		return AnalyzeResult{}, err
	}
	defer func() {
		_ = f.Close()
	}()

	size := computeSizeFromELF(f)
	syms := extractTopSymbols(f, topN)

	return AnalyzeResult{
		Size:    size,
		TopSyms: syms,
	}, nil
}

func computeSizeFromELF(f *elf.File) SizeInfo {
	var text, data, bss int64

	for _, s := range f.Sections {
		// Only sections that occupy memory at runtime.
		if s.Flags&elf.SHF_ALLOC == 0 {
			continue
		}

		sz := int64(s.Size)

		// NOBITS take RAM but not FLASH (e.g. .bss)
		if s.Type == elf.SHT_NOBITS {
			bss += sz
			continue
		}

		// Writable allocated sections are typically RAM (.data)
		if s.Flags&elf.SHF_WRITE != 0 {
			data += sz
		} else {
			// Executable or read-only allocated sections typically FLASH (.text/.rodata)
			text += sz
		}
	}

	return SizeInfo{
		Flash: text,       // approximate FLASH usage
		Ram:   data + bss, // approximate RAM usage
		Text:  text,
		Data:  data,
		Bss:   bss,
	}
}

func extractTopSymbols(f *elf.File, topN int) []SymbolInfo {
	if topN <= 0 {
		topN = 20
	}

	var all []elf.Symbol

	// .symtab (may be missing if stripped)
	if syms, err := f.Symbols(); err == nil {
		all = append(all, syms...)
	}
	// .dynsym (often present)
	if dsyms, err := f.DynamicSymbols(); err == nil {
		all = append(all, dsyms...)
	}

	// Deduplicate by name, keep the largest size
	seen := make(map[string]int64, len(all))
	for _, s := range all {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			continue
		}
		if s.Section == elf.SHN_UNDEF {
			continue
		}
		if s.Size == 0 {
			continue
		}

		// Skip some noise (optional)
		if strings.HasPrefix(name, ".") {
			continue
		}

		sz := int64(s.Size)
		if prev, ok := seen[name]; ok && prev >= sz {
			continue
		}
		seen[name] = sz
	}

	out := make([]SymbolInfo, 0, len(seen))
	for name, sz := range seen {
		out = append(out, SymbolInfo{Name: name, Size: sz})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Size > out[j].Size })

	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func (s SizeInfo) String() string {
	return fmt.Sprintf("FLASH=%d RAM=%d (text=%d data=%d bss=%d)", s.Flash, s.Ram, s.Text, s.Data, s.Bss)
}
