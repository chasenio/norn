package types

import "context"

type Option struct {
	SHA    string
	Branch string
	Prefix string
}

type CherryService interface {
	CherryPick(ctx context.Context, repo string, opt *Option) error
}
