package provider

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
	"github.com/go-go-golems/goja-git/pkg/gitjs"
)

const (
	PackageID  = "goja-git"
	ModuleName = "git"
)

type Config struct {
	AllowWrite bool `json:"allowWrite"`
}

var configSchema = json.RawMessage(`{
  "type": "object",
  "required": ["allowWrite"],
  "properties": {
    "allowWrite": {
      "type": "boolean",
      "const": true,
      "description": "Explicitly acknowledge that the git module can create and mutate repositories on disk. This is not a path sandbox."
    }
  },
  "additionalProperties": false
}`)

func Register(registry *providerapi.Registry) error {
	return registry.Package(PackageID, providerapi.Module{
		Name:         ModuleName,
		DefaultAs:    ModuleName,
		Description:  "Git repository automation backed by go-git and exposed as require(\"git\").",
		ConfigSchema: configSchema,
		New: func(ctx providerapi.ModuleContext) (require.ModuleLoader, error) {
			cfg, err := decodeConfig(ctx.Config)
			if err != nil {
				return nil, fmt.Errorf("goja-git provider config: %w", err)
			}
			if !cfg.AllowWrite {
				return nil, fmt.Errorf("goja-git provider requires config.allowWrite=true")
			}
			return gitjs.NewLoader(), nil
		},
	})
}

func decodeConfig(data json.RawMessage) (Config, error) {
	cfg := Config{}
	if len(data) > 0 && string(data) != "null" {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}
