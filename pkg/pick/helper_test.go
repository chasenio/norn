package pick

import "testing"

func TestEqualClice(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a", "b", "c"}
	c := []string{"a", "b", "d"}
	if !EqualSlice(a, b) {
		t.Errorf("EqualSlice(%v, %v) = false, want true", a, b)
	}

	if EqualSlice(a, c) {
		t.Errorf("EqualSlice(%v, %v) = true, want false", a, c)
	}

}
