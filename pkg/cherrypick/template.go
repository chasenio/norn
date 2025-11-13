package cherrypick

import tp "github.com/chasenio/norn/pkg/types"

// Templates holds the template configurations for cherry-pick comments.
type Templates struct {
	// SummaryTemplate is the template for generating cherry-pick summary comments.
	// It should contain a {{ .Message }} placeholder for the branch list.
	SummaryTemplate string
	
	// ResultTemplate is the template for generating cherry-pick result comments.
	// It should contain a {{ .Message }} placeholder for the result table.
	ResultTemplate string
}

// DefaultTemplates returns the default templates for cherry-pick operations.
func DefaultTemplates() *Templates {
	return &Templates{
		SummaryTemplate: tp.CherryPickTaskSummaryTemplate,
		ResultTemplate:  tp.PickResultTemplate,
	}
}

// NewTemplates creates a new Templates with the provided values.
// If any template is empty, it will use the default template.
func NewTemplates(summaryTemplate, resultTemplate string) *Templates {
	templates := &Templates{
		SummaryTemplate: summaryTemplate,
		ResultTemplate:  resultTemplate,
	}
	
	// Use defaults if not provided
	if templates.SummaryTemplate == "" {
		templates.SummaryTemplate = tp.CherryPickTaskSummaryTemplate
	}
	if templates.ResultTemplate == "" {
		templates.ResultTemplate = tp.PickResultTemplate
	}
	
	return templates
}
