package judgeserver

import (
	"sync"

	"github.com/super-yaoj/yaoj-core/pkg/problem"
)

// concurrent-safe storage
//
// 目前的 store 过于简陋，没有考虑到题目长时间不评测的空间回收问题
type Storage struct {
	Map sync.Map
}

func (r *Storage) Has(checksum string) bool {
	// logger.Printf("has %s", checksum)
	_, ok := r.Map.Load(checksum)
	return ok
}
func (r *Storage) Set(checksum string, prob *problem.Data) {
	// logger.Printf("set %s", checksum)
	r.Map.Store(checksum, prob)
}
func (r *Storage) Get(checksum string) *problem.Data {
	val, _ := r.Map.Load(checksum)
	return val.(*problem.Data)
}
