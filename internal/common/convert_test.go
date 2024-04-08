package common

import "testing"

func TestToPrivateKeys(t *testing.T) {
	key := ``
	_, err := ToPrivateKeys(key)
	if err != nil {
		t.Errorf("ToPrivateKeys error: %v", err)
	}
	t.Logf("ToPrivateKeys success, key: %v", key)
}
