package sqlquery

import "errors"

var (
	AuthFailure       = errors.New("Auth failure")
	DatabaseFailure   = errors.New("Database doesn't exist")
	ConnectionFailure = errors.New("Connection failure")
)
