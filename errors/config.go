package errors

import "net/http"

var (
	ConfigKeyAlreadyExists = New("config key already exists")
	ConfigNotFound         = New("config not found")
)

func init() {
	RegisterHTTPStatus(ConfigNotFound, http.StatusNotFound)
	RegisterHTTPStatus(ConfigKeyAlreadyExists, http.StatusConflict)
}
