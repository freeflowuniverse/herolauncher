#!/bin/bash

echo "Testing a more complex Jet template with multiple Jet features..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \"layout.jet\"}}
{{block bodyContent}}
  <h1>{{title}}</h1>
  <div class=\"content\">
    {{if len(items) > 0}}
      <ul>
        {{range items}}
          <li>{{.Name}} - {{if .Active}}Active{{else}}Inactive{{end}}</li>
        {{end}}
      </ul>
    {{else}}
      <p>No items found.</p>
    {{end}}
    
    {{include \"partials/footer.jet\" . }}
  </div>
{{end}}"
  }' | jq .

echo -e "\nTesting a complex template with syntax error..."
curl -X POST \
  http://localhost:9020/checkjet \
  -H 'Content-Type: application/json' \
  -d '{
    "template": "{{extends \"layout.jet\"}}
{{block bodyContent}}
  <h1>{{title}}</h1>
  <div class=\"content\">
    {{if len(items) > 0}}
      <ul>
        {{range items}}
          <li>{{.Name}} - {{if .Active}}Active{{else}}Inactive{{/end}}</li>
        {{end}}
      </ul>
    {{else}}
      <p>No items found.</p>
    {{end}}
  </div>
{{end}}"
  }' | jq .
