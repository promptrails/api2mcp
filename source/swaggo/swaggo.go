// Package swaggo turns a swaggo-generated Swagger (OpenAPI 2.0) document into
// an api2mcp Source. swaggo (github.com/swaggo/swag) is the most common way Go
// services document their handlers via // @Summary, // @Param annotations; it
// emits docs/swagger.json in OpenAPI 2.0. This source converts that to v3 and
// reuses the standard OpenAPI operation extraction, so an annotated Gin/Echo
// app gets full input schemas without an adapter.
package swaggo

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/promptrails/api2mcp/source/openapi"
)

// FromFile loads a swaggo-generated swagger.json (OpenAPI 2.0) and returns a
// Source backed by its v3 conversion.
func FromFile(path string) (*openapi.Source, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read swagger %q: %w", path, err)
	}
	return FromData(data)
}

// FromData parses a swaggo swagger.json (OpenAPI 2.0) from raw bytes.
func FromData(data []byte) (*openapi.Source, error) {
	var doc2 openapi2.T
	if err := doc2.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("parse swagger 2.0: %w", err)
	}
	doc3, err := openapi2conv.ToV3(&doc2)
	if err != nil {
		return nil, fmt.Errorf("convert swagger 2.0 -> openapi 3: %w", err)
	}
	return openapi.FromV3Doc(doc3), nil
}
