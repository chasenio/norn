package pick

import (
	"crypto/md5"
	"fmt"
	tp "github.com/kentio/norn/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

func sumMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// parseSelectedBranches parse selected branches from comment
func parseSelectedBranches(comment string) (selectedBranches []string) {
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
			selectedBranches = append(selectedBranches, line)
		}
	}
	return selectedBranches
}

func generateTargetBranches(task *Task) []string {
	var targets []string
	var startFlag bool
	for _, branch := range task.Branches {
		// Skip branches before the 'From' branch in the list
		if branch == task.Form {
			startFlag = true
			continue
		}
		if startFlag {
			targets = append(targets, branch)
		} else {
			logrus.Debugf("Skipping branch: %s, from: %s", branch, task.Form)
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
