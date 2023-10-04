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
	//go:embed all:partials all:layouts
	fsys embed.FS

	pm map[string]*template.Template
)

func init() {
	dir := "partials"

	// create map of partials
	// add to a global register
	if pm == nil {
		pm = make(map[string]*template.Template)
	}

	ff, err := fs.ReadDir(fsys, dir)
	if err != nil {
		panic(err)
	}

	for _, f := range ff {
		if f.IsDir() {
			continue
		}

		pt, err := template.ParseFS(fsys, dir+"/"+f.Name(), "layouts/*.html")
		if err != nil {
			panic(err)
		}

		// call without extension
		pm[f.Name()] = pt
	}

	fmt.Println(pm)
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
