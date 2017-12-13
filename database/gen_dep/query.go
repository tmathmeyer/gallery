package gen_dep

import (
	"fmt"
	"strings"
)

func CreateQuery(query map[string]interface{}) string {
	if len(query) == 0 {
		return ""
	}
	var conditionals []string
	for k, v := range query {
		add := ""
		switch v.(type) {
		case string:
			add = fmt.Sprintf("%s is '%s'", k, v)
		default:
			add = fmt.Sprintf("%s is %#v", k, v)
		}
		conditionals = append(conditionals, add)
	}
	return " where " + strings.Join(conditionals, " and ")
}