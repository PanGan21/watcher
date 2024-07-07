package main

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func matchPatterns(t string, pats []string) (bool, error) {
	for _, p := range pats {
		m, err := doublestar.Match(p, t)
		if err != nil {
			return false, fmt.Errorf("match(%v, %v): %w", p, t, err)
		}
		if m {
			return true, nil
		}
		if strings.HasPrefix(t, "./") {
			m, err = doublestar.Match(p, t[2:])
			if err != nil {
				return false, fmt.Errorf("match(%v, %v): %w", p, t[2:], err)
			}
			if m {
				return true, nil
			}
		}
	}
	return false, nil
}
