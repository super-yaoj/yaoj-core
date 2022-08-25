package analyzers

import (
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// workflow result builder
type Builder workflow.Result

func (r Builder) SetTitle(title string) Builder {
	r.Title = title
	return r
}
func (r Builder) SetScore(score float64) Builder {
	r.Score = score
	return r
}
func (r Builder) SetFullscore(score float64) Builder {
	r.Fullscore = score
	return r
}
func (r Builder) SetUsage(time time.Duration, memory utils.ByteValue) Builder {
	r.Time = time
	r.Memory = memory
	return r
}

func (r Builder) Show(store data.FileStore, title string, length int) Builder {
	r.File = append(r.File, show(store, title, length))
	return r
}

func (r Builder) Build() workflow.Result {
	return workflow.Result(r)
}
