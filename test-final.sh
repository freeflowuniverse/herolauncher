#!/bin/bash

echo "Testing a simple valid Jet template..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "<h1>Hello {{name}}</h1>"
  }' | jq .

echo -e "\nTesting a template with an if statement..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{if active}}<p>User is active</p>{{else}}<p>User is inactive</p>{{end}}"
  }' | jq .

echo -e "\nTesting a template with a syntax error..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{if active}}<p>User is active</p>{{else}}<p>User is inactive</p>"
  }' | jq .
