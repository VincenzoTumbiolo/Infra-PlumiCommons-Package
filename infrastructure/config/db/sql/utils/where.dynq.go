package sqlu

import (
	"github.com/amplifon-x/ax-go-application-layer/v5/slicex"
	"github.com/taleeus/sqld"
)

// Filter is a collection of Conditions
type Filter interface {
	// Parse transpiles the filter to an SQL string
	Parse(params *sqld.Params) string
}

// Where is a collection of Filters
type Where[F Filter] struct {
	Condition sqld.Op `json:"condition,omitempty" query:"condition"`
	Filters   []F     `json:"filters,omitempty" query:"filters"`
}

func (where Where[F]) Parse(params *sqld.Params) string {
	if where.Condition == "" {
		where.Condition = sqld.AND
	}

	return sqld.Cond(where.Condition, slicex.Map(where.Filters, func(f F) string {
		return f.Parse(params)
	})...)
}
