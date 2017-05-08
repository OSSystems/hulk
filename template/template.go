package template

import (
	"fmt"
	"regexp"
	"strings"
)

const templateRegexp = `{(?P<name>[a-zA-Z_][a-zA-Z0-9_]+)(?P<array>\[.?\])?(?P<optional>\?)?}`

type template struct {
	content   string
	variables []*templateVariable
}

func (t *template) parse() {
	for _, v := range t.match() {
		variable := &templateVariable{
			name:       v["name"],
			isArray:    v["array"] != "",
			isOptional: v["optional"] != "",
		}

		if variable.isArray {
			if v["array"] != "[]" {
				variable.arraySeparator = string(v["array"][1])
			}
		}

		t.variables = append(t.variables, variable)
	}
}

func (t *template) expand(values map[string]string) ([]string, error) {
	content := t.content
	arrayValues := map[int]map[*templateVariable][]string{}

	t.parse()

	for i, variable := range t.variables {
		// Check if contains the required variable value
		if _, ok := values[variable.name]; !ok {
			return nil, &VariableExpandError{Name: variable.name, IsOptional: variable.isOptional}
		}

		// Expand content
		if !variable.isArray {
			content = variable.expanded(content, values[variable.name])
		} else {
			if arrayValues[i] == nil {
				arrayValues[i] = make(map[*templateVariable][]string)
			}

			// Set default separator
			separator := variable.arraySeparator
			if separator == "" {
				separator = " "
			}

			arrayValues[i][variable] = strings.Split(values[variable.name], separator)
		}
	}

	// Returns the expanded template content if there is no array inside it
	if len(arrayValues) == 0 {
		return []string{content}, nil
	}

	return t.expandArrays(content, arrayValues)
}

func (t *template) expandArrays(content string, arrayValues map[int]map[*templateVariable][]string) ([]string, error) {
	arrayItems := map[int][]*arrayItem{}

	length := -1

	for _, variables := range arrayValues {
		for variable, values := range variables {
			if length > -1 && length != len(values) {
				return nil, fmt.Errorf("Array size differs: %s (%d should be %d)", variable.name, len(values), length)
			}

			length = len(values)

			for index, value := range values {
				item := &arrayItem{
					variable: variable,
					value:    value,
					index:    index,
				}

				arrayItems[index] = append(arrayItems[index], item)
			}
		}
	}

	expandedArrays := []string{}

	for _, items := range arrayItems {
		expandedItem := content

		for _, item := range items {
			expandedItem = item.variable.expanded(expandedItem, item.value)
		}

		expandedArrays = append(expandedArrays, expandedItem)
	}

	return expandedArrays, nil
}

func (t *template) match() []map[string]string {
	result := []map[string]string{}
	re := regexp.MustCompile(templateRegexp)
	groups := re.SubexpNames()[1:]
	matches := re.FindAllStringSubmatch(t.content, -1)

	if len(matches) == 0 {
		return result
	}

	for _, v := range matches {
		m := map[string]string{}

		for i, group := range groups {
			m[group] = v[i+1]
		}

		result = append(result, m)
	}

	return result
}

// Expand expands values into template content
func Expand(content string, values map[string]string) ([]string, error) {
	tpl := &template{
		content: content,
	}

	return tpl.expand(values)
}
