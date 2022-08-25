package analyzers

import (
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 用于在第一轮 std 执行完后获取数据
type Hack struct {
	iomap map[string]workflow.Outbound
	data  map[string]data.FileStore
}

func (r *Hack) Analyze(w *workflowruntime.RtWorkflow) workflow.Result {
	r.data = make(map[string]data.FileStore)
	for field, bound := range r.iomap {
		r.data[field] = w.RtNodes[bound.Name].Output[bound.Label]
	}
	return workflow.Result{}
}

var _ workflowruntime.Analyzer = (*Hack)(nil)
