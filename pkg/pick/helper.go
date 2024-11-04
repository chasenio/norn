package pick

import (
	"crypto/md5"
	"fmt"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

func sumMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// parseSelectedBranches parse selected branches from comment
func parseSelectedBranches(comment string) (selected []string) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		if strings.Contains(line, "- [x]") {
			line = strings.ReplaceAll(line, "- [x] ", "") // remove "- [x] "
			line = strings.ReplaceAll(line, " ", "")      // remove " "
			// remove enter
			line = strings.ReplaceAll(line, "\r", "")
			line = strings.ReplaceAll(line, "\n", "")

			if line == "" {
				continue
			}
			selected = append(selected, line)
		}
	}
	return selected
}

func generateTargetBranches(task *Task) []string {
	var targets []string
	var startFlag bool
	for _, branch := range task.Branches {
		// Skip branches before the 'From' branch in the list
		if branch == task.From {
			startFlag = true
			continue
		}
		if startFlag {
			targets = append(targets, branch)
		} else {
			logrus.Debugf("Skipping branch: %s, from: %s", branch, task.From)
		}
	}
	return targets
}

// EqualSlice compares two slices of strings
func EqualSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GetCheckConflictMode returns the check conflict mode for the provider
func GetCheckConflictMode(provider tp.ProviderType) tp.CheckConflictMode {
	switch provider {
	case tp.GitHubProvider:
		return tp.WithCommand
	default:
		return tp.WithAPI
	}
}

// NewResultComment generate comment content
func NewResultComment(layout string, result []*TaskResult) (string, error) {
	var resultContent strings.Builder
	var content strings.Builder
	type Msg struct {
		Message string `json:"message"`
	}
	table := tablewriter.NewWriter(&resultContent)
	table.SetHeader([]string{"Branch", "Status", "Reason"})
	for _, i := range result {
		s := fmt.Sprintf("%s %s", getStateEmoji(i.Status), i.Status)
		table.Append([]string{i.Branch, s, i.Reason})
	}
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.Render()

	tpl := template.Must(template.New("message").Parse(layout))
	data := Msg{
		Message: resultContent.String(),
	}
	err := tpl.Execute(&content, data)
	if err != nil {
		logrus.Errorf("Failed to execute NewResultComment err: %+v", err)
		return content.String(), err
	}
	return content.String(), nil
}

// NewSummaryComment NewSelectComment generate comment content
func NewSummaryComment(layout string, branches []string) (string, error) {
	var taskBranchLine strings.Builder
	var content strings.Builder
	type Msg struct {
		Message string `json:"message"`
	}
	for _, branch := range branches {
		taskBranchLine.WriteString("- [x] " + branch + "\n")
	}
	tpl := template.Must(template.New("message").Parse(layout))
	data := Msg{
		Message: taskBranchLine.String(),
	}
	err := tpl.Execute(&content, data)
	if err != nil {
		logrus.Warnf("Failed to execute template: %s \n branches: %s \n err: %+v", layout, branches, err)
		return content.String(), fmt.Errorf("failed to execute template: %w", err)
	}
	return content.String(), nil
}

// getStateEmoji returns the emoji for the state
func getStateEmoji(state Status) string {
	switch state {
	case SucceedStatus:
		return "✅"
	case FailedStatus:
		return "❌"
	case PendingStatus:
		return "⏳"
	case SkipStatus:
		return "⏭️"
	default:
		return "❓"
	}
}
