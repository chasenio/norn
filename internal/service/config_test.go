package service

import "testing"

func TestConfig_PrivateKey(t *testing.T) {
	cfg := &Config{
		Github: &GithubConfig{
			PrivateKey: ``,
		},
	}

	if cfg.Github.PrivateKey == "" {
		t.Errorf("PrivateKey is empty")
	}
	_, err := cfg.PrivateKey()
	if err != nil {
		t.Errorf("PrivateKey error: %v", err)
	}
	t.Logf("success.")
}
