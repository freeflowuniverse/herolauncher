package openapi

import (
	"encoding/json"
	"fmt"
)

// Example demonstrates how to use the OpenAPI package
func Example() {
	// Example OpenAPI spec as a string
	specJSON := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Example API",
			"version": "1.0.0"
		},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"summary": "Get all users",
					"responses": {
						"200": {
							"description": "Successful response",
							"content": {
								"application/json": {
									"example": [
										{
											"id": 1,
											"name": "John Doe",
											"email": "john@example.com"
										},
										{
											"id": 2,
											"name": "Jane Smith",
											"email": "jane@example.com"
										}
									]
								}
							}
						}
					}
				},
				"post": {
					"operationId": "createUser",
					"summary": "Create a new user",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {
											"type": "string"
										},
										"email": {
											"type": "string"
										}
									}
								}
							}
						}
					},
					"responses": {
						"201": {
							"description": "User created",
							"content": {
								"application/json": {
									"example": {
										"id": 3,
										"name": "New User",
										"email": "new@example.com"
									}
								}
							}
						}
					}
				}
			},
			"/users/{userId}": {
				"get": {
					"operationId": "getUserById",
					"summary": "Get user by ID",
					"parameters": [
						{
							"name": "userId",
							"in": "path",
							"required": true,
							"schema": {
								"type": "integer"
							}
						}
					],
					"responses": {
						"200": {
							"description": "Successful response",
							"content": {
								"application/json": {
									"example": {
										"id": 1,
										"name": "John Doe",
										"email": "john@example.com"
									}
								}
							}
						},
						"404": {
							"description": "User not found"
						}
					}
				}
			}
		}
	}`

	// Parse the OpenAPI spec
	spec, err := ParseFromBytes([]byte(specJSON))
	if err != nil {
		fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
		return
	}

	// Print the paths
	fmt.Println("Paths in the OpenAPI spec:")
	for path := range spec.GetPaths() {
		fmt.Printf("- %s\n", path)
	}

	// Print the operations
	fmt.Println("\nOperations in the OpenAPI spec:")
	for operationKey, operation := range spec.GetOperations() {
		operationID := operation.OperationId
		if operationID == "" {
			operationID = "unknown"
		}
		fmt.Printf("- %s (OperationID: %s)\n", operationKey, operationID)
	}

	// Print example responses
	fmt.Println("\nExample responses:")
	for operationKey, operation := range spec.GetOperations() {
		if operation.OperationId != "" {
			examples := spec.GetExampleResponses(operation.OperationId)
			if len(examples) > 0 {
				fmt.Printf("- %s (%s):\n", operationKey, operation.OperationId)
				for exampleKey, example := range examples {
					exampleJSON, _ := json.MarshalIndent(example, "  ", "  ")
					fmt.Printf("  - %s: %s\n", exampleKey, string(exampleJSON))
				}
			}
		}
	}

	// Create a server generator
	generator := NewServerGenerator(spec)
	
	// Generate server code
	serverCode := generator.GenerateServerCode()
	fmt.Println("\nGenerated server code (excerpt):")
	fmt.Println(serverCode[:500] + "...")
}
