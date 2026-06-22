package service

import (
	"errors"

	"connectrpc.com/connect"
	"gorm.io/gorm"
)

// notFoundOrInternal maps a GORM not-found error to CodeNotFound,
// otherwise wraps as CodeInternal.
func notFoundOrInternal(err error) *connect.Error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

// isNotFound reports whether err represents a GORM record-not-found condition.
func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
