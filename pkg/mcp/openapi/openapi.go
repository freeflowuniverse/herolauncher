package mcpopenapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
)

// ValidateOpenAPISpec validates an OpenAPI specification from a file path and returns any errors found
func ValidateOpenAPISpec(filePath string) (string, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read the file
	specContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}

	document, err := libopenapi.NewDocument(specContent)
	if err != nil {
		return "", fmt.Errorf("cannot create new document: %w", err)
	}

	docModel, errors := document.BuildV3Model()
	if len(errors) > 0 {
		var errorMessages []string
		for i := range errors {
			errorMessages = append(errorMessages, errors[i].Error())
		}
		return strings.Join(errorMessages, "\n"), nil
	}

	// Validate schemas
	var schemaInfo []string
	
	// Check if Components and Schemas are not nil
	if docModel != nil && docModel.Model.Components != nil && docModel.Model.Components.Schemas != nil {
		for schemaPairs := docModel.Model.Components.Schemas.First(); schemaPairs != nil; schemaPairs = schemaPairs.Next() {
			schemaName := schemaPairs.Key()
			schema := schemaPairs.Value()
			
			// Check if schema and its properties are not nil
			if schema != nil && schema.Schema() != nil && schema.Schema().Properties != nil {
				schemaInfo = append(schemaInfo, fmt.Sprintf("Schema '%s' has %d properties", 
					schemaName, schema.Schema().Properties.Len()))
			} else {
				schemaInfo = append(schemaInfo, fmt.Sprintf("Schema '%s' has no properties or is invalid", schemaName))
			}
		}
	} else {
		schemaInfo = append(schemaInfo, "No schemas found in the OpenAPI specification")
	}

	return strings.Join(schemaInfo, "\n"), nil
}
