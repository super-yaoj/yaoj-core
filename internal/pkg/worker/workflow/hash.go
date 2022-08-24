package workflowruntime

import (
	"crypto/sha256"
	"fmt"
	"hash"
)

type shaHash struct {
	hash.Hash
}

// convert hash.Hash to SHA aka [32]byte
func (r *shaHash) SHA() (res SHA) {
	var b = r.Sum(nil)
	if len(b) != 32 {
		panic(b)
	}
	res = *(*SHA)(b)
	return
}

func (r *shaHash) WriteString(s string) (n int, err error) {
	return r.Write([]byte(s))
}

func newShaHash() *shaHash {
	return &shaHash{Hash: sha256.New()}
}

// 使用长度为 32 的字节数组存储 SHA 值
type SHA [32]byte

func (r SHA) String() string {
	s := ""
	for _, v := range r {
		s += fmt.Sprintf("%02x", v)
	}
	return s
}
