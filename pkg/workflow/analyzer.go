package workflow

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	goPlugin "plugin"
	"strings"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"golang.org/x/text/encoding/charmap"
)

// Analyzer generates result of a workflow.
type Analyzer interface {
	Analyze(w Workflow, nodes map[string]RuntimeNode, fullscore float64) Result
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

func (r DefaultAnalyzer) Analyze(w Workflow, nodes map[string]RuntimeNode, fullscore float64) Result {
	for _, bounds := range *w.Inbound[Gsubm] {
		for _, bound := range bounds {
			nodes[bound.Name].Attr["dependon"] = "user"
		}
	}
	for {
		flag := false
		for _, edge := range w.Edge {
			if nodes[edge.From.Name].Attr["dependon"] == "user" &&
				nodes[edge.To.Name].Attr["dependon"] != "user" {

				nodes[edge.To.Name].Attr["dependon"] = "user"
				flag = true
			}
		}
		if !flag {
			break
		}
	}
	res := Result{
		ResultMeta: ResultMeta{
			Score:     fullscore,
			Fullscore: fullscore,
			Title:     "Accepted",
		},
		File: []ResultFileDisplay{},
	}

	// system error
	for name, node := range nodes {
		if node.Result == nil || node.Attr["dependon"] == "user" {
			continue
		}
		if node.Result.Code != processor.Ok {
			res.Title = "System Error"
			res.Score = 0
			res.File = append(res.File, ResultFileDisplay{
				Title:   "message",
				Content: name + ": " + node.Result.Msg,
			})
			res.File = append(res.File, autoFileDisplay(node)...)
			return res
		}
	}

	// compile error
	for name, node := range nodes {
		if node.Result == nil || node.Attr["dependon"] != "user" {
			continue
		}
		if node.Result.Code != processor.Ok && strings.Contains(node.ProcName, "compile") {
			res.Title = "Compile Error"
			res.Score = 0
			res.File = append(res.File, ResultFileDisplay{
				Title:   "message",
				Content: name + ": " + node.Result.Msg,
			})
			res.File = append(res.File, autoFileDisplay(node)...)
			return res
		}
	}

	// key node info
	for _, node := range nodes {
		if node.Result == nil {
			continue
		}

		if node.Key {
			if node.Result.Memory != nil {
				res.ResultMeta.Memory += utils.ByteValue(*node.Result.Memory)
			}
			if node.Result.CpuTime != nil {
				res.ResultMeta.Time += *node.Result.CpuTime
			}
		}
	}

	// common files
	for _, node := range nodes {
		if node.Result == nil {
			continue
		}
		if node.Attr["dependon"] == "user" {
			res.File = append(res.File, autoFileDisplay(node)...)
		}
	}
	// common error
	for _, node := range nodes {
		if node.Result == nil {
			continue
		}
		if node.Result.Code != processor.Ok {
			if node.ProcName == "checker:testlib" {
				type Result struct {
					XMLName xml.Name `xml:"result"`
					Outcome string   `xml:"outcome,attr"`
				}
				var result Result
				file, _ := os.Open(node.Output[0])
				// parse xml encoded windows1251
				d := xml.NewDecoder(file)
				d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
					switch charset {
					case "windows-1251":
						return charmap.Windows1251.NewDecoder().Reader(input), nil
					default:
						return nil, fmt.Errorf("unknown charset: %s", charset)
					}
				}
				d.Decode(&result)

				if result.Outcome != "accepted" {
					res.Title = "Wrong Answer"
					res.Score = 0
					pp.Print(result)
				}
				return res
			}
			if node.Attr["dependon"] == "user" {
				var title = "Unaccepted"
				switch node.Result.Code {
				case processor.TimeExceed:
					title = "Time Limit Exceed"
				case processor.RuntimeError:
					title = "Runtime Error"
				case processor.DangerousSyscall:
					title = "Dangerous System Call"
				case processor.ExitError:
					title = "Exit Code Error"
				case processor.OutputExceed:
					title = "Output Limit Exceed"
				case processor.MemoryExceed:
					title = "Memory Limit Exceed"
				}

				res.Title = title
				res.Score = 0
				return res
			}
			panic("system error")
		}
	}
	return res
}

func autoFileDisplay(node RuntimeNode) []ResultFileDisplay {
	switch node.ProcName {
	case "checker:hcmp":
		return []ResultFileDisplay{FileDisplay(node.Input[1], "answer", 5000)}
	case "checker:testlib":
		return []ResultFileDisplay{FileDisplay(node.Input[3], "answer", 5000)}
	case "compiler", "compiler:testlib", "compiler:auto":
		return []ResultFileDisplay{FileDisplay(node.Output[1], "compile log", 5000)}
	case "inputmaker":
		return []ResultFileDisplay{FileDisplay(node.Input[0], "input source", 5000)}
	case "generator:testlib":
		return []ResultFileDisplay{FileDisplay(node.Input[1], "generator arguments", 5000)}
	case "runner:fileio", "runner:stdio":
		return []ResultFileDisplay{FileDisplay(node.Output[0], "stdout", 5000), FileDisplay(node.Output[1], "stderr", 5000)}
	default:
		return nil
	}
}

var _ Analyzer = DefaultAnalyzer{}
