package cmd

// ExitError is an error that carries a specific exit code.
// The main function should check for this error type and exit with the code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

func (e *ExitError) Unwrap() error {
	return e.Err
}
