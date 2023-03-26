package global

const (
	CherryPickSummaryFlag = "<!-- Do not edit or delete , This is a cherry-pick summary flag. | o((>ω< ))o -->"
)

const (
	CherryPickTaskSummaryTemplate = "" +
		"Will be cherry-picked to the following branches:\n\n" +
		"{{ .Message }}\n\n" +
		CherryPickSummaryFlag
	CherryPickTaskDoneTemplate = "" +
		"Cherry-picked to the following branches:\n\n" +
		"{{ .Message }}\n\n"
	CherryPickTaskFailedTemplate = "" +
		"Cherry-pick failed to the following branches:\n\n" +
		"{{ .Message }}\n\n"
)
