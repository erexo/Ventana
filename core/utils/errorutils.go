package utils

import "github.com/pkg/errors"

func ConcatErrors(e1, e2 error) error {
	if e1 == nil {
		return e2
	} else if e2 == nil {
		return e1
	}
	return errors.Wrap(e1, e2.Error())
}
