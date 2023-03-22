package types

import "time"

type GetCommitOption struct {
	Owner string
	Repo  string
	Sha   string
}

type CommitInfo struct {
	// Name of the commit.
	NodeId    *string
	SHA       *string
	Commit    *CommitInfo `json:"commit,omitempty"`
	Author    *UserSpec   `json:"author,omitempty"`
	Committer *UserSpec   `json:"committer,omitempty"`
	HTMLURL   *string     `json:"html_url,omitempty"`
}

type UserSpec struct {
	Name *string
}

type Commit struct {
	SHA       *string       `json:"sha,omitempty"`
	Author    *CommitAuthor `json:"author,omitempty"`
	Committer *CommitAuthor `json:"committer,omitempty"`
	Message   *string       `json:"message,omitempty"`
	HTMLURL   *string       `json:"html_url,omitempty"`
	URL       *string       `json:"url,omitempty"`
	NodeID    *string       `json:"node_id,omitempty"`

	CommentCount *int `json:"comment_count,omitempty"`
}

type CommitAuthor struct {
	Date  *time.Time `json:"date,omitempty"`
	Name  *string    `json:"name,omitempty"`
	Email *string    `json:"email,omitempty"`

	Login *string `json:"username,omitempty"` // Renamed for go-github consistency.
}

type CommitStats struct {
	Additions *int `json:"additions,omitempty"`
	Deletions *int `json:"deletions,omitempty"`
	Total     *int `json:"total,omitempty"`
}
