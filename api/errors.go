package api

var (
	invalidAuthHeader = ErrorResponse{Error: "invalid auth header"}
	invalidToken = ErrorResponse{Error: "invalid token"}
)
