package auth

type AuthorizationError struct {
	originalError error
	message       map[string]string
}

func (err AuthorizationError) Error() string {
	return err.originalError.Error()
}

func (err AuthorizationError) Message() map[string]string {
	return err.message
}

func (err AuthorizationError) OriginalError() error {
	return err
}
