package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	yutils "github.com/super-yaoj/yaoj-utils"
)

type HashValue = yutils.HashValue

type ByteValue = yutils.ByteValue

func Map[T any, M any](s []T, f func(T) M) []M {
	var a []M = make([]M, len(s))
	for i, v := range s {
		a[i] = f(v)
	}
	return a
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func CopyFile(src, dst string) (int64, error) {
	// log.Printf("CopyFile %s %s", src, dst)
	if src == dst {
		return 0, fmt.Errorf("same path")
	}
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func ReaderChecksum(reader io.Reader) Checksum {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return Checksum{}
	}
	var b = hash.Sum(nil)
	if len(b) != 32 {
		panic(b)
	}
	return *(*Checksum)(b)
}

// SHA256 hash for file content.
// for any error, return empty hash
func FileChecksum(name string) Checksum {
	f, err := os.Open(name)
	if err != nil {
		return Checksum{}
	}
	defer f.Close()

	return ReaderChecksum(f)
}

// comparable
type Checksum = yutils.Checksum

type LangTag = yutils.LangTag

const (
	Lcpp LangTag = iota
	Lcpp11
	Lcpp14
	Lcpp17
	Lcpp20
	Lpython2
	Lpython3
	Lgo
	Ljava
	Lc
	Lplain
	Lpython
)

// 根据字符串推断程序语言
func SourceLang(s string) LangTag {
	if strings.Contains(s, "java") {
		return Ljava
	}
	if strings.Contains(s, "cpp") || strings.Contains(s, "cc") {
		if strings.Contains(s, fmt.Sprint(11)) {
			return Lcpp11
		}
		if strings.Contains(s, fmt.Sprint(14)) {
			return Lcpp14
		}
		if strings.Contains(s, fmt.Sprint(17)) {
			return Lcpp17
		}
		if strings.Contains(s, fmt.Sprint(20)) {
			return Lcpp20
		}
		return Lcpp
	}
	if strings.Contains(s, "py") {
		if strings.Contains(s, fmt.Sprint(2)) {
			return Lpython2
		}
		if strings.Contains(s, fmt.Sprint(3)) {
			return Lpython3
		}
		return Lpython
	}
	if strings.Contains(s, "go") {
		return Lgo
	}
	if strings.Contains(s, "c") {
		return Lc
	}
	return Lplain
}

type CtntType = yutils.CtntType

const (
	Cplain CtntType = iota
	Cbinary
	Csource
)

func init() {
	rand.Seed(time.Now().Unix())
}

// dependon: whether i dependon j.
// Complexity: O(n^2)
func TopoSort(size int, dependon func(i, j int) bool) (res []int, err error) {
	indegree := make([]int, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i != j && dependon(i, j) {
				indegree[i]++
			}
		}
	}
	res = make([]int, 0, size)
	err = nil
	for {
		pre := len(res)
		for i := 0; i < size; i++ {
			if indegree[i] == 0 {
				res = append(res, i)
				indegree[i] = -1
			}
		}
		if pre == len(res) {
			break
		}
		for id := pre; id < len(res); id++ {
			i := res[id]
			for j := 0; j < size; j++ {
				if i != j && dependon(j, i) {
					if indegree[j] < 0 {
						panic("topo sort error")
					}
					indegree[j]--
				}
			}
		}
	}
	if len(res) != size {
		err = fmt.Errorf("not a DAG")
	}
	return
}

// index of the first element equaling to v, otherwise return -1
func FindIndex[T comparable](array []T, v T) int {
	for i, item := range array {
		if item == v {
			return i
		}
	}
	return -1
}
