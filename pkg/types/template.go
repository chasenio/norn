package types

const (
	CherryPickSummaryFlag = "<!-- Do not edit or delete , This is a cherry-pick summary flag. | o((>ω< ))o -->"
)

const (
	CherryPickTaskSummaryTemplate = "" +
		"Will be cherry-picked to the following branches:\n\n" +
		"{{ .Message }}\n\n" +
		CherryPickSummaryFlag
	PickResultTemplate = "" +
		"Pick Result: \n" +
		"{{ .Message }}\n\n"
)
