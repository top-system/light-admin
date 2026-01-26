package errors

var (
	ConfigKeyAlreadyExists = New("config key already exists")
	ConfigNotFound         = New("config not found")
)
