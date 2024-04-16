package github

import (
	"testing"
)

func Test_Checkout(t *testing.T) {
	opt := CheckoutOption{
		RepoPath: "",
		Branch:   "master",
	}
	err := Checkout(&opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Checkout success")
}

func Test_ApplyPatch(t *testing.T) {
	opt := ApplyPatchOption{
		Patch:    "",
		RepoPath: "",
	}
	err := ApplyPatch(&opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ApplyPatch success")
}

func Test_CherryPick(t *testing.T) {
	opt := &CherryPickOption{
		RepoPath: "",
		Commit:   "",
	}
	err := CherryPick(opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("CherryPick success")
}
