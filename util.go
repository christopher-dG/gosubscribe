package gosubscribe

import (
	"fmt"
	"sort"
	"strings"
)

// FormatCounts converts a mapper -> subscriber count mapping to an evenly spaced
// table, ordering the counts in descending order.
func FormatCounts(counts map[*Mapper]uint) string {
	// First, get the maximum width a mapper's name for formatting.
	maxWidth := -1
	for mapper := range counts {
		if len(mapper.Username) > maxWidth {
			maxWidth = len(mapper.Username)
		}
	}
	maxWidth += 5 // Some padding.

	var s string

	// Now add each line to the output, ordered by count (descending).
	for len(counts) > 0 {
		maxSubs := uint(0)
		maxSubsMapper := new(Mapper)
		for mapper, count := range counts {
			if count >= maxSubs {
				maxSubs = count
				maxSubsMapper = mapper
			}
		}
		padding := strings.Repeat(" ", maxWidth-len(maxSubsMapper.Username))
		var plural string
		if maxSubs != 1 {
			plural = "s"
		}
		s += fmt.Sprintf(
			"%s%s%d subscriber%s\n",
			maxSubsMapper.Username, padding, maxSubs, plural,
		)
		delete(counts, maxSubsMapper)
	}

	return s
}

// HasMapper determines whether or not the map contains a mapper key with the given name.
func HasMapper(mappers map[*Mapper]uint, key string) bool {
	for mapper := range mappers {
		if strings.EqualFold(mapper.Username, key) {
			return true
		}
	}
	return false
}

// HasMapset determines whether or not the given key is contained in a list.
func HasMapset(mapsets []*Mapset, key *Mapset) bool {
	for _, mapset := range mapsets {
		if mapset.ID == key.ID {
			return true
		}
	}
	return false
}

// GetTokens splits a comma-delimited string into tokens and returns the unique ones in
// sorted order.
func GetTokens(input string) []string {
	uniq := []string{}
	for _, s := range strings.Split(input, ", ") {
		s = strings.TrimSpace(s)
		contains := false
		for _, t := range uniq {
			if strings.EqualFold(s, t) {
				contains = true
				break
			}
		}
		if !contains {
			uniq = append(uniq, s)
		}
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i] < uniq[j] })
	return uniq
}
