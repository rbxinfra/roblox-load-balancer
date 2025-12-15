package haproxy

import (
	"bufio"
	"io"
	"os"
	"strings"
	"text/template"

	"github.rbx.com/roblox/roblox-load-balancer/configuration"
)

// BuildTemplateFile constructs the output template file.
func BuildTemplateFile(backendsMap, rulesMap map[string]string, config *configuration.Config) (string, error) {
	tpl := template.New("haproxy_config")

	funcMap := template.FuncMap{
		"backends": func(entryPoint string) string {
			return backendsMap[entryPoint]
		},
		"rules": func(entryPoint string) string {
			return rulesMap[entryPoint]
		},
	}

	tpl.Funcs(funcMap)

	tplFile, err := os.Open(config.TemplateFilePath)
	if err != nil {
		return "", err
	}
	defer tplFile.Close()

	stat, err := tplFile.Stat()
	if err != nil {
		return "", err
	}

	bytes := make([]byte, stat.Size())
	_, err = bufio.NewReader(tplFile).Read(bytes)
	if err != nil && err != io.EOF {
		return "", err
	}

	if tpl, err = tpl.Parse(string(bytes)); err != nil {
		return "", err
	}

	var textWriter strings.Builder
	if err = tpl.Execute(&textWriter, nil); err != nil {
		return "", err
	}

	return textWriter.String(), nil
}
