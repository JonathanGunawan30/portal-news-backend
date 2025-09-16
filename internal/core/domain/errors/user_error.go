package errors

import "errors"

var (
	ErrorUserNotFound    = errors.New("user not found")
	ErrorPasswordInvalid = errors.New("current password is wrong")
)
