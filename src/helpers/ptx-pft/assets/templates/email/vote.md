[{{.ProductName}}] Vote request: {{.Title}}
---
Hello {{.UserName}},

We would like your opinion on this requirement:

{{.Title}}
{{if .Description}}
{{truncate .Description 200}}
{{end}}
To vote, reply to this email with:
  +1  = I support this requirement
  -1  = I do not support this requirement
   0  = I abstain

Item ID: {{.ItemID}}

Regards,
Product Team
