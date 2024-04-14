package pick

import (
	"context"
	"github.com/kentio/norn/internal"
	"github.com/kentio/norn/pkg/common"
	"github.com/kentio/norn/pkg/github"
	"github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestPickFeature_DoPickSummaryComment(t *testing.T) {
	ctx := context.Background()
	provider := github.NewProvider(ctx, "")
	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		Form:           "release/23.03",
		IsSummary:      true,
		SHA:            common.String("cc382f5c74a879bda50cc5a8a73090ba83068733"),
		MergeRequestID: "60",
	}
	pick := NewPickService(provider, pickOpt.Branches)
	err := pick.CreateSummaryComment(ctx, pickOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}

func TestPick(t *testing.T) {
	ctx := context.Background()
	provider := github.NewProvider(ctx, "")
	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		Form:           "release/23.03",
		IsSummary:      false,
		SHA:            common.String("9fe34a912edd44ef07052aa1305aea72adee3638"),
		MergeRequestID: "59",
	}
	pick := NewPickService(provider, pickOpt.Branches)

	err := pick.PerformPick(ctx, &CherryPick{
		SHA:    *pickOpt.SHA,
		Repo:   pickOpt.Repo,
		Target: "master"})

	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}

func TestPickFeature_IsInMergeRequestComments(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	provider := github.NewProvider(ctx, "")
	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		Form:           "release/23.03",
		IsSummary:      false,
		SHA:            common.String(""),
		MergeRequestID: "54",
	}
	pick := NewPickService(provider, pickOpt.Branches)
	// Is Exist
	comment, err := pick.CheckSummaryExist(ctx, pickOpt.Repo, pickOpt.MergeRequestID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if comment == nil {
		t.Fatalf("err: %v", err)
	}

	pickOpt.MergeRequestID = "45"
	// Is Not Exist
	comment, err = pick.CheckSummaryExist(ctx, pickOpt.Repo, pickOpt.MergeRequestID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if comment != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestParseSelectedBranches(t *testing.T) {
	text := `Will be cherry-picked to the following branches:

---
- [x] master
- [x] dev
- [x] release/23.03


<!-- Do not edit or delete , This is a cherry-pick summary flag. | o((>ω< ))o -->
---`

	results := parseSelectedBranches(text)
	t.Logf("results: %+v", results)

	case1 := []string{"master", "dev", "release/23.03"}
	if len(results) != len(case1) {
		t.Fatalf("parse selected branches failed")
	}

	for _, v := range results {
		if !internal.StringInSlice(v, case1) {
			t.Fatalf("parse selected branches failed")
		}
	}
}

func TestDoPickToBranchesFromMergeRequest(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	provider, _ := common.NewProvider(ctx, "github", token)

	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		Form:           "release/23.03",
		IsSummary:      false,
		SHA:            common.String("a569472376cd1f5ff8403811ceb67b9f809f961f"),
		MergeRequestID: "60",
	}
	pick := NewPickService(provider, pickOpt.Branches)

	// test is summary task
	done, faild, err := pick.PerformPickToBranches(ctx, pickOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("done: %v, faild: %v", done, faild)

	// test done comment

}

func TestNewMergeReqeustComment(t *testing.T) {
	// test is summary task comment
	isSummaryOpt := Result{
		branches: []string{"master", "dev"},
	}
	isSummaryResult, err := NewSummaryComment(true, &isSummaryOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("isSummaryResult: %s", isSummaryResult)

	// test done comment
	doneOpt := Result{
		done: []string{"master", "dev"},
	}
	doneResult, err := NewSummaryComment(false, &doneOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("doneResult: %s", doneResult)
	//if !strings.Contains(doneResult, global.CherryPickTaskDoneTemplate) {
	//	t.Fatalf("err: %v", err)
	//}

	// test failed comment
	failedOpt := Result{
		failed: []string{"master", "dev"},
	}
	failedResult, err := NewSummaryComment(false, &failedOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("failedResult: %s", failedResult)

	// test done and failed comment
	doneAndFailedOpt := Result{
		done:   []string{"master", "dev"},
		failed: []string{"aa", "bb"},
	}
	doneAndFailedResult, err := NewSummaryComment(false, &doneAndFailedOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("doneAndFailedResult: %s", doneAndFailedResult)
}

func TestNewCommentContent(t *testing.T) {
	branches := []string{"master", "dev"}

	taskSummaryResult, err := NewSelectComment(types.CherryPickTaskSummaryTemplate, branches)
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
	doneResult, err := NewSelectComment(types.CherryPickTaskDoneTemplate, branches)
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
	failedResult, err := NewSelectComment(types.CherryPickTaskFailedTemplate, branches)
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
