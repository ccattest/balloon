package store

import (
	"fmt"
)

var (
	// ErrInvalidParam is returned when an invalid parameter has been passsed
	ErrInvalidParam = fmt.Errorf("error: invalid parameter")
	// ErrMissingParam is an error returned when the caller is missing a necessary parameter
	ErrMissingParam = fmt.Errorf("error: missing param %w", ErrInvalidParam)
	// ErrFailedScan is returned when a row fails to scan into a typpe
	ErrFailedScan = fmt.Errorf("error: failed to scan row")
)
