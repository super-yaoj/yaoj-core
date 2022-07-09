package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

type HashValue []byte

func HashSum(a []HashValue) (sum HashValue) {
	h := sha256.New()
	for _, v := range a {
		h.Write(v)
	}
	h.Sum(sum)
	return
}

type ByteValue int64

func (r ByteValue) String() string {
	num := float64(r)
	if num < 1000 {
		return fmt.Sprint(int64(num), "B")
	} else if num < 1e6 {
		return fmt.Sprintf("%.1f%s", num/1e3, "KB")
	} else if num < 1e9 {
		return fmt.Sprintf("%.1f%s", num/1e6, "MB")
	} else {
		return fmt.Sprintf("%.1f%s", num/1e9, "GB")
	}
}

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
type Checksum [32]byte

func (r Checksum) String() string {
	s := ""
	for _, v := range r {
		s += fmt.Sprintf("%02x", v)
	}
	return s
}

type LangTag int

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
)

// 根据字符串推断程序语言
func SourceLang(s string) LangTag {
	if strings.Contains(s, "java") {
		return Ljava
	}
	if strings.Contains(s, "cpp") {
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
		return Lpython3
	}
	if strings.Contains(s, "go") {
		return Lgo
	}
	if strings.Contains(s, "c") {
		return Lc
	}
	return Lplain
}

type CtntType int

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

type RatingRater interface {
	Rate(rating int)
	Rating() int
	// 此前参加了多少场比赛（不算当前这场）
	Count() int
}

// 注意 list 应按照比赛成绩从高到底排序
// implementing https://codeforces.com/blog/entry/20762
// https://codeforces.com/blog/entry/77890
func CalcRating[T RatingRater](list []T) error {
	var beginnerStage = []int{500, 350, 250, 150, 100, 50}

	// probability of A winning B
	Pwin := func(ratingA int, ratingB int) float64 {
		return 1 / float64(1+math.Pow(10, float64(ratingB-ratingA)/400))
	}
	// expect place before contest
	SeedOf := func(rating int) (res float64) {
		res = 1
		for _, ctst := range list {
			res += Pwin(ctst.Rating(), rating)
		}
		return res
	}

	ExpectRatingOf := func(rating int, rank int) int {
		midRank := math.Sqrt(SeedOf(rating) * float64(rank))
		var l, r = 1, 8000
		for l < r {
			mid := (l + r) / 2
			if SeedOf(mid) < midRank {
				r = mid
			} else {
				l = mid + 1
			}
		}
		return l
	}

	// Total sum should not be more than zero.
	var deltas = make([]int, len(list))
	var sumDelta int
	for i, ctst := range list {
		deltas[i] = (ExpectRatingOf(ctst.Rating(), i+1) - ctst.Rating()) / 2
		sumDelta += deltas[i]
	}
	inc := -sumDelta/len(list) - 1
	for i := range deltas {
		deltas[i] += inc
	}

	// Sum of top-4*sqrt should be adjusted to zero.
	sortedList := make([]int, 0, len(list))
	for i := range list {
		sortedList = append(sortedList, i)
	}
	sort.Slice(sortedList, func(i, j int) bool {
		return list[sortedList[i]].Rating() > list[sortedList[j]].Rating()
	})

	s := int(math.Floor(math.Min(float64(len(list)), 4*math.Sqrt(float64(len(list))))))
	var sumd int
	for _, id := range sortedList[:s] {
		sumd += deltas[id]
	}
	inc = int(math.Ceil(math.Min(math.Max(float64(-sumd/s), -10), 0)))
	for i := range deltas {
		deltas[i] += inc
	}

	// validation
	for i := range list {
		for j := i + 1; j < len(list); j++ {
			if list[i].Rating() > list[j].Rating() {
				if !(list[i].Rating()+deltas[i] >= list[j].Rating()+deltas[j]) {
					return fmt.Errorf("first rating invariant failed")
				}
			}
			if list[i].Rating() < list[j].Rating() {
				if !(deltas[i] >= deltas[j]) {
					return fmt.Errorf("second rating invariant failed")
				}
			}
		}
	}

	for i := range list {
		if cnt := list[i].Count(); cnt < len(beginnerStage) {
			list[i].Rate(list[i].Rating() + deltas[i] + beginnerStage[cnt])
		} else {
			list[i].Rate(list[i].Rating() + deltas[i])
		}
	}

	return nil
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
