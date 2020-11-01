package entity

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	minNameLength = 4
	maxNameLength = 255
)

var (
	EmptyErr = errors.New("Empty")
	ShortErr = fmt.Errorf("Must be equal to or longer than %d", minNameLength)
	LongErr  = fmt.Errorf("Must be equal to or shorter than %d", maxNameLength)
)

func ValidateName(name *string) error {
	u := strings.TrimSpace(strings.ToLower(*name))
	if u == "" {
		return EmptyErr
	}
	if len(u) < minNameLength {
		return ShortErr
	}
	if len(u) > maxNameLength {
		return LongErr
	}
	*name = u
	return nil
}

func ValidateEmpty(str *string) error {
	u := strings.TrimSpace(strings.ToLower(*str))
	if u == "" {
		return EmptyErr
	}
	*str = u
	return nil
}
