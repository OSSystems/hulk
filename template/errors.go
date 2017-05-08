package template

// Errors returned when expanding a variable
var (
	ErrRequiredValueNotFound = "No value for required variable"
	ErrOptionalValueNotFound = "No value for optional variable"
)

// VariableExpandError implements an error returned when expanding a variable
type VariableExpandError struct {
	Name       string
	IsOptional bool
}

// Error returns a string representation of an VariableExpandError
func (e *VariableExpandError) Error() string {
	if e.IsOptional {
		return ErrOptionalValueNotFound
	}

	return ErrRequiredValueNotFound
}
