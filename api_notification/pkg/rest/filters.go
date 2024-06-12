package rest

import (
	"strings"
)

type FilterOptions struct {
	Field  string
	Values []string
}

func (f *FilterOptions) ToString() string {
	return strings.Join(f.Values, ",")
}
