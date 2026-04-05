package privacylens

import (
	"fmt"
	"sort"
	"strings"
)

type TokenPair struct {
	Token string
	Value string
}

type TokenizeResult struct {
	TokenizedText string
	Pairs         []TokenPair
}

func Tokenize(text string, spans []EntitySpan, vault SessionVault, sessionID string) TokenizeResult {
	if len(spans) == 0 {
		return TokenizeResult{TokenizedText: text, Pairs: nil}
	}
	sorted := append([]EntitySpan(nil), spans...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		li := sorted[i].End - sorted[i].Start
		lj := sorted[j].End - sorted[j].Start
		return li > lj
	})

	valueToToken := map[string]string{}
	getOrCreate := func(entityType, value string) string {
		if token, ok := valueToToken[value]; ok {
			return token
		}
		n := 1
		for {
			candidate := fmt.Sprintf("[%s_%d]", entityType, n)
			stored, err := vault.Retrieve(sessionID, candidate)
			if err != nil {
				token := candidate
				vault.Store(sessionID, token, value)
				valueToToken[value] = token
				return token
			}
			if stored == value {
				valueToToken[value] = candidate
				return candidate
			}
			n++
		}
	}

	var parts []string
	pairs := make([]TokenPair, 0)
	cursor := 0
	lastEnd := -1
	for _, span := range sorted {
		if span.Start < lastEnd && span.End <= lastEnd {
			continue
		}
		if span.Start < 0 || span.End > len(text) || span.Start > len(text) || span.End < span.Start {
			continue
		}
		parts = append(parts, text[cursor:span.Start])
		token := getOrCreate(span.EntityType, span.Value)
		parts = append(parts, token)
		pairs = append(pairs, TokenPair{Token: token, Value: span.Value})
		cursor = span.End
		lastEnd = span.End
	}
	parts = append(parts, text[cursor:])
	return TokenizeResult{
		TokenizedText: strings.Join(parts, ""),
		Pairs:         pairs,
	}
}
