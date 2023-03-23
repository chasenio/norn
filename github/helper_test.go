package github

import "testing"

func TestNewGithubClient(t *testing.T) {
	token := "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"

	client := NewGithubClient(nil, token)

	t.Logf("client: %v", client)
}
