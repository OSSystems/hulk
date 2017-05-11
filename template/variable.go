package template

import (
	"fmt"
	"regexp"
)

type templateVariable struct {
	name           string
	isArray        bool
	arraySeparator string
	isOptional     bool
}

func (v *templateVariable) expanded(content string, value string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(v.string()))
	return re.ReplaceAllString(content, value)
}

func (v *templateVariable) string() string {
	array := map[bool]string{true: "[" + v.arraySeparator + "]", false: ""}
	optional := map[bool]string{true: "?", false: ""}

	return fmt.Sprintf("{%s%s}%s", v.name, array[v.isArray], optional[v.isOptional])
}

type arrayItem struct {
	variable *templateVariable
	value    string
	index    int
}
