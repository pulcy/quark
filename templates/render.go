package templates

import (
	"bytes"
	"strconv"
	"text/template"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

func Render(templateName string, options interface{}) (string, error) {
	asset, err := Asset(templateName)
	if err != nil {
		return "", maskAny(err)
	}

	// parse template
	var tmpl *template.Template
	tmpl = template.New(templateName)
	funcMap := template.FuncMap{
		"escape": escape,
		"quote":  strconv.Quote,
	}
	tmpl.Funcs(funcMap)
	_, err = tmpl.Parse(string(asset))
	if err != nil {
		return "", maskAny(err)
	}
	// write file to buffer
	tmpl.Funcs(funcMap)
	buffer := &bytes.Buffer{}
	err = tmpl.Execute(buffer, options)
	if err != nil {
		return "", maskAny(err)
	}

	return buffer.String(), nil
}

func escape(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}
