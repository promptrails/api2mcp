// Command api2mcp is a standalone binary that serves an existing HTTP API as an
// MCP server, configured entirely from a YAML file — no Go code required.
//
//	api2mcp serve -config api2mcp.yaml
//
// Or the quick, config-less path for an OpenAPI spec:
//
//	api2mcp serve -openapi ./openapi.yaml -base https://api.internal -read-only -http :8080
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/promptrails/api2mcp"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "path to YAML config (overrides the flags below)")
	openapiPath := fs.String("openapi", "", "OpenAPI 3 spec path (quick path)")
	swaggoPath := fs.String("swaggo", "", "swaggo swagger.json path (quick path)")
	base := fs.String("base", "", "upstream API base URL")
	httpAddr := fs.String("http", "", "serve streamable-HTTP on this addr (e.g. :8080); default is stdio")
	readOnly := fs.Bool("read-only", false, "expose only read-only (GET/HEAD) operations")
	var includeTags multiFlag
	fs.Var(&includeTags, "include-tag", "only expose operations with this tag (repeatable)")
	_ = fs.Parse(os.Args[2:])

	cfg, err := resolveConfig(*configPath, *openapiPath, *swaggoPath, *base, *httpAddr, *readOnly, includeTags)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	src, err := cfg.buildSource()
	if err != nil {
		log.Fatalf("source: %v", err)
	}
	srv := api2mcp.New(src, cfg.options()...)

	ctx := context.Background()
	if cfg.Transport.Type == "http" {
		log.Printf("api2mcp: serving %q over streamable-HTTP on %s%s", cfg.Name, cfg.Transport.Addr, orDefault(cfg.Transport.Path, "/mcp"))
		if err := srv.ServeHTTP(ctx, cfg.Transport.Addr); err != nil {
			log.Fatalf("serve: %v", err)
		}
		return
	}
	log.Printf("api2mcp: serving %q over stdio", cfg.Name)
	if err := srv.ServeStdio(ctx); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// resolveConfig loads the YAML config when -config is given, otherwise builds a
// config from the quick-path flags.
func resolveConfig(configPath, openapiPath, swaggoPath, base, httpAddr string, readOnly bool, includeTags []string) (*Config, error) {
	if configPath != "" {
		return LoadConfig(configPath)
	}
	c := &Config{Name: "api2mcp", BaseURL: base}
	switch {
	case swaggoPath != "":
		c.Source.Type, c.Source.Path = "swaggo", swaggoPath
	case openapiPath != "":
		c.Source.Type, c.Source.Path = "openapi", openapiPath
	default:
		return nil, fmt.Errorf("need -config, or one of -openapi/-swaggo")
	}
	c.Curation.ReadOnly = readOnly
	c.Curation.IncludeTags = includeTags
	if httpAddr != "" {
		c.Transport.Type, c.Transport.Addr = "http", httpAddr
	} else {
		c.Transport.Type = "stdio"
	}
	return c, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func usage() {
	fmt.Fprintln(os.Stderr, `api2mcp — serve an existing HTTP API as an MCP server

usage:
  api2mcp serve -config api2mcp.yaml
  api2mcp serve -openapi spec.yaml -base https://api.internal -read-only [-http :8080]
  api2mcp serve -swaggo docs/swagger.json -base https://api.internal -include-tag public`)
}

// multiFlag collects a repeatable string flag.
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }
