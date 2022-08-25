package workflowruntime

import (
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// Analyzer generates result of a runtime workflow.
type Analyzer interface {
	Analyze(w *RtWorkflow) workflow.Result
}

// func LoadAnalyzer(plugin string) (Analyzer, error) {
// 	p, err := goPlugin.Open(plugin)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	label, err := p.Lookup("AnalyzerPlugin")
// 	if err != nil {
// 		return nil, err
// 	}
// 	analyzer, ok := label.(*Analyzer)
// 	if ok {
// 		return *analyzer, nil
// 	} else {
// 		return nil, fmt.Errorf("AnalyzerPlugin not implement Analyzer")
// 	}
// }

// Try to display content of a text file with max-length limitation.
func FileDisplay(store data.FileStore, title string, length int) workflow.ResultFile {
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
