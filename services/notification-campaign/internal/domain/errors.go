package domain

import "errors"

func IsDomainError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}
