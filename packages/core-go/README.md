# PrivacyLens Go SDK (core-go)

Transparent PII masking primitives for Go applications.

## Install

```bash
go get github.com/govardhansatya/privacylens/packages/core-go
```

## Quick Start

```go
package main

import (
	"fmt"
	privacylens "github.com/govardhansatya/privacylens/packages/core-go"
)

func main() {
	cfg, _ := privacylens.LoadConfig(nil)
	pipeline := privacylens.NewPipeline(cfg)

	sessionID := "s1"
	messages := []map[string]any{
		{"role": "user", "content": "My email is john@example.com"},
	}

	masked := pipeline.TokenizeMessages(messages, sessionID)
	fmt.Println(masked[0]["content"]) // "My email is [EMAIL_1]"

	restored := pipeline.Detokenize("Contact [EMAIL_1]", sessionID)
	fmt.Println(restored) // "Contact john@example.com"
}
```

## Public API

- `LoadConfig(opts *LoadConfigOptions) (Config, error)`
- `Inspect(text string, config *Config) ([]EntitySpan, error)`
- `NewPipeline(config Config) *Pipeline`
- `Tokenize(text string, spans []EntitySpan, vault SessionVault, sessionID string) TokenizeResult`
- `Detokenize(text string, vault SessionVault, sessionID string) string`

## Notes

- Current Go SDK scope provides the full core tokenize/detokenize pipeline.
- Adapter auto-detection for specific LLM clients is not included yet in `core-go`.
