package openrpcmanager

// OpenRPCSchema represents the OpenRPC specification document
// Based on OpenRPC Specification 1.2.6: https://spec.open-rpc.org/
type OpenRPCSchema struct {
	OpenRPC    string            `json:"openrpc"`              // Required: Version of the OpenRPC specification
	Info       InfoObject        `json:"info"`                 // Required: Information about the API
	Methods    []MethodObject    `json:"methods"`              // Required: List of method objects
	ExternalDocs *ExternalDocsObject `json:"externalDocs,omitempty"` // Optional: External documentation
	Servers    []ServerObject    `json:"servers,omitempty"`    // Optional: List of servers
	Components *ComponentsObject `json:"components,omitempty"` // Optional: Reusable components
}

// InfoObject provides metadata about the API
type InfoObject struct {
	Title          string        `json:"title"`                    // Required: Title of the API
	Description    string        `json:"description,omitempty"`    // Optional: Description of the API
	Version        string        `json:"version"`                  // Required: Version of the API
	TermsOfService string        `json:"termsOfService,omitempty"` // Optional: Terms of service URL
	Contact        *ContactObject `json:"contact,omitempty"`       // Optional: Contact information
	License        *LicenseObject `json:"license,omitempty"`       // Optional: License information
}

// ContactObject provides contact information for the API
type ContactObject struct {
	Name  string `json:"name,omitempty"`  // Optional: Name of the contact
	URL   string `json:"url,omitempty"`   // Optional: URL of the contact
	Email string `json:"email,omitempty"` // Optional: Email of the contact
}

// LicenseObject provides license information for the API
type LicenseObject struct {
	Name string `json:"name"`           // Required: Name of the license
	URL  string `json:"url,omitempty"`  // Optional: URL of the license
}

// ExternalDocsObject provides a URL to external documentation
type ExternalDocsObject struct {
	Description string `json:"description,omitempty"` // Optional: Description of the external docs
	URL         string `json:"url"`                   // Required: URL of the external docs
}

// ServerObject provides connection information to a server
type ServerObject struct {
	Name        string                    `json:"name,omitempty"`        // Optional: Name of the server
	Description string                    `json:"description,omitempty"` // Optional: Description of the server
	URL         string                    `json:"url"`                   // Required: URL of the server
	Variables   map[string]ServerVariable `json:"variables,omitempty"`   // Optional: Server variables
}

// ServerVariable is a variable for server URL template substitution
type ServerVariable struct {
	Default     string   `json:"default"`               // Required: Default value of the variable
	Description string   `json:"description,omitempty"` // Optional: Description of the variable
	Enum        []string `json:"enum,omitempty"`        // Optional: Enumeration of possible values
}

// MethodObject describes an RPC method
type MethodObject struct {
	Name        string                 `json:"name"`                  // Required: Name of the method
	Description string                 `json:"description,omitempty"` // Optional: Description of the method
	Summary     string                 `json:"summary,omitempty"`     // Optional: Summary of the method
	Params      []ContentDescriptorObject `json:"params"`            // Required: List of parameters
	Result      *ContentDescriptorObject  `json:"result"`            // Required: Description of the result
	Deprecated  bool                   `json:"deprecated,omitempty"`  // Optional: Whether the method is deprecated
	Errors      []ErrorObject          `json:"errors,omitempty"`      // Optional: List of possible errors
	Tags        []TagObject            `json:"tags,omitempty"`        // Optional: List of tags
	ExternalDocs *ExternalDocsObject   `json:"externalDocs,omitempty"` // Optional: External documentation
	ParamStructure string              `json:"paramStructure,omitempty"` // Optional: Structure of the parameters
}

// ContentDescriptorObject describes the content of a parameter or result
type ContentDescriptorObject struct {
	Name        string      `json:"name"`                  // Required: Name of the parameter
	Description string      `json:"description,omitempty"` // Optional: Description of the parameter
	Summary     string      `json:"summary,omitempty"`     // Optional: Summary of the parameter
	Required    bool        `json:"required,omitempty"`    // Optional: Whether the parameter is required
	Deprecated  bool        `json:"deprecated,omitempty"`  // Optional: Whether the parameter is deprecated
	Schema      SchemaObject `json:"schema"`               // Required: JSON Schema of the parameter
}

// SchemaObject is a JSON Schema definition
// This is a simplified version, in a real implementation you would use a full JSON Schema library
type SchemaObject map[string]interface{}

// ErrorObject describes an error that may be returned
type ErrorObject struct {
	Code    int         `json:"code"`                  // Required: Error code
	Message string      `json:"message"`               // Required: Error message
	Data    interface{} `json:"data,omitempty"`        // Optional: Additional error data
}

// TagObject describes a tag for documentation purposes
type TagObject struct {
	Name         string             `json:"name"`                   // Required: Name of the tag
	Description  string             `json:"description,omitempty"`  // Optional: Description of the tag
	ExternalDocs *ExternalDocsObject `json:"externalDocs,omitempty"` // Optional: External documentation
}

// ComponentsObject holds reusable objects for different aspects of the OpenRPC spec
type ComponentsObject struct {
	Schemas         map[string]SchemaObject          `json:"schemas,omitempty"`         // Optional: Reusable schemas
	ContentDescriptors map[string]ContentDescriptorObject `json:"contentDescriptors,omitempty"` // Optional: Reusable content descriptors
	Examples        map[string]interface{}           `json:"examples,omitempty"`        // Optional: Reusable examples
	Links           map[string]LinkObject            `json:"links,omitempty"`           // Optional: Reusable links
	Errors          map[string]ErrorObject           `json:"errors,omitempty"`          // Optional: Reusable errors
}

// LinkObject describes a link between operations
type LinkObject struct {
	Name        string                 `json:"name,omitempty"`        // Optional: Name of the link
	Description string                 `json:"description,omitempty"` // Optional: Description of the link
	Summary     string                 `json:"summary,omitempty"`     // Optional: Summary of the link
	Method      string                 `json:"method"`                // Required: Method name
	Params      map[string]interface{} `json:"params,omitempty"`      // Optional: Parameters for the method
	Server      *ServerObject          `json:"server,omitempty"`      // Optional: Server for the method
}
