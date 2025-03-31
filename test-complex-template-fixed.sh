#!/bin/bash

echo "Testing a more complex Jet template with multiple Jet features..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \\\"layout.jet\\\"}}\\n{{block bodyContent}}\\n  <h1>{{title}}</h1>\\n  <div class=\\\"content\\\">\\n    {{if len(items) > 0}}\\n      <ul>\\n        {{range items}}\\n          <li>{{.Name}} - {{if .Active}}Active{{else}}Inactive{{end}}</li>\\n        {{end}}\\n      </ul>\\n    {{else}}\\n      <p>No items found.</p>\\n    {{end}}\\n    \\n    {{include \\\"partials/footer.jet\\\" . }}\\n  </div>\\n{{end}}"
  }' | jq .

echo -e "\nTesting a complex template with syntax error..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \\\"layout.jet\\\"}}\\n{{block bodyContent}}\\n  <h1>{{title}}</h1>\\n  <div class=\\\"content\\\">\\n    {{if len(items) > 0}}\\n      <ul>\\n        {{range items}}\\n          <li>{{.Name}} - {{if .Active}}Active{{else}}Inactive{{/end}}</li>\\n        {{end}}\\n      </ul>\\n    {{else}}\\n      <p>No items found.</p>\\n    {{end}}\\n  </div>\\n{{end}}"
  }' | jq .
