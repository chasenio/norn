package types

import "context"

type PickOption struct {
	SHA    string
	Branch string
}

type PickService interface {
	Pick(ctx context.Context, repo string, opt *PickOption) error
}
