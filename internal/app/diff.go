package app

import (
	"sort"
)

// SymDelta represents the difference in size between two symbols.
type SymDelta struct {
	Name  string
	A     int64
	B     int64
	Delta int64 // B - A
}

// DiffSymbols returns the biggest symbol size increases and decreases
// between A and B, limited to topN results per group.
func DiffSymbols(a, b []SymbolInfo, topN int) (grown []SymDelta, shrunk []SymDelta) {
	am := make(map[string]int64, len(a))
	bm := make(map[string]int64, len(b))

	for _, s := range a {
		am[s.Name] = s.Size
	}
	for _, s := range b {
		bm[s.Name] = s.Size
	}

	seen := make(map[string]struct{}, len(am)+len(bm))
	for k := range am {
		seen[k] = struct{}{}
	}
	for k := range bm {
		seen[k] = struct{}{}
	}

	var deltas []SymDelta
	for name := range seen {
		aa := am[name]
		bb := bm[name]
		d := bb - aa
		if d == 0 {
			continue
		}
		deltas = append(deltas, SymDelta{Name: name, A: aa, B: bb, Delta: d})
	}

	// Grown: biggest positive deltas
	sort.Slice(deltas, func(i, j int) bool {
		return deltas[i].Delta > deltas[j].Delta
	})

	for _, d := range deltas {
		if d.Delta > 0 {
			grown = append(grown, d)
		}
	}

	// Shrunk: most negative deltas
	sort.Slice(deltas, func(i, j int) bool {
		return deltas[i].Delta < deltas[j].Delta
	})

	for _, d := range deltas {
		if d.Delta < 0 {
			shrunk = append(shrunk, d)
		}
	}

	if topN <= 0 {
		topN = 10
	}
	if len(grown) > topN {
		grown = grown[:topN]
	}
	if len(shrunk) > topN {
		shrunk = shrunk[:topN]
	}
	return grown, shrunk
}
