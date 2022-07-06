package run

import (
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

type InMemoryCache[T any] struct {
	data map[sha]T
}

func (r *InMemoryCache[T]) Set(hash sha, outputs T) {
	r.data[hash] = outputs
}
func (r *InMemoryCache[T]) Get(hash sha) T {
	return r.data[hash]
}
func (r *InMemoryCache[T]) Has(hash sha) bool {
	_, ok := r.data[hash]
	return ok
}
func (r *InMemoryCache[T]) Reset() {
	r.data = map[sha]T{}
}

var gOutputCache = InMemoryCache[[]string]{
	data: map[sha][]string{},
}
var gResultCache = InMemoryCache[processor.Result]{
	data: map[sha]processor.Result{},
}
