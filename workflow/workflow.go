package workflow

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sshwy/yaoj-core/processor"
	"github.com/sshwy/yaoj-core/utils"
)

type Bound struct {
	// index of the node in the array
	Index int
	// index of the file in input (output) array
	LabelIndex int
}

// When it comes to out edge, label denotes output label and bound denotes the destination.
// Otherwise (in edge), label denotes input label and bound denotes the origin.
// Actually an edge is just a file in os.
// If a Edge has no Bound, it should be workflow inbound(outbound) edge.
type Edge struct {
	Bound *Bound
	Label string
}

type Node struct {
	// processor name
	ProcName string
	InEdge   []Edge
	OutEdge  []Edge
}

func (r *Node) Processor() processor.Processor {
	return processor.Get(r.ProcName)
}

// workflow describes how to perform a single testcase's judgement
type Workflow struct {
	// a node itself is just a processor
	Node []Node
	// inbound consists a series of data group
	Inbound []DataBoundGroup
}

// check whether it's a well-formatted DAG, its inbound coverage and sth else
func (r *Workflow) Valid() error {
	var inboundCnt int
	for i, node := range r.Node {
		proc := node.Processor()
		if proc == nil {
			return fmt.Errorf("node[%d] has invalid processor name (%s)", i, node.ProcName)
		}
		inLabel, outLabel := proc.Label()
		if len(node.InEdge) != len(inLabel) || len(node.OutEdge) != len(outLabel) {
			return fmt.Errorf("node[%d] has invalid number of in edge or out edge", i)
		}
		for j, edge := range node.InEdge {
			if inLabel[j] != edge.Label {
				return fmt.Errorf("node[%d]'s InEdge[%d] has invalid label %s, expect %s", i, j, edge.Label, inLabel[j])
			}
			if edge.Bound == nil {
				inboundCnt++
				continue
			}
			if edge.Bound.Index >= i || edge.Bound.Index < 0 {
				return fmt.Errorf("node[%d] has invalid InEdge %+v", i, edge)
			}
		}
		for j, edge := range node.OutEdge {
			if outLabel[j] != edge.Label {
				return fmt.Errorf("node[%d]'s OutEdge[%d] has invalid label %s, expect %s", i, j, edge.Label, outLabel[j])
			}
			if edge.Bound == nil {
				continue
			}
			if edge.Bound.Index <= i || edge.Bound.Index >= len(r.Node) {
				return fmt.Errorf("node[%d] has invalid OutEdge %+v", i, edge)
			}
		}
	}
	for i, group := range r.Inbound {
		inboundCnt -= len(group)
		for j, bound := range group {
			if bound.Bound.Index >= len(r.Node) {
				return fmt.Errorf("inbound[%d][%d] has invalid node index %d", i, j, bound.Bound.Index)
			}
			node := r.Node[bound.Bound.Index]
			if bound.Bound.LabelIndex >= len(node.InEdge) {
				return fmt.Errorf("inbound[%d][%d] has invalid node label index %d", i, j, bound.Bound.LabelIndex)
			}
			if node.InEdge[bound.Bound.LabelIndex].Bound != nil {
				return fmt.Errorf("node[%d].InEdge[%d] conflict with Inbound[%d][%d]",
					bound.Bound.Index, bound.Bound.LabelIndex, i, j)
			}
		}
	}
	if inboundCnt != 0 {
		return fmt.Errorf("invalid inbound num (diff=%d)", inboundCnt)
	}
	return nil
}

// transform to dot file content
// func (r *Workflow) Dot() string

// parse dot file to workflow
// func (r *Workflow) ParseDot(content string) error

// use a string represents a data field
type DataLabel string

// connect data with bound
type DataBound struct {
	Data  DataLabel
	Bound Bound
}

// a series of data
type DataBoundGroup []DataBound

type Result struct {
	Score     float64
	Fullscore float64
	Time      time.Duration
	Memory    utils.ByteValue
	// e. g. "Accepted", "Wrong Answer"
	Title string
	// a list of file content to display
	File []ResultFileDisplay
	// other tags
	Property map[string]string
}

type ResultFileDisplay struct {
	Title   string
	Content string
}

func fetchFileContent(path string, len int) []byte {
	file, err := os.Open(path)
	if err != nil {
		return []byte("[error] " + err.Error())
	}
	defer file.Close()
	b := make([]byte, len)
	file.Read(b)
	return b
}

func FileDisplay(path string, title string, len int) ResultFileDisplay {
	content := strings.TrimRight(string(fetchFileContent(path, len)), "\x00 \n\t\r")
	return ResultFileDisplay{
		Title:   title,
		Content: content,
	}
}
