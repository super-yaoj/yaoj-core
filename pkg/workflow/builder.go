package workflow

import (
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

// Builder builds a workflow. It doesn't need initialization manually
type Builder struct {
	Nodes    map[string]Node `json:"nodes"`
	Inbounds [][4]string     `json:"inbounds"`
	Edges    [][4]string     `json:"edges"`
	err      error
}

func (r *Builder) tryInit() {
	if r.Nodes == nil {
		r.Nodes = map[string]Node{}
	}
	if r.Edges == nil {
		r.Edges = [][4]string{}
	}
	if r.Inbounds == nil {
		r.Inbounds = [][4]string{}
	}
}

// Add a node to the workflow.
//
// procName: specify its processor.
//
// key: whether its a key node. (deprecated)
//
// cache: whether caching its result in global cache.
func (r *Builder) SetNode(name string, procName string, key bool, cache bool) {
	r.tryInit()
	r.Nodes[name] = Node{
		ProcName: procName,
		Cache:    cache,
	}
}

func (r *Builder) AddEdge(from, frlabel, to, tolabel string) {
	r.tryInit()
	r.Edges = append(r.Edges, [4]string{from, frlabel, to, tolabel})
}

func (r *Builder) AddInbound(group Groupname, field, to, tolabel string) {
	r.tryInit()
	if group != Gtests && group != Gstatic && group != Gsubm {
		r.err = yerrors.Situated("Builder.AddInbound", ErrInvalidGroupname)
		return
	}
	r.Inbounds = append(r.Inbounds, [4]string{string(group), field, to, tolabel})
}

func (r *Builder) Workflow() (*Workflow, error) {
	if r.err != nil {
		return nil, r.err
	}
	graph := New()
	for name, node := range r.Nodes {
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

	for _, edge := range r.Edges {
		from, frlabel, to, tolabel := edge[0], edge[1], edge[2], edge[3]
		frlabelIndex := idxOf(processor.OutputLabel(graph.Node[from].ProcName), frlabel)
		tolabelIndex := idxOf(processor.InputLabel(graph.Node[to].ProcName), tolabel)

		if _, ok := graph.Node[from]; !ok {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidEdge)
		} else if _, ok := graph.Node[to]; !ok {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidEdge)
		} else if frlabelIndex == -1 {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidOutputLabel)
		} else if tolabelIndex == -1 {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidInputLabel)
		}

		if get(to, tolabel) {
			return nil, yerrors.Annotated("edge", edge, ErrDuplicateDest)
		} else {
			mark(to, tolabel)
		}
		graph.Edge = append(graph.Edge, Edge{
			Outbound{from, frlabel},
			Inbound{to, tolabel},
		})
	}
	for _, edge := range r.Inbounds {
		group, field, to, tolabel := edge[0], edge[1], edge[2], edge[3]
		tolabelIndex := idxOf(processor.InputLabel(graph.Node[to].ProcName), tolabel)

		if _, ok := graph.Node[to]; !ok {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidEdge)
		} else if tolabelIndex == -1 {
			return nil, yerrors.Annotated("edge", edge, ErrInvalidInputLabel)
		}

		if get(to, tolabel) {
			return nil, yerrors.Annotated("edge", edge, ErrDuplicateDest)
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
				return nil, yerrors.Annotated(name, label, ErrIncompleteNodeInput)
			}
		}
	}
	return graph, nil
}

func idxOf(s []string, t string) int {
	return utils.FindIndex(s, t)
}
