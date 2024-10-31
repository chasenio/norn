package github

import (
	"context"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestPullRequestService_FindComment(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	repo := "kentio/test_cherry_pick"
	mergeId := "53"
	client := NewProvider(ctx, &tp.CreateProviderOption{Token: token})

	comments, err := client.Comment().Find(ctx, &tp.FindCommentOption{Repo: repo, MergeRequestID: mergeId})
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
	comment, err := commentService.Create(ctx, &tp.CreateCommentOption{
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

func TestCommentService_Delete(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	token := ""
	client := NewGithubClient(ctx, token)

	commentService := NewCommentService(client)

	err := commentService.Delete(ctx, &tp.DeleteCommentOption{
		CommentID: "2449993187",
		Repo:      "",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("delete comment success")
}
