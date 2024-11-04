package pick

import (
	tp "github.com/kentio/norn/pkg/types"
	"testing"
)

func TestEqualSlice(t *testing.T) {
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

func TestNewResultComment(t *testing.T) {
	result := []*TaskResult{
		&TaskResult{
			Status: SucceedStatus,
			Branch: "branch1",
			Reason: "",
		},
		&TaskResult{
			Status: SucceedStatus,
			Branch: "branch2",
			Reason: "adsfasdf",
		},
	}
	comment, err := NewResultComment(tp.PickResultTemplate, result)
	if err != nil {
		t.Error("NewResultComment() = nil, want not nil")
	}
	t.Logf("comment: \n%s", comment)
}
