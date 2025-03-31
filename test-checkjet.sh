#!/bin/bash

echo "Testing valid Jet template..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{range .items}}<p>{{.}}</p>{{end}}"
  }' | jq .

echo -e "\nTesting invalid Jet template (missing closing tag)..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{range .items}}<p>{{.}}</p>"
  }' | jq .
