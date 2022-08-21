package workflow

import (
	"encoding/json"
	"os"
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// workflow 数据的来源（域）
type Groupname string

const (
	// 测试点的数据
	Gtests Groupname = "tests"
	// 整个题目的静态数据
	Gstatic Groupname = "static"
	// 参数者提交的数据
	Gsubm Groupname = "submission"
)

// Bound 可以理解为 Workflow 中结点的端口
type Bound struct {
	// name of the node
	Name string
	// label of the input/output store
	Label string
}

// Inbound 是读入数据的端口
type Inbound Bound

// Outboud 是输出数据的端口
type Outbound Bound

// Edge between nodes. Edges from field to node are stored in
// Workflow.Inbound.
type Edge struct {
	From Outbound
	To   Inbound
}

type Node struct {
	// processor name
	ProcName string
	// key node is attached importance by Analyzer
	Key bool
	// whether caching its result in global cache
	Cache bool
}

// store the file path of workflow's inbound data
type InboundGroups map[Groupname]map[string]data.FileStore

// Generate json content
/*
func (r *Workflow) Serialize() []byte {
	res, err := json.Marshal(*r)
	if err != nil {
		panic(err)
	}
	return res
}
*/

// Return all edges starting from Node[nodeid]
func (r *Workflow) EdgeFrom(name string) []Edge {
	res := []Edge{}
	for _, edge := range r.Edge {
		if edge.From.Name == name {
			res = append(res, edge)
		}
	}
	return res
}

// Return all edges ending at Node[nodeid]
func (r *Workflow) EdgeTo(name string) []Edge {
	res := []Edge{}
	for _, edge := range r.Edge {
		if edge.To.Name == name {
			res = append(res, edge)
		}
	}
	return res
}

// Load graph from serialized data (json)
func Load(serial []byte) (*Workflow, error) {
	var graph Workflow
	err := json.Unmarshal(serial, &graph)
	if err != nil {
		return nil, err
	}
	return &graph, nil
}

// Load graph from (json) file.
func LoadFile(path string) (*Workflow, error) {
	serial, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Load(serial)
}

// workflow describes how to perform a single testcase's judgement
//
// json marshalable
type Workflow struct {
	// a node itself is just a processor
	Node map[string]Node
	Edge []Edge
	// inbound consists a series of data group.
	// Inbound: map[datagroup_name]map[field]Bound
	Inbound map[Groupname]map[string][]Inbound
}

type ResultMeta struct {
	// e. g. "Accepted", "Wrong Answer"
	Title     string
	Score     float64
	Fullscore float64
	Time      time.Duration
	Memory    utils.ByteValue
}

// Result of a workflow, typically generated by Analyzer.
type Result struct {
	ResultMeta
	// a list of file content to display
	File []ResultFileDisplay
}

// json content
func (r *Result) Byte() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

type ResultFileDisplay struct {
	Title   string
	Content string
}

// Create an empty Workflow
func New() *Workflow {
	return &Workflow{
		Node:    map[string]Node{},
		Edge:    []Edge{},
		Inbound: map[Groupname]map[string][]Inbound{},
	}
}
