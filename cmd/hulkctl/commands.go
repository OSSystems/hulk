package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/zyedidia/highlight"
	yaml "gopkg.in/yaml.v2"

	"github.com/OSSystems/hulk/client"
)

// Text formatting
var (
	Green = color.New(color.FgGreen).SprintFunc()
	Bold  = color.New(color.Bold).SprintFunc()
)

func ListServices(cli *client.Client) error {
	services, err := cli.ServiceList()
	if err != nil {
		return err
	}

	fmt.Printf("%s  %s\n", Bold("ENABLED"), Bold("SERVICE"))

	for _, service := range services {
		status := Green("x")

		if !service.Enabled {
			status = " "
		}

		fmt.Printf("[ %s ]    %s\n", status, service.Name)
	}

	return nil
}

func InspectService(cli *client.Client, name string) error {
	service, err := cli.GetService(name)
	if err != nil {
		return err
	}

	output, err := yaml.Marshal(service)
	if err != nil {
		return err
	}

	syntaxDef, _ := highlight.ParseDef([]byte(yamlSyntax))

	h := highlight.NewHighlighter(syntaxDef)
	matches := h.HighlightString(string(output))
	lines := strings.Split(string(output), "\n")

	for lineN, l := range lines {
		for colN, c := range l {
			if group, ok := matches[lineN][colN]; ok {
				if group == highlight.Groups["statement"] {
					color.Set(color.FgGreen)
				} else if group == highlight.Groups["preproc"] {
					color.Set(color.FgHiRed)
				} else if group == highlight.Groups["special"] {
					color.Set(color.FgBlue)
				} else if group == highlight.Groups["constant.string"] {
					color.Set(color.FgCyan)
				} else if group == highlight.Groups["constant.specialChar"] {
					color.Set(color.FgHiMagenta)
				} else if group == highlight.Groups["type"] {
					color.Set(color.FgYellow)
				} else if group == highlight.Groups["constant.number"] {
					color.Set(color.FgCyan)
				} else if group == highlight.Groups["comment"] {
					color.Set(color.FgHiGreen)
				} else {
					color.Unset()
				}
			}

			fmt.Print(string(c))
		}

		if group, ok := matches[lineN][len(l)]; ok {
			if group == highlight.Groups["default"] || group == highlight.Groups[""] {
				color.Unset()
			}
		}

		fmt.Print("\n")
	}

	return nil
}
