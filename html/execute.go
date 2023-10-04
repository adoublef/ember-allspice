// https://charly3pins.dev/blog/learn-how-to-use-the-embed-package-in-go-by-building-a-web-page-easily/
package html

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"text/template"
)

var (
	//go:embed all:partials
	fsys embed.FS

	pm map[string]*template.Template
)

func init() {
	if pm == nil {
		pm = make(map[string]*template.Template)
	}

	ff, err := fs.ReadDir(fsys, "partials")
	if err != nil {
		panic(err)
	}

	// TODO
	// 1. traverse nested folders
	// 2. ignore files with _*.html pattern
	for _, f := range ff {
		if f.IsDir() {
			continue
		}

		pt, err := template.ParseFS(fsys, "partials/"+f.Name(), "partials/_*.html")
		if err != nil {
			panic(err)
		}

		// call without extension
		pm[f.Name()] = pt
	}

	// fmt.Println(pm)
}

func Execute(wr io.Writer, name string, data any) error {
	t, ok := pm[name]
	if !ok {
		return fmt.Errorf("partial with name %s not found", name)
	}

	err := t.Execute(wr, data)
	if err != nil {
		return fmt.Errorf("error writing to output: %w", err)
	}

	return nil
}
