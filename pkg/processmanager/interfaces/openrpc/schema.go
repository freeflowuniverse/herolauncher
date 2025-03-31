package openrpc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
)

// LoadSchema loads the OpenRPC schema from the embedded JSON file
func LoadSchema() (openrpcmanager.OpenRPCSchema, error) {
	// Get the absolute path to the schema.json file
	_, filename, _, _ := runtime.Caller(0)
	schemaPath := filepath.Join(filepath.Dir(filename), "schema.json")

	// Read the schema file
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return openrpcmanager.OpenRPCSchema{}, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Unmarshal the schema
	var schema openrpcmanager.OpenRPCSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return openrpcmanager.OpenRPCSchema{}, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return schema, nil
}
