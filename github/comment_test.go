package github

import (
	"context"
	types2 "github.com/kentio/norn/types"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestPullRequestService_FindComment(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	repo := "kentio/test_cherry_pick"
	mergeId := "53"
	client := NewProvider(ctx, token)

	comments, err := client.Comment().Find(ctx, &types2.FindCommentOption{Repo: repo, MergeRequestID: mergeId})
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
	comment, err := commentService.Create(ctx, &types2.CreateCommentOption{
		Repo:           "kentio/test_cherry_pick",
		Body:           commentString,
		MergeRequestID: "58",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if comment.Body() != commentString {
		t.Fatalf("comment body not equal")
	}
	t.Logf("add comment success")
}
