package run

// submission-wide cache
type inMemoryCache[T any] struct {
	data map[sha]T
}

func (r *inMemoryCache[T]) Set(hash sha, outputs T) {
	r.data[hash] = outputs
}
func (r *inMemoryCache[T]) Get(hash sha) T {
	return r.data[hash]
}
func (r *inMemoryCache[T]) Has(hash sha) bool {
	_, ok := r.data[hash]
	return ok
}
func (r *inMemoryCache[T]) Reset() {
	r.data = map[sha]T{}
}
