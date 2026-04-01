package model

import (
	"fmt"
	"strings"
)

var priceTypeOrder = []string{"hourly", "project", "negotiable"}

// NormalizePriceType parses comma-separated modes and returns a canonical string
// (hourly, project, negotiable in fixed order, deduplicated). Empty defaults to negotiable.
func NormalizePriceType(raw string) (string, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "negotiable", nil
	}
	seen := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if p != "hourly" && p != "project" && p != "negotiable" {
			return "", fmt.Errorf("invalid price_type token: %q", p)
		}
		seen[p] = struct{}{}
	}
	if len(seen) == 0 {
		return "negotiable", nil
	}
	var out []string
	for _, k := range priceTypeOrder {
		if _, ok := seen[k]; ok {
			out = append(out, k)
		}
	}
	return strings.Join(out, ","), nil
}
