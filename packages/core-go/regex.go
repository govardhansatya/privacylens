package privacylens

import "regexp"

type compiledPattern struct {
	entityType string
	re         *regexp.Regexp
}

type RegexDetector struct {
	patterns []compiledPattern
}

func NewRegexDetector(cfg *DetectorConfig) *RegexDetector {
	patterns := []compiledPattern{
		{entityType: "EMAIL", re: regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)},
		{entityType: "PHONE", re: regexp.MustCompile(`(?:\(\d{3}\)\s?\d{3}-\d{4}|\d{3}-\d{3}-\d{4}|\+1\d{10})`)},
		{entityType: "SSN", re: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)},
	}
	if cfg != nil {
		for _, p := range cfg.Patterns {
			typ := p.Type()
			if typ == "" || p.Pattern == "" {
				continue
			}
			re, err := regexp.Compile(p.Pattern)
			if err != nil {
				continue
			}
			patterns = append(patterns, compiledPattern{entityType: typ, re: re})
		}
	}
	return &RegexDetector{patterns: patterns}
}

func (r *RegexDetector) Detect(text string) []EntitySpan {
	spans := make([]EntitySpan, 0)
	for _, p := range r.patterns {
		locs := p.re.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			start, end := loc[0], loc[1]
			spans = append(spans, EntitySpan{
				Start:      start,
				End:        end,
				EntityType: p.entityType,
				Value:      text[start:end],
			})
		}
	}
	return spans
}
