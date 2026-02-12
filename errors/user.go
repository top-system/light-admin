package errors

import "net/http"

var (
	UserRecordNotFound   = New("user record not found")
	UserInvalidPassword  = New("invalid user password")
	UserIsDisable        = New("user is disabled")
	UserPasswordRequired = New("user password is required")
	UserInvalidUsername   = New("invalid username")
	UserAlreadyExists    = New("user already exists")
	UserNoPermission     = New("user no permission")
	UserCannotUpdate     = New("super admin cannot update profile")
)

func init() {
	RegisterHTTPStatus(UserRecordNotFound, http.StatusNotFound)
	RegisterHTTPStatus(UserAlreadyExists, http.StatusConflict)
	RegisterHTTPStatus(UserInvalidPassword, http.StatusUnauthorized)
	RegisterHTTPStatus(UserNoPermission, http.StatusForbidden)
	RegisterHTTPStatus(UserIsDisable, http.StatusForbidden)
	RegisterHTTPStatus(UserCannotUpdate, http.StatusForbidden)
}
