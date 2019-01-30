package blog

import "text/template"

var (
	ConfigTOMLTemplate = template.Must(template.New("config-toml").Parse(`
languageCode = "en-us"
languageLang = "en"
defaultContentLanguage = "en"
title = "{{ .Spec.Title }}"

baseURL = "/" 
theme = "kiss"

[blackfriday]
hrefTargetBlank = true
`))

	PostTemplate = template.Must(template.New("post-md").Parse(`
---
title: {{ .Post.Title }}
date: 2019-01-29T14:53:18-08:00
draft: false
---

{{ .Post.Content }}
`))
)
