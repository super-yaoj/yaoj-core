package workflow

import (
	"fmt"
	goPlugin "plugin"

	"github.com/sshwy/yaoj-core/judger"
)

// Analyzer generates result of a workflow.
type Analyzer interface {
	Analyze(nodes []RuntimeNode, fullscore float64) Result
}

func LoadAnalyzer(plugin string) (Analyzer, error) {
	p, err := goPlugin.Open(plugin)
	if err != nil {
		return nil, err
	}

	label, err := p.Lookup("AnalyzerPlugin")
	if err != nil {
		return nil, err
	}
	analyzer, ok := label.(*Analyzer)
	if ok {
		return *analyzer, nil
	} else {
		return nil, fmt.Errorf("AnalyzerPlugin not implement Analyzer")
	}
}

type DefaultAnalyzer struct {
}

func (r DefaultAnalyzer) Analyze(nodes []RuntimeNode, fullscore float64) Result {
	for i, node := range nodes {
		if node.Result == nil {
			continue
		}
		if node.Result.Code != judger.Ok {
			return Result{
				Score:     0,
				Fullscore: fullscore,
				Title:     "Not Accepted",
				File: []ResultFileDisplay{
					{
						Title:   "Error Node",
						Content: fmt.Sprintf("id=%d, proc=%s, Code=%v %s", i, node.ProcName, node.Result.Code, nodes[0].Output[1]),
					},
				},
			}
		}
	}
	return Result{
		Score:     fullscore,
		Fullscore: fullscore,
		Title:     "Accepted",
	}
}

var _ Analyzer = DefaultAnalyzer{}
