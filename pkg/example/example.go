package example

import (
	"errors"
)

type Example struct {
	UserId  string
	Id      string
	Message string
}

var EmptyMessageError = errors.New("message length must be greater than 0")

func New(userId, msg string) (Example, error) {
	ok := CannotBeEmpty(msg)

	if !ok {
		return Example{}, EmptyMessageError
	}

	return Example{
		UserId:  userId,
		Message: msg,
	}, nil
}

func Nil() Example {
	return Example{}
}

func (e *Example) SetMessage(msg string) error {
	ok := CannotBeEmpty(msg)

	if !ok {
		return EmptyMessageError
	}

	e.Message = msg

	return nil
}
