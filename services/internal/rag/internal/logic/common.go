package logic

import "errors"

var (
	errPermissionDenied   = errors.New("permission denied")
	errKBNotFound         = errors.New("knowledge base not found")
	errDocNotFound        = errors.New("document not found")
	errInvalidParam       = errors.New("invalid parameter")
)