package privacylens

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenizeDetokenizeRoundTrip(t *testing.T) {
	vault := NewMemoryVault()
	text := "Email alice@example.com and phone 555-123-4567"
	spans := []EntitySpan{
		{Start: 6, End: 23, EntityType: "EMAIL", Value: "alice@example.com"},
		{Start: 34, End: 46, EntityType: "PHONE", Value: "555-123-4567"},
	}
	tokenized := Tokenize(text, spans, vault, "s1")
	if strings.Contains(tokenized.TokenizedText, "alice@example.com") || strings.Contains(tokenized.TokenizedText, "555-123-4567") {
		t.Fatalf("expected pii masked, got: %s", tokenized.TokenizedText)
	}
	restored := Detokenize(tokenized.TokenizedText, vault, "s1")
	if restored != text {
		t.Fatalf("expected round-trip restore %q, got %q", text, restored)
	}
}

func TestStableTokenAssignmentAndSessionIsolation(t *testing.T) {
	vault := NewMemoryVault()
	span := EntitySpan{Start: 0, End: 17, EntityType: "EMAIL", Value: "alice@example.com"}
	res1 := Tokenize("alice@example.com", []EntitySpan{span}, vault, "s1")
	res2 := Tokenize("alice@example.com", []EntitySpan{span}, vault, "s1")
	if len(res1.Pairs) == 0 || len(res2.Pairs) == 0 {
		t.Fatal("expected pairs")
	}
	if res1.Pairs[0].Token != res2.Pairs[0].Token {
		t.Fatalf("expected stable token in same session, got %s vs %s", res1.Pairs[0].Token, res2.Pairs[0].Token)
	}

	res3 := Tokenize("alice@example.com", []EntitySpan{span}, vault, "s2")
	if len(res3.Pairs) == 0 || res3.Pairs[0].Token != "[EMAIL_1]" {
		t.Fatalf("expected first token in separate session, got %+v", res3.Pairs)
	}
}

func TestOverlapResolutionLongestFirst(t *testing.T) {
	cfg := DefaultConfig()
	analyzer := NewAnalyzer([]Detector{
		detectorFn(func(_ string) []EntitySpan {
			return []EntitySpan{
				{Start: 0, End: 17, EntityType: "EMAIL", Value: "alice@example.com"},
				{Start: 0, End: 5, EntityType: "NAME", Value: "alice"},
			}
		}),
	}, &cfg)
	got := analyzer.Analyze("alice@example.com")
	if len(got) != 1 {
		t.Fatalf("expected 1 span after overlap resolution, got %d", len(got))
	}
	if got[0].EntityType != "EMAIL" {
		t.Fatalf("expected EMAIL span retained, got %s", got[0].EntityType)
	}
}

func TestRegexDetectorBuiltinsAndCustomPattern(t *testing.T) {
	custom := DetectorConfig{
		Patterns: []PatternConfig{
			{EntityType: "EMPLOYEE_ID", Pattern: `EMP-\d{4,}`},
		},
	}
	d := NewRegexDetector(&custom)
	text := "alice@example.com EMP-12345 123-45-6789"
	spans := d.Detect(text)
	types := map[string]bool{}
	for _, s := range spans {
		types[s.EntityType] = true
	}
	for _, want := range []string{"EMAIL", "SSN", "EMPLOYEE_ID"} {
		if !types[want] {
			t.Fatalf("expected to detect %s", want)
		}
	}
}

func TestLoadConfigDefaultsAndFile(t *testing.T) {
	cfg, err := LoadConfig(nil)
	if err != nil {
		t.Fatalf("load default config error: %v", err)
	}
	if cfg.Vault != "memory" {
		t.Fatalf("expected default memory vault, got %s", cfg.Vault)
	}
	if _, ok := cfg.Detectors["regex"]; !ok {
		t.Fatalf("expected default regex detector")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "privacylens.yaml")
	content := []byte("vault: sqlite\ndetectors:\n  regex:\n    enabled: false\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	cfg2, err := LoadConfig(&LoadConfigOptions{ConfigPath: path})
	if err != nil {
		t.Fatalf("load config file error: %v", err)
	}
	if cfg2.Vault != "sqlite" {
		t.Fatalf("expected sqlite vault from file, got %s", cfg2.Vault)
	}
	if cfg2.Detectors["regex"].IsEnabled() {
		t.Fatalf("expected regex disabled from file")
	}
}

func TestPipelineTokenizeAndDetokenizeResponse(t *testing.T) {
	p := NewPipeline(DefaultConfig())
	sessionID := "s1"
	msgs := []map[string]any{
		{"role": "user", "content": "My email is alice@example.com"},
	}
	masked := p.TokenizeMessages(msgs, sessionID)
	content := masked[0]["content"].(string)
	if strings.Contains(content, "alice@example.com") {
		t.Fatalf("expected masked content, got %s", content)
	}
	response := map[string]any{
		"choices": []any{
			map[string]any{
				"message": map[string]any{
					"content": "Email seen as [EMAIL_1]",
				},
			},
		},
	}
	restored := p.DetokenizeResponse(response, sessionID)
	choices := restored["choices"].([]any)
	msg := choices[0].(map[string]any)["message"].(map[string]any)
	got := msg["content"].(string)
	if !strings.Contains(got, "alice@example.com") {
		t.Fatalf("expected restored PII in response, got %s", got)
	}
}

type detectorFn func(text string) []EntitySpan

func (f detectorFn) Detect(text string) []EntitySpan { return f(text) }
