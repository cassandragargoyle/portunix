[{{.ProductName}}] Vote request: {{.Title}}
---
Hello {{.UserName}},

We would like your opinion on this requirement:

{{.Title}}
{{if .Description}}
{{truncate .Description 200}}
{{end}}
Please vote here:
{{.FiderURL}}/posts/{{.PostNumber}}

Regards,
Product Team
