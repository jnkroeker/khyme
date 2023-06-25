// Package order provides support for describing the ordering of data.
package order

// Set of directions for ordering data.
const (
	ASC  = "ASC"
	DESC = "DESC"
)

var directions = map[string]string{
	ASC:  "ASC",
	DESC: "DESC",
}

// ======================================================

// By represents a field used to order by and direction.
type By struct {
	Field     string
	Direction string
}

// NewBy constructs a new By value with not checks.
func NewBy(field string, direction string) By {
	return By{
		Field:     field,
		Direction: direction,
	}
}

// ======================================================
