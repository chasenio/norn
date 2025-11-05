package github

import (
	"github.com/chasenio/norn/pkg/types"
	"testing"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider(nil, &types.CreateProviderOption{Token: ""})

	t.Logf("provider: %v", provider)
}
