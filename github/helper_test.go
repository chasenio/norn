package github

import "testing"

func TestNewGithubClient(t *testing.T) {
	token := ""

	client := NewGithubClient(nil, token)

	t.Logf("client: %v", client)
}
