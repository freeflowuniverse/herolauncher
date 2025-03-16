package openapi

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// OpenAPISpec represents a parsed OpenAPI specification
type OpenAPISpec struct {
	Document v3.Document
	RawSpec  []byte
}

// ParseFromFile loads and parses an OpenAPI specification from a file
func ParseFromFile(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec file: %w", err)
	}

	return ParseFromBytes(data)
}

// ParseFromBytes parses an OpenAPI specification from a byte slice
func ParseFromBytes(data []byte) (*OpenAPISpec, error) {
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Note: Validation is handled internally during parsing

	// Get the high-level model
	highLevelModel, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to build V3 model: %v", errs[0])
	}

	return &OpenAPISpec{
		Document: highLevelModel.Model,
		RawSpec:  data,
	}, nil
}

// GetPaths returns all paths defined in the OpenAPI specification
func (s *OpenAPISpec) GetPaths() map[string]*v3.PathItem {
	result := make(map[string]*v3.PathItem)
	for pair := s.Document.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		result[pair.Key()] = pair.Value()
	}
	return result
}

// GetOperations returns all operations defined in the OpenAPI specification
// The returned map is keyed by "method:path"
func (s *OpenAPISpec) GetOperations() map[string]*v3.Operation {
	operations := make(map[string]*v3.Operation)

	for pair := s.Document.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()
		if pathItem.Get != nil {
			operations[fmt.Sprintf("GET:%s", path)] = pathItem.Get
		}
		if pathItem.Post != nil {
			operations[fmt.Sprintf("POST:%s", path)] = pathItem.Post
		}
		if pathItem.Put != nil {
			operations[fmt.Sprintf("PUT:%s", path)] = pathItem.Put
		}
		if pathItem.Delete != nil {
			operations[fmt.Sprintf("DELETE:%s", path)] = pathItem.Delete
		}
		if pathItem.Options != nil {
			operations[fmt.Sprintf("OPTIONS:%s", path)] = pathItem.Options
		}
		if pathItem.Head != nil {
			operations[fmt.Sprintf("HEAD:%s", path)] = pathItem.Head
		}
		if pathItem.Patch != nil {
			operations[fmt.Sprintf("PATCH:%s", path)] = pathItem.Patch
		}
		if pathItem.Trace != nil {
			operations[fmt.Sprintf("TRACE:%s", path)] = pathItem.Trace
		}
	}

	return operations
}

// GetExampleResponses returns example responses for an operation if available
func (s *OpenAPISpec) GetExampleResponses(operationID string) map[string]interface{} {
	examples := make(map[string]interface{})

	operations := s.GetOperations()
	for _, operation := range operations {
		if operation.OperationId == operationID {
			for pair := operation.Responses.Codes.First(); pair != nil; pair = pair.Next() {
				statusCode := pair.Key()
				response := pair.Value()
				if response.Content != nil {
					for contentPair := response.Content.First(); contentPair != nil; contentPair = contentPair.Next() {
						mediaType := contentPair.Key()
						mediaTypeObject := contentPair.Value()
						if mediaTypeObject.Example != nil {
							examples[fmt.Sprintf("%s:%s", statusCode, mediaType)] = mediaTypeObject.Example
						}
						if mediaTypeObject.Examples != nil {
							for examplePair := mediaTypeObject.Examples.First(); examplePair != nil; examplePair = examplePair.Next() {
								exampleName := examplePair.Key()
								example := examplePair.Value()
								if example.Value != nil {
									examples[fmt.Sprintf("%s:%s:%s", statusCode, mediaType, exampleName)] = example.Value
								}
							}
						}
					}
				}
			}
		}
	}

	return examples
}
