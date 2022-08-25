package utils

import (
	"errors"
)

var ErrNotDAG = errors.New("not a DAG")

// another version of topsort
//
// slice: node key array (key should be comparable)
//
// hasEdgeFromTo: whether theres an edge from u to v.
//
// Complexity: O(n^2)
func TopSort[K comparable](nodes []K, hasEdgeFromTo func(u, v K) bool) (res []K, err error) {
	n := len(nodes)
	indegree := map[K]int{}
	for _, u := range nodes {
		for _, v := range nodes {
			if u != v && hasEdgeFromTo(u, v) {
				indegree[v]++
			}
		}
	}
	res = make([]K, 0, n)
	err = nil
	for {
		pre := len(res)
		for _, u := range nodes {
			if indegree[u] == 0 {
				res = append(res, u)
				indegree[u] = -1
			}
		}
		if pre == len(res) {
			break
		}
		for id := pre; id < len(res); id++ {
			u := res[id]
			for _, v := range nodes {
				if u != v && hasEdgeFromTo(u, v) {
					if indegree[v] < 0 {
						panic("topo sort error")
					}
					indegree[v]--
				}
			}
		}
	}
	if len(res) != len(nodes) {
		err = ErrNotDAG
	}
	return
}
