package privacylens

import (
	"log"
	"sort"
)

type Analyzer struct {
	detectors []Detector
	config    Config
}

func NewAnalyzer(detectors []Detector, cfg *Config) *Analyzer {
	c := DefaultConfig()
	if cfg != nil {
		c = *cfg
	}
	return &Analyzer{detectors: detectors, config: c}
}

func resolveOverlaps(spans []EntitySpan) []EntitySpan {
	if len(spans) == 0 {
		return nil
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
	result := make([]EntitySpan, 0, len(sorted))
	lastEnd := -1
	for _, span := range sorted {
		if span.Start >= lastEnd {
			result = append(result, span)
			lastEnd = span.End
		} else if span.End > lastEnd {
			result[len(result)-1] = span
			lastEnd = span.End
		}
	}
	return result
}

func (a *Analyzer) Analyze(text string) []EntitySpan {
	all := make([]EntitySpan, 0)
	for _, detector := range a.detectors {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[privacylens] detector panic recovered: %v", r)
				}
			}()
			all = append(all, detector.Detect(text)...)
		}()
	}
	resolved := resolveOverlaps(all)
	if a.config.OnDetection != nil {
		for _, span := range resolved {
			a.config.OnDetection(span.EntityType)
		}
	}
	if len(resolved) > 0 {
		types := make([]string, 0, len(resolved))
		for _, s := range resolved {
			types = append(types, s.EntityType)
		}
		log.Printf("[privacylens] Detected %d entities: %v", len(resolved), types)
	}
	return resolved
}
