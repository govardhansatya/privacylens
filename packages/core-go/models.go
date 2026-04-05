package privacylens

import "fmt"

type EntitySpan struct {
	Start      int
	End        int
	EntityType string
	Value      string
}

func NewEntitySpan(start, end int, entityType, value string) (EntitySpan, error) {
	if start > end {
		return EntitySpan{}, fmt.Errorf("EntitySpan.start (%d) must be <= end (%d)", start, end)
	}
	return EntitySpan{
		Start:      start,
		End:        end,
		EntityType: entityType,
		Value:      value,
	}, nil
}

type PatternConfig struct {
	EntityType string `json:"entityType" yaml:"entityType"`
	EntityTypeAlt string `json:"entity_type" yaml:"entity_type"`
	Pattern    string `json:"pattern" yaml:"pattern"`
}

func (p PatternConfig) Type() string {
	if p.EntityType != "" {
		return p.EntityType
	}
	return p.EntityTypeAlt
}

type DetectorConfig struct {
	Enabled  *bool          `json:"enabled" yaml:"enabled"`
	Patterns []PatternConfig `json:"patterns" yaml:"patterns"`
}

func (d DetectorConfig) IsEnabled() bool {
	if d.Enabled == nil {
		return true
	}
	return *d.Enabled
}

type Config struct {
	Version   string                    `json:"version" yaml:"version"`
	Detectors map[string]DetectorConfig `json:"detectors" yaml:"detectors"`
	Vault     string                    `json:"vault" yaml:"vault"`
	OnDetection func(entityType string) `json:"-" yaml:"-"`
}

type Detector interface {
	Detect(text string) []EntitySpan
}
