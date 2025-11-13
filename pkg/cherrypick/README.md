# Cherry-Pick Package

The `cherrypick` package provides functionality for cherry-picking commits across branches with customizable templates for generating comments.

## Features

- Cherry-pick commits from one branch to multiple target branches
- Configurable templates for summary and result comments
- Support for both GitHub and other Git providers
- Automatic conflict detection and reporting

## Usage

### Basic Usage with Default Templates

```go
import (
    "context"
    "github.com/chasenio/norn/pkg/cherrypick"
    "github.com/chasenio/norn/pkg/types"
)

// Create a service with default templates
service := cherrypick.NewService(provider)

// Create a cherry-pick task
task := &cherrypick.Task{
    Repo:           "owner/repo",
    Branches:       []string{"main", "release-1.0", "release-1.1"},
    From:           "main",
    SHA:            &commitSHA,
    MergeRequestID: "123",
    IsSummary:      true,
}

// Process the cherry-pick
err := service.CherryPick(context.Background(), task, nil)
```

### Using Custom Templates

You can provide custom Go templates for both summary and result comments. Templates must include a `{{ .Message }}` placeholder.

```go
// Create custom templates
templates := &cherrypick.Templates{
    SummaryTemplate: `
🍒 Cherry-pick Request
Please select branches to cherry-pick to:

{{ .Message }}

<!-- cherry-pick-summary -->
`,
    ResultTemplate: `
🎉 Cherry-pick Results

{{ .Message }}

<!-- cherry-pick-result -->
`,
}

// Create service with custom templates
service := cherrypick.NewServiceWithTemplates(provider, templates)

// Or use the helper to create templates with defaults fallback
templates := cherrypick.NewTemplates(summaryTemplate, resultTemplate)
service := cherrypick.NewServiceWithTemplates(provider, templates)
```

### CLI Usage

The cherry-pick functionality is available via CLI:

```bash
norn pick \
  --repo owner/repo \
  --token YOUR_TOKEN \
  --sha COMMIT_SHA \
  --for main \
  --merge-request-id 123 \
  --is-summary
```

Note: Template customization is available through the Go API only.

## Template Variables

Templates use Go's `text/template` syntax and receive the following data:

- `{{ .Message }}`: The formatted content (branch list for summary, result table for results)

## Default Templates

If no custom templates are provided, the following defaults are used:

**Summary Template:**
```
Will be cherry-picked to the following branches:

{{ .Message }}

<!-- Do not edit or delete , This is a cherry-pick summary flag. | o((>ω< ))o -->
```

**Result Template:**
```
Result: 
{{ .Message }}

<!-- Do not edit or delete , This is a cherry-pick result flag. | o((>ω< ))o -->
```

## API Reference

### NewService

```go
func NewService(provider types.Provider) *Service
```

Creates a new cherry-pick service with the given provider using default templates.

**Parameters:**
- `provider`: The git provider implementation (e.g., GitHub, GitLab)

**Returns:**
- `*Service`: A configured cherry-pick service instance with default templates

### NewServiceWithTemplates

```go
func NewServiceWithTemplates(provider types.Provider, templates *Templates) *Service
```

Creates a new cherry-pick service with custom templates.

**Parameters:**
- `provider`: The git provider implementation (e.g., GitHub, GitLab)
- `templates`: Custom template configuration (uses defaults if nil)

**Returns:**
- `*Service`: A configured cherry-pick service instance

### Service Methods

#### ProcessPick

```go
func (s *Service) ProcessPick(ctx context.Context, task *Task) error
```

Processes a cherry-pick task, either creating a summary or performing the actual cherry-pick.

#### PerformPickToBranches

```go
func (s *Service) PerformPickToBranches(ctx context.Context, task *Task, comment types.Comment) ([]*TaskResult, error)
```

Performs cherry-pick operations to multiple branches based on user selections.

#### CreateSummaryWithTask

```go
func (s *Service) CreateSummaryWithTask(ctx context.Context, task *Task) error
```

Creates or updates a cherry-pick summary comment on a merge request.

## Task Structure

```go
type Task struct {
    Repo           string   // Repository name (e.g., "owner/repo")
    Branches       []string // Target branches for cherry-picking
    From           string   // Source branch
    SHA            *string  // Commit SHA to cherry-pick
    MergeRequestID string   // Pull request or merge request ID
    IsSummary      bool     // Generate summary comment instead of performing pick
    PickMode       Mode     // Pick mode (MergeRequest or CheeryPick)
    RepoPath       string   // Local repository path
    BranchPrefix   string   // Prefix for temporary branches
}
```

## Examples

See the tests in `pick_test.go` and `helper_test.go` for more usage examples.
