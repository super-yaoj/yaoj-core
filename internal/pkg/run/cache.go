package run

import (
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// submission-wide cache
type inmemoryStore[T any] struct {
	data map[sha]T
}

func (r *inmemoryStore[T]) Set(hash sha, outputs T) {
	r.data[hash] = outputs
}
func (r *inmemoryStore[T]) Get(hash sha) T {
	return r.data[hash]
}
func (r *inmemoryStore[T]) Has(hash sha) bool {
	_, ok := r.data[hash]
	return ok
}
func (r *inmemoryStore[T]) Reset() {
	r.data = map[sha]T{}
}

// Global cache
type fsCache struct {
	dir       string
	existence map[string]bool
	list      []string
}

func (r *fsCache) assign(key string) {
	if r.existence[key] { // remove it
		logger.Printf("warning: %q already cached", key)
		id := utils.FindIndex(r.list, key)
		if id == -1 {
			panic("key exist but not found in list")
		}
		r.list = append(r.list[:id], r.list[id+1:]...)
	}
	r.existence[key] = true
	r.list = append(r.list, key)
}
func (r *fsCache) Set(hash sha, decorator string, data []byte) {
	key := hash.String() + decorator
	file, _ := os.Create(path.Join(r.dir, key))
	file.Write(data)
	file.Close()

	r.assign(key)
}
func (r *fsCache) SetSource(hash sha, decorator string, name string) {
	key := hash.String() + decorator
	utils.CopyFile(name, path.Join(r.dir, key))

	r.assign(key)
}
func (r *fsCache) Get(hash sha, decorator string) []byte {
	key := hash.String() + decorator
	data, _ := os.ReadFile(path.Join(r.dir, key))
	return data
}
func (r *fsCache) GetSource(hash sha, decorator string) string {
	key := hash.String() + decorator
	return path.Join(r.dir, key)
}

func (r *fsCache) Has(hash sha, decorator string) bool {
	key := hash.String() + decorator
	return r.existence[key]
}
func (r *fsCache) Reset() {
	os.RemoveAll(r.dir)
	os.MkdirAll(r.dir, os.ModePerm)
	r.existence = map[string]bool{}
	r.list = make([]string, 0)
}
func (r *fsCache) Resize(size int) {
	logger.Printf("global cache trying to resize")
	resizeMutex.Lock()
	defer resizeMutex.Unlock()

	if len(r.list) > size {
		remlist := r.list[:len(r.list)-size]

		for _, key := range remlist {
			delete(r.existence, key)
			os.RemoveAll(path.Join(r.dir, key))
		}

		r.list = r.list[len(r.list)-size:]
	}
}

var gcache *fsCache

// make sure init cache before problem judging!
func CacheInit(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	logger.Printf("cache data in %q", dir)
	if gcache != nil {
		gcache.Reset()
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	gcache = &fsCache{
		dir:       dir,
		existence: map[string]bool{},
		list:      make([]string, 0),
	}
	return nil
}

// 防止在评测题目的时候 resize
var resizeMutex sync.Mutex
var CacheSize = 1000
