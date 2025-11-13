package cherrypick

import tp "github.com/chasenio/norn/pkg/types"

type MessageTemplate struct {
	UniqueID string
	Text     string
}

// Templates holds the template configurations for cherry-pick comments.
type Templates struct {
	Summary          *MessageTemplate
	CherryPickResult *MessageTemplate
}

// DefaultTemplates returns the default templates for cherry-pick operations.
func DefaultTemplates() *Templates {
	return &Templates{
		Summary: &MessageTemplate{
			UniqueID: tp.CherryPickSummaryFlag,
			Text:     tp.CherryPickTaskSummaryTemplate},
		CherryPickResult: &MessageTemplate{
			UniqueID: tp.CherryPickResultFlag,
			Text:     tp.CherryPickResultTemplate,
		},
	}
}
