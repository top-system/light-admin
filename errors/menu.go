package errors

import "net/http"

var (
	MenuRecordNotFound          = New("menu record not found")
	MenuAlreadyExists           = New("menu already exists")
	MenuInvalidParent           = New("menu invalid parent")
	MenuNotAllowDeleteWithChild = New("contains children, cannot be deleted")
)

func init() {
	RegisterHTTPStatus(MenuRecordNotFound, http.StatusNotFound)
	RegisterHTTPStatus(MenuAlreadyExists, http.StatusConflict)
}
