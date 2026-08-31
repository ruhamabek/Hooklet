package web

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/* static/*
var FS embed.FS

var (
 	DashboardTemplate *template.Template

 	IndexHTML []byte
)

func init() {
	var err error
	DashboardTemplate, err = template.ParseFS(FS,
		"templates/index.html",
		"templates/components/*.html",
		"templates/modals/*.html",
	)
	if err != nil {
		panic("web: failed to parse templates: " + err.Error())
	}

	var buf bytes.Buffer
	if err := DashboardTemplate.Execute(&buf, nil); err == nil {
		IndexHTML = buf.Bytes()
	}
}

 func StaticHandler() http.Handler {
	sub, err := fs.Sub(FS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}