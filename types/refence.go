package types

import "context"

type Reference struct {
	// Name of the branch.
	Ref string
	SHA string
}

type GetRefOption struct {
	Repo string
	Ref  string
}

type UpdateOption struct {
	Repo string
	Ref  string
	SHA  string
}

type ReferenceService interface {
	Get(ctx context.Context, opt *GetRefOption) (*Reference, error)
	Update(ctx context.Context, opt *UpdateOption) (*Reference, error)
}
