package feature

import (
	"context"
	"github.com/kentio/norn/github"
	"github.com/kentio/norn/global"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestPick(t *testing.T) {
	ctx := context.Background()
	provider, _ := github.NewProvider(ctx, "")

	err := Pick(ctx, provider, &PickOption{
		SHA:    "696b3168704d0d5b811d80615b3e1a6a31b2d2a5",
		Repo:   "",
		Target: "master"})

	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}

func TestParseSelectedBranches(t *testing.T) {
	text := `Will be cherry-picked to the following branches:

---
- [x] master
- [x] dev
- [x] release/23.03


<!-- Do not edit or delete , This is a cherry-pick summary flag. | o((>ω< ))o -->
---`

	results := ParseSelectedBranches(text)
	t.Logf("results: %+v", results)

	case1 := []string{"master", "dev", "release/23.03"}
	if len(results) != len(case1) {
		t.Fatalf("parse selected branches failed")
	}

	for _, v := range results {
		if !global.StringInSlice(v, case1) {
			t.Fatalf("parse selected branches failed")
		}
	}
}

func TestDoPickToBranchesFromMergeRequest(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	provider, _ := global.NewProvider(ctx, "github", token)

	// test is summary task
	done, faild, err := DoPickToBranchesFromMergeRequest(ctx, provider, &PickToRefMROpt{
		Repo:           "kentio/test_cherry_pick",
		Branches:       []string{"master", "dev"},
		Form:           "release/23.03",
		SHA:            "xxx",
		MergeRequestID: "53",
		IsSummaryTask:  true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("done: %v, faild: %v", done, faild)

	// test done comment

}

func TestNewMergeReqeustComment(t *testing.T) {
	// test is summary task comment
	isSummaryOpt := MergeCommentOpt{
		branches: []string{"master", "dev"},
	}
	isSummaryResult, err := NewMergeReqeustComment(true, &isSummaryOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("isSummaryResult: %s", isSummaryResult)

	// test done comment
	doneOpt := MergeCommentOpt{
		done: []string{"master", "dev"},
	}
	doneResult, err := NewMergeReqeustComment(false, &doneOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("doneResult: %s", doneResult)
	//if !strings.Contains(doneResult, global.CherryPickTaskDoneTemplate) {
	//	t.Fatalf("err: %v", err)
	//}

	// test failed comment
	failedOpt := MergeCommentOpt{
		failed: []string{"master", "dev"},
	}
	failedResult, err := NewMergeReqeustComment(false, &failedOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("failedResult: %s", failedResult)

	// test done and failed comment
	doneAndFailedOpt := MergeCommentOpt{
		done:   []string{"master", "dev"},
		failed: []string{"aa", "bb"},
	}
	doneAndFailedResult, err := NewMergeReqeustComment(false, &doneAndFailedOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("doneAndFailedResult: %s", doneAndFailedResult)
}

func TestNewCommentContent(t *testing.T) {
	branches := []string{"master", "dev"}

	taskSummaryResult, err := NewCommentContent(global.CherryPickTaskSummaryTemplate, branches)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	taskSummaryContent := taskSummaryResult.String()
	for _, v := range branches {
		if !strings.Contains(taskSummaryContent, v) {
			t.Fatalf("err: %v", err)
		}
	}

	// test Done template
	doneResult, err := NewCommentContent(global.CherryPickTaskDoneTemplate, branches)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	doneContent := doneResult.String()
	for _, v := range branches {
		if !strings.Contains(doneContent, v) {
			t.Fatalf("err: %v", err)
		}
	}

	// test Failed template
	failedResult, err := NewCommentContent(global.CherryPickTaskFailedTemplate, branches)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	failedContent := failedResult.String()
	for _, v := range branches {
		if !strings.Contains(failedContent, v) {
			t.Fatalf("err: %v", err)
		}
	}

}
