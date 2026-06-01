package api2mcp

import (
	"log"

	"github.com/promptrails/api2mcp/engine"
)

// StdAuditLogger logs each tool call to the standard logger in a single line:
//
//	api2mcp: getUser GET /users/{id} -> 200 (12ms)
//
// Pass it to WithAuditLogger for out-of-the-box observability.
func StdAuditLogger(e engine.AuditEvent) {
	if e.Err != nil {
		log.Printf("api2mcp: %s %s %s -> error: %v (%s)", e.OperationID, e.Method, e.Path, e.Err, e.Duration.Round(1e6))
		return
	}
	log.Printf("api2mcp: %s %s %s -> %d (%s)", e.OperationID, e.Method, e.Path, e.Status, e.Duration.Round(1e6))
}
