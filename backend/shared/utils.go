package shared

import "regexp"

var QueryParameterPlaceholder = regexp.MustCompile(`\$\d+`)
