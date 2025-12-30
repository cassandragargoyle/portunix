[{{.ProductName}}] Define acceptance criteria: {{.Title}}
---
Hello {{.UserName}},

Please define acceptance criteria for:

{{.Title}}
{{if .Description}}
{{.Description}}
{{end}}
To define acceptance criteria, reply to this email with:
- Given [context]
- When [action]
- Then [expected result]

Item ID: {{.ItemID}}

Regards,
Product Team
