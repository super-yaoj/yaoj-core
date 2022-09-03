package analyzers

import (
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

type Analyzer = workflowruntime.Analyzer

// Try to display content of a text file with max-length limitation.
func show(store data.FileStore, title string, length int) workflow.ResultFile {
	bytes, _ := store.Get()
	content := string(bytes)
	if len(content) > length {
		content = content[:length]
	}
	return workflow.ResultFile{
		Title:   title,
		Content: content,
	}
}
