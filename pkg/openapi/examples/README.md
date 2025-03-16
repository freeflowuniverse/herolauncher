# OpenAPI Examples

This directory contains examples of using the OpenAPI package to generate server code from OpenAPI specifications.

## Structure

- `main.go`: Main server implementation that hosts multiple APIs and Swagger UI
- `petstore.yaml`: OpenAPI specification for the Petstore API
- `actions.yaml`: OpenAPI specification for the Actions API
- `petstoreapi/`: Generated code for the Petstore API
- `actionsapi/`: Generated code for the Actions API
- `test/`: Tests for the OpenAPI code generation
- `run_test.sh`: Script to run tests and start the server

## Running the Example

To run the example, execute the `run_test.sh` script:

```bash
./run_test.sh
```

This will:
1. Run the tests to generate server code for both APIs
2. Start a server on port 9091 that hosts both APIs and Swagger UI

## Accessing the APIs

Once the server is running, you can access the following endpoints:

- API Home: http://localhost:9091/api
- Petstore API: http://localhost:9091/api/petstore
- Petstore API Documentation: http://localhost:9091/api/swagger/petstore
- Actions API: http://localhost:9091/api/actions
- Actions API Documentation: http://localhost:9091/api/swagger/actions

## Features

This example demonstrates the following features:

1. **OpenAPI Code Generation**: Generating server code from OpenAPI specifications
2. **Multiple API Hosting**: Hosting multiple APIs under a single server
3. **Swagger UI Integration**: Providing API documentation with Swagger UI
4. **Mock Implementations**: Creating mock implementations based on examples in the OpenAPI spec

## Implementation Details

The server is implemented using the [Fiber](https://github.com/gofiber/fiber) web framework. Each API is mounted as a sub-app under the main server, allowing for clean separation of concerns.

The Swagger UI is implemented using the [swagger-ui-dist](https://www.npmjs.com/package/swagger-ui-dist) package, which is loaded from a CDN.

The OpenAPI specifications are parsed using the [libopenapi](https://github.com/pb33f/libopenapi) package, which provides a high-level API for working with OpenAPI specifications.
