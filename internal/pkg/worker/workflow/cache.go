package workflowruntime

import (
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

// 针对 workflow 的结点的输出结果的缓存
//
// 利用已经计算好的 hash 值寻找缓存
//
// 缓存的值包括：Output, Result
type RtNodeCache interface {
	// add node to cache (by hash)
	//
	// 具体是否加入缓存取决于 node 本身
	Add(node *RtNode) error
	// check if cache exist
	Exist(node *RtNode) bool
	// assign cache to node
	Assign(node *RtNode) error
}

type GlobalCache struct {
	// 所有缓存数据的存放位置
	dir   string
	store map[string]struct{}
}

func (r *GlobalCache) Add(node *RtNode) error {
	if !node.Cache { // 不缓存
		return nil
	}

	key := node.Hash().String()
	err := os.WriteFile(path.Join(r.dir, key+".result"), node.Result.Serialize(), 0777)
	if err != nil {
		return err
	}
	for field, store := range node.Output {
		ctnt, err := store.Get()
		if err != nil {
			return err
		}
		err = os.WriteFile(path.Join(r.dir, key+field), ctnt, 0777)
		if err != nil {
			return err
		}
	}
	r.store[key] = struct{}{}
	return nil
}

func (r *GlobalCache) Exist(node *RtNode) bool {
	_, exist := r.store[node.Hash().String()]
	return exist
}

func (r *GlobalCache) Assign(node *RtNode) error {
	key := node.Hash().String()
	node.Output = make(processor.Outbounds)
	for _, field := range processor.OutputLabel(node.ProcName) {
		node.Output[field] = data.NewFlexFile(path.Join(r.dir, key+field))
	}
	res_ctnt, err := os.ReadFile(path.Join(r.dir, key+".result"))
	if err != nil {
		return err
	}
	node.Result = &processor.Result{}
	err = node.Result.Deserialize(res_ctnt)
	if err != nil {
		return err
	}
	return nil
}

// create dir if necessary
func NewCache(dir string) (*GlobalCache, error) {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return nil, err
	}
	return &GlobalCache{dir: dir, store: map[string]struct{}{}}, nil
}
