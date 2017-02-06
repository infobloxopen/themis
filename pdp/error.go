package pdp

type MissingValueError struct {
	Err error
}

func (e MissingValueError) Error() string {
	return e.Err.Error()
}
