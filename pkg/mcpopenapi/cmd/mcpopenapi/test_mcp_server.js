#!/usr/bin/env node

/**
 * Test script for the MCP OpenAPI server
 * This script demonstrates how to interact with the MCP server programmatically
 * It tests both the hello tool and the OpenAPI validation tool
 */

const fs = require('fs');
const { spawn } = require('child_process');
const path = require('path');

// Path to the sample OpenAPI spec
const sampleSpecPath = path.join(__dirname, 'sample_openapi.json');

// Function to send a request to the MCP server
function sendRequest(server, request) {
  return new Promise((resolve) => {
    server.stdin.write(JSON.stringify(request) + '\n');
    
    // Set up a listener for the response
    const onData = (data) => {
      try {
        const response = JSON.parse(data.toString());
        if (response.id === request.id) {
          resolve(response);
        }
      } catch (error) {
        console.error('Error parsing response:', error);
      }
    };
    
    server.stdout.on('data', onData);
  });
}

// Main function to run the tests
async function runTests() {
  console.log('Starting MCP OpenAPI server tests...');
  
  // Start the MCP server
  console.log('Starting the MCP server...');
  const server = spawn('./mcpopenapi', [], {
    cwd: __dirname,
    stdio: ['pipe', 'pipe', process.stderr]
  });
  
  // Wait for the server to initialize
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  console.log('Server started. Running tests...');
  
  // Test discovery
  console.log('\n--- Testing Discovery ---');
  const discoveryRequest = {
    jsonrpc: '2.0',
    id: 1,
    method: 'tool.list'
  };
  
  const discoveryResponse = await sendRequest(server, discoveryRequest);
  console.log('Discovery response:');
  console.log(JSON.stringify(discoveryResponse, null, 2));
  
  // Test hello tool
  console.log('\n--- Testing Hello Tool ---');
  const helloRequest = {
    jsonrpc: '2.0',
    id: 2,
    method: 'tool.call',
    params: {
      name: 'hello',
      arguments: {
        submitter: 'World'
      }
    }
  };
  
  const helloResponse = await sendRequest(server, helloRequest);
  console.log('Hello tool response:');
  console.log(JSON.stringify(helloResponse, null, 2));
  
  // Test OpenAPI validation tool
  console.log('\n--- Testing OpenAPI Validation Tool ---');
  const openApiSpec = fs.readFileSync(sampleSpecPath, 'utf8');
  
  const validateRequest = {
    jsonrpc: '2.0',
    id: 3,
    method: 'tool.call',
    params: {
      name: 'validate_openapi',
      arguments: {
        spec: openApiSpec
      }
    }
  };
  
  const validateResponse = await sendRequest(server, validateRequest);
  console.log('OpenAPI validation response:');
  console.log(JSON.stringify(validateResponse, null, 2));
  
  // Clean up
  console.log('\nTests completed. Shutting down server...');
  server.kill();
  process.exit(0);
}

// Run the tests
runTests().catch(error => {
  console.error('Error running tests:', error);
  process.exit(1);
});
