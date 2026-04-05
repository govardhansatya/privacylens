package privacylens

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func boolPtr(v bool) *bool { return &v }

func DefaultConfig() Config {
	return Config{
		Version: "1",
		Detectors: map[string]DetectorConfig{
			"regex": {Enabled: boolPtr(true)},
		},
		Vault: "memory",
	}
}

type LoadConfigOptions struct {
	ConfigPath  string
	Overrides   *Config
	OnDetection func(entityType string)
}

func deepMergeConfig(base, override Config) Config {
	out := base
	if override.Version != "" {
		out.Version = override.Version
	}
	if override.Vault != "" {
		out.Vault = override.Vault
	}
	if override.Detectors != nil {
		if out.Detectors == nil {
			out.Detectors = map[string]DetectorConfig{}
		}
		for k, v := range override.Detectors {
			cur, ok := out.Detectors[k]
			if !ok {
				out.Detectors[k] = v
				continue
			}
			if v.Enabled != nil {
				cur.Enabled = v.Enabled
			}
			if len(v.Patterns) > 0 {
				cur.Patterns = v.Patterns
			}
			out.Detectors[k] = cur
		}
	}
	if override.OnDetection != nil {
		out.OnDetection = override.OnDetection
	}
	return out
}

func loadConfigFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(b, &cfg); err != nil {
			return Config{}, err
		}
	default:
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

func normalizeConfig(cfg Config) Config {
	out := cfg
	if out.Version == "" {
		out.Version = "1"
	}
	if out.Detectors == nil || len(out.Detectors) == 0 {
		out.Detectors = DefaultConfig().Detectors
	}
	switch out.Vault {
	case "memory", "redis", "sqlite":
	default:
		out.Vault = "memory"
	}
	return out
}

func LoadConfig(opts *LoadConfigOptions) (Config, error) {
	merged := DefaultConfig()

	// Priority order (lowest -> highest):
	// default config, cwd privacylens.yaml, explicit opts.ConfigPath, direct overrides.
	cwdPath := filepath.Join(".", "privacylens.yaml")
	if _, err := os.Stat(cwdPath); err == nil {
		fileCfg, err := loadConfigFile(cwdPath)
		if err != nil {
			return Config{}, err
		}
		merged = deepMergeConfig(merged, fileCfg)
	}

	if opts != nil && opts.ConfigPath != "" {
		if _, err := os.Stat(opts.ConfigPath); err != nil {
			return Config{}, fmt.Errorf("config file not found: %s", opts.ConfigPath)
		}
		fileCfg, err := loadConfigFile(opts.ConfigPath)
		if err != nil {
			return Config{}, err
		}
		merged = deepMergeConfig(merged, fileCfg)
	}

	if opts != nil && opts.Overrides != nil {
		merged = deepMergeConfig(merged, *opts.Overrides)
	}
	if opts != nil && opts.OnDetection != nil {
		merged.OnDetection = opts.OnDetection
	}
	return normalizeConfig(merged), nil
}
