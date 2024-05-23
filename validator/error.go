package validator

type ValidateError struct {
	err error
}

func (e *ValidateError) Error() string {
	return e.err.Error()
}

func NewValidateError(err error) *ValidateError {
	if err == nil {
		return nil
	}
	return &ValidateError{
		err: err,
	}
}
