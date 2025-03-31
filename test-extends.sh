#!/bin/bash

echo "Testing a template with extends (should now be valid)..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \"layout.jet\"}}{{block bodyContent}}<h1>Hello World</h1>{{end}}"
  }' | jq .

echo -e "\nTesting a template with include (should now be valid)..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "<div>{{include \"partials/header.jet\" .}}<h1>Main Content</h1></div>"
  }' | jq .

echo -e "\nTesting a template with extends but having syntax error..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \"layout.jet\"}}{{block bodyContent}}<h1>Hello World</h1>{{else}}Not Found{{end}}"
  }' | jq .
