package main

import "html/template"

type view struct {
	layout *template.Template
	dev    bool
}

// newView initializes the view layer
func newView(dev bool) view {
	return view{
		layout: template.Must(template.New("layout").Parse(FSMustString(dev, "/tmpls/_layout.go.html"))),
		dev:    dev,
	}
}

func (v *view) Tmpl(name string) *template.Template {
	l := template.Must(v.layout.Clone())
	n := "/tmpls/" + name + ".go.html"
	return template.Must(l.Parse(FSMustString(v.dev, n)))
}
