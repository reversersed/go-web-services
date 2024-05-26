package rest

import (
	"fmt"
	"strings"
)

type FilterOptions struct {
	Field    string
	Operator string
	Values   []string
}

func (f *FilterOptions) ToString() string {
	return fmt.Sprintf("%s%s", f.Operator, strings.Join(f.Values, ","))
}
