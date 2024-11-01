package types

type ProviderError struct {
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}

func NewProviderError(message string) *ProviderError {
	return &ProviderError{
		Message: message,
	}
}

var (
	ErrInvalidOptions = NewProviderError("invalid parameter")
	ErrConflict       = NewProviderError("conflict")

	NotFound = NewProviderError("not found")

	ErrUnknownProvider = NewProviderError("unknown provider")
)
