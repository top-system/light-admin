package errors

import "net/http"

var (
	RoleRecordNotFound         = New("role record not found")
	RoleIsDisable              = New("role is disabled")
	RoleAlreadyExists          = New("role already exists")
	RoleCodeAlreadyExists      = New("role code already exists")
	RoleNotAllowDeleteWithUser = New("used by users, cannot be deleted")
)

func init() {
	RegisterHTTPStatus(RoleRecordNotFound, http.StatusNotFound)
	RegisterHTTPStatus(RoleAlreadyExists, http.StatusConflict)
	RegisterHTTPStatus(RoleCodeAlreadyExists, http.StatusConflict)
	RegisterHTTPStatus(RoleIsDisable, http.StatusForbidden)
}
