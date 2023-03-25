package github

import (
	"context"
	"github.com/kentio/norn/types"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestPullRequestService_FindComment(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	repo := "kentio/test_cherry_pick"
	mergeId := "53"
	client, _ := NewProvider(ctx, token)

	comments, err := client.Comment().Find(ctx, &types.FindCommentOption{Repo: repo, MergeRequestID: mergeId})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("comments: %+v", comments)
}

func TestCommentService_Create(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	client := NewGithubClient(ctx, token)

	commentService := NewCommentService(client)

	commentString := `
branches:
 - [ ] release/23.03
 - [ ] release/23.04
 - [ ] master
`
	err, _ := commentService.Create(ctx, &types.CreateCommentOption{
		Repo:           "kentio/test_cherry_pick",
		Body:           commentString,
		MergeRequestID: "53",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("add comment success")
}
