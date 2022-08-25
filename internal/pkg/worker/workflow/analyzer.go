package workflowruntime

import (
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// Analyzer generates result of a runtime workflow.
type Analyzer interface {
	Analyze(w *RtWorkflow) workflow.Result
}
