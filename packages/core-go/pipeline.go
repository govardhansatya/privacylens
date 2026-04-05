package privacylens

import "fmt"

type Pipeline struct {
	config   Config
	vault    SessionVault
	analyzer *Analyzer
}

func buildDetectors(config Config) []Detector {
	if len(config.Detectors) == 0 {
		return []Detector{NewRegexDetector(nil)}
	}
	out := make([]Detector, 0)
	for name, detCfg := range config.Detectors {
		if !detCfg.IsEnabled() {
			continue
		}
		if name == "regex" {
			cfg := detCfg
			out = append(out, NewRegexDetector(&cfg))
		}
	}
	if len(out) == 0 {
		out = append(out, NewRegexDetector(nil))
	}
	return out
}

func NewPipeline(config Config) *Pipeline {
	vault := NewMemoryVault()
	return &Pipeline{
		config:   config,
		vault:    vault,
		analyzer: NewAnalyzer(buildDetectors(config), &config),
	}
}

func (p *Pipeline) TokenizeMessages(messages []map[string]any, sessionID string) []map[string]any {
	res := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		out := map[string]any{}
		for k, v := range msg {
			out[k] = v
		}
		content, ok := out["content"].(string)
		if !ok {
			res = append(res, out)
			continue
		}
		spans := p.analyzer.Analyze(content)
		tokenized := Tokenize(content, spans, p.vault, sessionID)
		out["content"] = tokenized.TokenizedText
		res = append(res, out)
	}
	return res
}

func (p *Pipeline) Detokenize(text, sessionID string) string {
	return Detokenize(text, p.vault, sessionID)
}

func (p *Pipeline) DetokenizeResponse(response map[string]any, sessionID string) map[string]any {
	choicesRaw, ok := response["choices"]
	if !ok {
		return response
	}
	choices, ok := choicesRaw.([]any)
	if !ok {
		return response
	}
	for _, choiceRaw := range choices {
		choice, ok := choiceRaw.(map[string]any)
		if !ok {
			continue
		}
		msgRaw, ok := choice["message"]
		if !ok {
			continue
		}
		msg, ok := msgRaw.(map[string]any)
		if !ok {
			continue
		}
		content, ok := msg["content"].(string)
		if !ok {
			continue
		}
		msg["content"] = p.Detokenize(content, sessionID)
	}
	return response
}

type ShieldResult struct {
	Pipeline *Pipeline
	Messages []map[string]any
}

func Shield(messages []map[string]any, config *Config) (ShieldResult, error) {
	var cfg Config
	var err error
	if config != nil {
		cfg = *config
		cfg = normalizeConfig(cfg)
	} else {
		cfg, err = LoadConfig(nil)
		if err != nil {
			return ShieldResult{}, err
		}
	}
	pipeline := NewPipeline(cfg)
	sessionID := "session"
	masked := pipeline.TokenizeMessages(messages, sessionID)
	return ShieldResult{
		Pipeline: pipeline,
		Messages: masked,
	}, nil
}

func Inspect(text string, config *Config) ([]EntitySpan, error) {
	var cfg Config
	if config == nil {
		loaded, err := LoadConfig(nil)
		if err != nil {
			return nil, err
		}
		cfg = loaded
	} else {
		cfg = normalizeConfig(*config)
	}
	analyzer := NewAnalyzer(buildDetectors(cfg), &cfg)
	return analyzer.Analyze(text), nil
}

func ShieldClient[T any](_ T, _ *Config) (T, error) {
	var zero T
	return zero, fmt.Errorf("shield(client) adapter auto-detection is not implemented in core-go yet")
}
