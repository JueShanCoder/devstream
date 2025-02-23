//go:build ignore
// +build ignore

// This program is run via "go generate" to generate the code.

package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Option struct {
	Plugins []string
	Dir     string
	Path    string
	Package string
	Funcs   template.FuncMap
}

func main() {
	srcDir := flag.String("dir", "plugins", "source directory for yaml files")
	packageName := flag.String("pkg", "config", "package name")
	flag.Parse()

	plugins, err := getYamlFiles(*srcDir)
	if err != nil {
		log.Fatal(err)
	}

	generate(&Option{
		Plugins: plugins,
		Dir:     *srcDir,
		Path:    "embed_gen.go",
		Package: *packageName,
		Funcs: template.FuncMap{
			"UpperCamelCase": UpperCamelCase,
		},
	})
}

var templateCode = `// Code generated by gen_embed_var.go; DO NOT EDIT.
package {{.Package}}

import _ "embed"

//go:embed default.yaml
var DefaultConfig string

// plugin default config
var (
	{{range .Plugins}}
	//go:embed {{$.Dir}}/{{.}}.yaml
	{{UpperCamelCase .}}DefaultConfig string
	{{end}}
)

var pluginDefaultConfigs = map[string]string{
	{{- range .Plugins}}
	"{{.}}":{{UpperCamelCase . -}}DefaultConfig,
	{{- end}}
}

//go:embed quickstart.yaml
var QuickStart string
`

// generate generates the code for Option `o` into a file named by `o.Path`.
func generate(o *Option) {
	tmpl, err := template.New("gen").Funcs(o.Funcs).Parse(templateCode)
	if err != nil {
		log.Fatal("template Parse:", err)
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, o)
	if err != nil {
		log.Fatal("template Execute:", err)
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Fatal("format:", err)
	}

	if err := ioutil.WriteFile(o.Path, formatted, 0644); err != nil {
		log.Fatal("writeFile:", err)
	}

}

// getYamlFiles returns a list of YAML files' names in the given directory.
func getYamlFiles(dir string) ([]string, error) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		fileNames = append(fileNames, strings.TrimSuffix(file.Name(), ".yaml"))
	}

	return fileNames, nil
}

// UpperCamelCase returns a string with the first letter in upper case.
func UpperCamelCase(s string) string {
	s = strings.Replace(s, "-", " ", -1)
	s = cases.Title(language.English).String(s)
	return strings.Replace(s, " ", "", -1)
}
