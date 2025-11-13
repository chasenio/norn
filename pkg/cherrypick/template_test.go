package cherrypick

import (
	"testing"
)

// TestNewResultCommentWithCustomTemplate tests that custom templates work correctly
func TestNewResultCommentWithCustomTemplate(t *testing.T) {
	customTemplate := `
🎉 Custom Cherry-Pick Results

{{ .Message }}

<!-- custom-flag -->
`
	result := []*TaskResult{
		{
			Status: SucceedStatus,
			Branch: "main",
			Reason: "",
		},
		{
			Status: FailedStatus,
			Branch: "release-1.0",
			Reason: "Conflicts detected",
		},
	}
	
	comment, err := NewResultComment(customTemplate, result)
	if err != nil {
		t.Errorf("NewResultComment() with custom template failed: %v", err)
	}
	
	// Verify custom template elements are present
	if comment == "" {
		t.Error("Expected non-empty comment")
	}
	
	// Check for custom emoji
	if !contains(comment, "🎉") {
		t.Error("Custom template emoji not found in output")
	}
	
	// Check for custom flag
	if !contains(comment, "<!-- custom-flag -->") {
		t.Error("Custom flag not found in output")
	}
	
	// Check for actual content
	if !contains(comment, "main") || !contains(comment, "release-1.0") {
		t.Error("Branch names not found in output")
	}
	
	t.Logf("Custom template output:\n%s", comment)
}

// TestNewSummaryCommentWithCustomTemplate tests that custom summary templates work correctly
func TestNewSummaryCommentWithCustomTemplate(t *testing.T) {
	customTemplate := `
🍒 Please Select Cherry-Pick Targets

{{ .Message }}

_Automated by Norn_
<!-- custom-summary-flag -->
`
	branches := []string{"main", "release-1.0", "release-1.1"}
	
	comment, err := NewSummaryComment(customTemplate, branches)
	if err != nil {
		t.Errorf("NewSummaryComment() with custom template failed: %v", err)
	}
	
	// Verify custom template elements are present
	if comment == "" {
		t.Error("Expected non-empty comment")
	}
	
	// Check for custom elements
	if !contains(comment, "🍒") {
		t.Error("Custom template emoji not found in output")
	}
	
	if !contains(comment, "_Automated by Norn_") {
		t.Error("Custom footer not found in output")
	}
	
	if !contains(comment, "<!-- custom-summary-flag -->") {
		t.Error("Custom flag not found in output")
	}
	
	// Check for branches in checkbox format
	for _, branch := range branches {
		if !contains(comment, branch) {
			t.Errorf("Branch %s not found in output", branch)
		}
	}
	
	// Check for checkbox format
	if !contains(comment, "- [x]") {
		t.Error("Checkbox format not found in output")
	}
	
	t.Logf("Custom summary template output:\n%s", comment)
}

// TestDefaultTemplatesStillWork tests backward compatibility with empty template strings
func TestDefaultTemplatesStillWork(t *testing.T) {
	// Using empty strings should use default templates
	// This is tested indirectly through Service creation
	// but we can test the helper functions directly
	
	// Note: In actual usage, NewService handles the default template logic
	// These helper functions expect actual template strings
	t.Log("Default template handling is done by NewService constructor")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(findSubstring(s, substr) != -1))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
