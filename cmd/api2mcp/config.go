package main

import (
	"fmt"
	"os"

	"github.com/promptrails/api2mcp"
	"github.com/promptrails/api2mcp/source"
	"github.com/promptrails/api2mcp/source/openapi"
	"github.com/promptrails/api2mcp/source/swaggo"
	"gopkg.in/yaml.v3"
)

// Config is the declarative form of an api2mcp server, loaded from YAML so a
// non-Go user can stand one up without writing code.
type Config struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	BaseURL string `yaml:"baseURL"`

	Source struct {
		Type string `yaml:"type"` // openapi | swaggo
		Path string `yaml:"path"`
		URL  string `yaml:"url"`
	} `yaml:"source"`

	Transport struct {
		Type string `yaml:"type"` // stdio | http
		Addr string `yaml:"addr"`
		Path string `yaml:"path"`
	} `yaml:"transport"`

	Curation struct {
		ReadOnly          bool     `yaml:"readOnly"`
		IncludeTags       []string `yaml:"includeTags"`
		ExcludeTags       []string `yaml:"excludeTags"`
		IncludePaths      []string `yaml:"includePaths"`
		ExcludePaths      []string `yaml:"excludePaths"`
		IncludeOperations []string `yaml:"includeOperations"`
		ExcludeOperations []string `yaml:"excludeOperations"`
	} `yaml:"curation"`

	ForwardHeaders   []string          `yaml:"forwardHeaders"`
	StaticHeaders    map[string]string `yaml:"staticHeaders"`
	MaxResponseBytes int               `yaml:"maxResponseBytes"`
	Audit            bool              `yaml:"audit"`
}

// LoadConfig reads and parses a YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &c, nil
}

// buildSource constructs the Source described by the config.
func (c *Config) buildSource() (source.Source, error) {
	switch c.Source.Type {
	case "", "openapi":
		if c.Source.URL != "" {
			return openapi.FromURL(c.Source.URL)
		}
		return openapi.FromFile(c.Source.Path)
	case "swaggo":
		return swaggo.FromFile(c.Source.Path)
	default:
		return nil, fmt.Errorf("unknown source type %q (want openapi|swaggo)", c.Source.Type)
	}
}

// options translates the config into api2mcp.Options.
func (c *Config) options() []api2mcp.Option {
	opts := []api2mcp.Option{api2mcp.WithBaseURL(c.BaseURL)}
	if c.Name != "" {
		opts = append(opts, api2mcp.WithName(c.Name))
	}
	if c.Version != "" {
		opts = append(opts, api2mcp.WithVersion(c.Version))
	}
	if c.Curation.ReadOnly {
		opts = append(opts, api2mcp.ReadOnly())
	}
	if len(c.Curation.IncludeTags) > 0 {
		opts = append(opts, api2mcp.IncludeTags(c.Curation.IncludeTags...))
	}
	if len(c.Curation.ExcludeTags) > 0 {
		opts = append(opts, api2mcp.ExcludeTags(c.Curation.ExcludeTags...))
	}
	if len(c.Curation.IncludePaths) > 0 {
		opts = append(opts, api2mcp.IncludePaths(c.Curation.IncludePaths...))
	}
	if len(c.Curation.ExcludePaths) > 0 {
		opts = append(opts, api2mcp.ExcludePaths(c.Curation.ExcludePaths...))
	}
	if len(c.Curation.IncludeOperations) > 0 {
		opts = append(opts, api2mcp.IncludeOperations(c.Curation.IncludeOperations...))
	}
	if len(c.Curation.ExcludeOperations) > 0 {
		opts = append(opts, api2mcp.ExcludeOperations(c.Curation.ExcludeOperations...))
	}
	if len(c.ForwardHeaders) > 0 {
		opts = append(opts, api2mcp.WithForwardHeaders(c.ForwardHeaders...))
	}
	for k, v := range c.StaticHeaders {
		opts = append(opts, api2mcp.WithStaticHeader(k, v))
	}
	if c.MaxResponseBytes > 0 {
		opts = append(opts, api2mcp.WithMaxResponseBytes(c.MaxResponseBytes))
	}
	if c.Audit {
		opts = append(opts, api2mcp.WithAuditLogger(api2mcp.StdAuditLogger))
	}
	if c.Transport.Path != "" {
		opts = append(opts, api2mcp.WithEndpointPath(c.Transport.Path))
	}
	return opts
}
