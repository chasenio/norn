package cherrypick

import (
	"context"
	"errors"
	"github.com/chasenio/norn/internal"
	"github.com/chasenio/norn/pkg/common"
	"github.com/chasenio/norn/pkg/github"
	tp "github.com/chasenio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestPick_CreateSummaryWithTask(t *testing.T) {
	ctx := context.Background()
	provider := github.NewProvider(ctx, &tp.CreateProviderOption{Token: ""})
	pickOpt := &Task{
		Repo: "kentio/pick",
		Branches: []string{
			"r1",
			"r2",
			"master",
		},
		From:           "r1",
		IsSummary:      false,
		SHA:            common.String(""),
		MergeRequestID: "64",
	}
	pick := NewPickService(provider, "", "")
	err := pick.CreateSummaryWithTask(ctx, pickOpt)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}

func TestPick(t *testing.T) {
	ctx := context.Background()
	provider := github.NewProvider(ctx, &tp.CreateProviderOption{Token: ""})
	task := &Task{
		Repo: "kentio/pick",
		Branches: []string{
			"r1",
			"r2",
			"master",
		},
		From:           "r1",
		IsSummary:      false,
		SHA:            common.String(""),
		MergeRequestID: "2",
	}
	pick := NewPickService(provider, "", "")

	//err := pick.ProcessPick(ctx, task)
	err := pick.PerformPick(ctx, &CherryPickOptions{
		SHA:    *task.SHA,
		Repo:   task.Repo,
		Target: "master",
		Pr:     64,
	})

	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("err: %v", err)
}

func TestPick_CheckSummaryExist(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	provider := github.NewProvider(ctx, &tp.CreateProviderOption{Token: ""})
	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		From:           "release/23.03",
		IsSummary:      false,
		SHA:            common.String(""),
		MergeRequestID: "54",
	}
	pick := NewPickService(provider, "", "")
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

func TestPerformPickToBranches(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	provider, _ := common.NewProvider(ctx, "github", &tp.CreateProviderOption{Token: token})

	pickOpt := &Task{
		Repo: "kentio/test_cherry_pick",
		Branches: []string{
			"release/23.03",
			"release/23.04",
			"master",
		},
		From:           "release/23.03",
		IsSummary:      true,
		SHA:            common.String(""),
		MergeRequestID: "66",
	}
	pick := NewPickService(provider, "", "")

	_, comment, err := pick.FindCommentWithTask(ctx, pickOpt, tp.CherryPickSummaryFlag)

	// test is summary task
	result, err := pick.PerformPickToBranches(ctx, pickOpt, comment)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("result %v", &result)

	// test done comment

}

func Test_GetTree(t *testing.T) {
	ctx := context.Background()
	client := github.NewGithubClient(ctx, "")
	tree, _, err := client.Git.GetTree(ctx, "", "", "", true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, v := range tree.Entries {
		t.Logf("tree: %s", *v.Path)
	}

}

func Test_CreateTree(t *testing.T) {
	ctx := context.Background()
	client := github.NewGithubClient(ctx, "")
	tree, _, err := client.Git.CreateTree(ctx, "", "", "", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("tree: %v", tree)
}

func Test_PickErrMessage(t *testing.T) {
	err := errors.New("https://api.github.com/repos/xxx/xxx/merges: 404 Base does not exist []")
	message := strings.Split(err.Error(), " ")
	if len(message) > 1 {
		err = errors.New(strings.Join(message[1:], " "))
	}
	logrus.Infof("err: %v", err)
}
