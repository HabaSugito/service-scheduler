package repository

import "errors"

var (
	ErrConflict             = errors.New("technician has a conflicting job in this time window")
	ErrQuoteAlreadyAssigned = errors.New("quote is already assigned to a job")
)
