package sqlu

// Enum needs to be implemented by Go enums (typed strings) in order to
// make them compatible with sqlu integration and fuzzing tests
type Enum interface {
	Enumerate() []string
}

// Enumerable is a type constraint for SQL enums
type Enumerable interface {
	~string
	Enum
}
