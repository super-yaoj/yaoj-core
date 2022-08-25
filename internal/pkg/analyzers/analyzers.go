package analyzers

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/k0kubun/pp/v3"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"golang.org/x/text/encoding/charmap"
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

// DefaultAnalyzer 会首先分析哪些结点的信息与用户（Gsubm）有关，以此判定
// 是否出现系统错误。
//
// 它将关键结点的时空资源信息汇总做为结果的时空信息。
//
// 带有 compile、runner、checker:testlib 子串的结点会被特殊处理。
type DefaultAnalyzer struct {
}

func (r DefaultAnalyzer) Analyze(w *workflowruntime.RtWorkflow) workflow.Result {
	for _, bounds := range w.Inbound[workflow.Gsubm] {
		for _, bound := range bounds {
			w.RtNodes[bound.Name].Attr["dependon"] = "user"
		}
	}
	for {
		flag := false
		for _, edge := range w.Edge {
			if w.RtNodes[edge.From.Name].Attr["dependon"] == "user" &&
				w.RtNodes[edge.To.Name].Attr["dependon"] != "user" {

				w.RtNodes[edge.To.Name].Attr["dependon"] = "user"
				flag = true
			}
		}
		if !flag {
			break
		}
	}
	res := workflow.Result{
		ResultMeta: workflow.ResultMeta{
			Score:     w.Fullscore,
			Fullscore: w.Fullscore,
			Title:     "Accepted",
		},
		File: []workflow.ResultFile{},
	}

	// system error
	for name, node := range w.RtNodes {
		if node.Result == nil || node.Attr["dependon"] == "user" {
			continue
		}
		if node.Result.Code != processor.Ok {
			res.Title = "System Error"
			res.Score = 0
			res.File = append(res.File, workflow.ResultFile{
				Title:   "message",
				Content: name + ": " + node.Result.Msg,
			})
			res.File = append(res.File, autoFileDisplay(node)...)
			return res
		}
	}

	// compile error
	for name, node := range w.RtNodes {
		if node.Result == nil || node.Attr["dependon"] != "user" {
			continue
		}
		if node.Result.Code != processor.Ok && strings.Contains(node.ProcName, "compile") {
			res.Title = "Compile Error"
			res.Score = 0
			res.File = append(res.File, workflow.ResultFile{
				Title:   "message",
				Content: name + ": " + node.Result.Msg,
			})
			res.File = append(res.File, autoFileDisplay(node)...)
			return res
		}
	}

	// key node info
	for _, node := range w.RtNodes {
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
	for _, node := range w.RtNodes {
		if node.Result == nil {
			continue
		}
		if node.Attr["dependon"] == "user" {
			res.File = append(res.File, autoFileDisplay(node)...)
		}
	}
	nameOfCode := func(code processor.Code) string {
		switch code {
		case processor.Ok:
			return "Accepted"
		case processor.TimeExceed:
			return "Time Limit Exceed"
		case processor.RuntimeError:
			return "Runtime Error"
		case processor.DangerousSyscall:
			return "Dangerous System Call"
		case processor.ExitError:
			return "Exit Code Error"
		case processor.OutputExceed:
			return "Output Limit Exceed"
		case processor.MemoryExceed:
			return "Memory Limit Exceed"
		case processor.SystemError:
			return "System Error"
		}
		return "Unknown Error"
	}
	// runner error
	for _, node := range w.RtNodes {
		if node.Result == nil {
			continue
		}
		if node.Result.Code != processor.Ok {
			if node.Attr["dependon"] == "user" &&
				strings.Contains(node.ProcName, "runner") {

				res.Title = nameOfCode(node.Result.Code)
				res.Score = 0
				return res
			}
		}
	}
	// common error
	for _, node := range w.RtNodes {
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
				file, _ := node.Output["xmlreport"].File()
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
				file.Close()
				return res
			}
			if node.Attr["dependon"] == "user" {
				res.Title = nameOfCode(node.Result.Code)
				res.Score = 0
				return res
			}
			panic("system error")
		}
	}
	return res
}

// TODO: 优化草率的实现
func autoFileDisplay(node *workflowruntime.RtNode) []workflow.ResultFile {
	switch node.ProcName {
	case "checker:testlib":
		return []workflow.ResultFile{show(node.Input["answer"], "answer", 5000)}
	case "compiler:testlib", "compiler:auto":
		return []workflow.ResultFile{show(node.Output["log"], "compile log", 5000)}
	case "runner:auto":
		return []workflow.ResultFile{
			show(node.Output["stdout"], "stdout", 5000),
			show(node.Output["stderr"], "stderr", 5000),
		}
	default:
		return nil
	}
}

var _ workflowruntime.Analyzer = DefaultAnalyzer{}
