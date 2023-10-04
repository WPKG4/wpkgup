package server

import (
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
)

const FilesTemplate = `<!DOCTYPE html>
<head>
<title>WpkgUp File Server | {{.path}}</title>
</head>
<body>
<h3>WpkgUp File Server</h3>
<h1>Directory listing for {{.path}}</h1>
<hr>
<ul>
{{ range $key, $value := .list }}
    <li><a href="{{ $value.Href }}">{{ $value.Name }}</a></li>
{{end}}
</ul>
<hr>
</body>`

func ProcessTemplate(name, templ string, data gin.H) string {
	var result strings.Builder
	template.Must(template.New(name).Parse(templ)).Execute(&result, data)
	return result.String()
}
