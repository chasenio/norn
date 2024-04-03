package github

import "testing"

func TestNewProvider(t *testing.T) {
	token := ""
	provider := NewProvider(nil, token)

	t.Logf("provider: %v", provider)
}
