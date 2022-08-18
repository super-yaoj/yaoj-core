package workflow

import (
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Builder builds a workflow. It doesn't need initialization manually
type Builder struct {
	node          map[string]Node
	inbound, edge [][4]string
	err           error
}

func (r *Builder) tryInit() {
	if r.node == nil {
		r.node = map[string]Node{}
	}
	if r.edge == nil {
		r.edge = [][4]string{}
	}
	if r.inbound == nil {
		r.inbound = [][4]string{}
	}
}

// Add a node to the workflow.
//
// procName: specify its processor.
//
// key: whether its a key node.
//
// cache: whether caching its result in global cache.
func (r *Builder) SetNode(name string, procName string, key bool, cache bool) {
	r.tryInit()
	r.node[name] = Node{
		ProcName: procName,
		Key:      key,
		Cache:    cache,
	}
}

func (r *Builder) AddEdge(from, frlabel, to, tolabel string) {
	r.tryInit()
	r.edge = append(r.edge, [4]string{from, frlabel, to, tolabel})
}

type Groupname string

const (
	Gtests  Groupname = "tests"
	Gsubt   Groupname = "Subtask"
	Gstatic Groupname = "static"
	Gsubm   Groupname = "submission"
)

func (r *Builder) AddInbound(group Groupname, field, to, tolabel string) {
	r.tryInit()
	if group != Gtests && group != Gstatic && group != Gsubm && group != Gsubt {
		r.err = &Error{"AddInbound", ErrInvalidGroupname}
		return
	}
	r.inbound = append(r.inbound, [4]string{string(group), field, to, tolabel})
}

func (r *Builder) WorkflowGraph() (*WorkflowGraph, error) {
	if r.err != nil {
		return nil, r.err
	}
	graph := NewGraph()
	for name, node := range r.node {
		graph.Node[name] = node
	}

	var vis = map[string]map[string]bool{}
	mark := func(name string, label string) {
		if vis[name] == nil {
			vis[name] = map[string]bool{}
		}
		vis[name][label] = true
	}
	get := func(name string, label string) bool {
		if vis[name] == nil {
			return false
		}
		return vis[name][label]
	}

	for _, edge := range r.edge {
		from, frlabel, to, tolabel := edge[0], edge[1], edge[2], edge[3]
		frlabelIndex := idxOf(processor.OutputLabel(graph.Node[from].ProcName), frlabel)
		tolabelIndex := idxOf(processor.InputLabel(graph.Node[to].ProcName), tolabel)

		if _, ok := graph.Node[from]; !ok {
			return nil, &DataError{edge, ErrInvalidEdge}
		} else if _, ok := graph.Node[to]; !ok {
			return nil, &DataError{edge, ErrInvalidEdge}
		} else if frlabelIndex == -1 {
			return nil, &DataError{edge, ErrInvalidOutputLabel}
		} else if tolabelIndex == -1 {
			return nil, &DataError{edge, ErrInvalidInputLabel}
		}

		if get(to, tolabel) {
			return nil, &DataError{edge, ErrDuplicateDest}
		} else {
			mark(to, tolabel)
		}
		graph.Edge = append(graph.Edge, Edge{
			Outbound{from, frlabel},
			Inbound{to, tolabel},
		})
	}
	for _, edge := range r.inbound {
		group, field, to, tolabel := edge[0], edge[1], edge[2], edge[3]
		tolabelIndex := idxOf(processor.InputLabel(graph.Node[to].ProcName), tolabel)

		if _, ok := graph.Node[to]; !ok {
			return nil, &DataError{edge, ErrInvalidEdge}
		} else if tolabelIndex == -1 {
			return nil, &DataError{edge, ErrInvalidInputLabel}
		}

		if get(to, tolabel) {
			return nil, &DataError{edge, ErrDuplicateDest}
		} else {
			mark(to, tolabel)
		}
		if graph.Inbound[Groupname(group)] == nil {
			graph.Inbound[Groupname(group)] = map[string][]Inbound{}
		}
		grp := graph.Inbound[Groupname(group)]
		if grp[field] == nil {
			grp[field] = []Inbound{}
		}
		grp[field] = append(grp[field], Inbound{to, tolabel})
	}
	for name, node := range graph.Node {
		for _, label := range processor.InputLabel(node.ProcName) {
			if !get(name, label) {
				return nil, &DataError{name + ":" + label, ErrIncompleteNodeInput}
			}
		}
	}
	return &graph, nil
}

func idxOf(s []string, t string) int {
	return utils.FindIndex(s, t)
}
