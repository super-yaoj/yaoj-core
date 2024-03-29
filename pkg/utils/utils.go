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
	"golang.org/x/text/language"
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

// const (
// 	Lcpp yutils.LangTag = iota
// 	Lcpp11
// 	Lcpp14
// 	Lcpp17
// 	Lcpp20
// 	Lpython2
// 	Lpython3
// 	Lgo
// 	Ljava
// 	Lc
// 	Lplain
// 	Lpython
// )

// 根据字符串推断程序语言
func SourceLang(s string) yutils.LangTag {
	if strings.Contains(s, "java") {
		return yutils.Ljava
	}
	if strings.Contains(s, "cpp") || strings.Contains(s, "cc") {
		if strings.Contains(s, fmt.Sprint(11)) {
			return yutils.Lcpp11
		}
		if strings.Contains(s, fmt.Sprint(14)) {
			return yutils.Lcpp14
		}
		if strings.Contains(s, fmt.Sprint(17)) {
			return yutils.Lcpp17
		}
		if strings.Contains(s, fmt.Sprint(20)) {
			return yutils.Lcpp20
		}
		return yutils.Lcpp
	}
	if strings.Contains(s, "py") {
		if strings.Contains(s, fmt.Sprint(2)) {
			return yutils.Lpython2
		}
		if strings.Contains(s, fmt.Sprint(3)) {
			return yutils.Lpython3
		}
		return yutils.Lpython
	}
	if strings.Contains(s, "go") {
		return yutils.Lgo
	}
	if strings.Contains(s, "c") {
		return yutils.Lc
	}
	return yutils.Lplain
}

type CtntType = yutils.CtntType

const (
	Cplain CtntType = iota
	Cbinary
	Csource
	Ccompconf // 编译选项，即 CompileConf 序列化后的文件
)

func init() {
	rand.Seed(time.Now().Unix())
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

var SupportLangs = []language.Tag{
	language.Chinese,
	language.English,
	language.Und,
}

var langMatcher = language.NewMatcher(SupportLangs)

// 猜测 locale 与支持的语言中匹配的语言。如果是 Und 那么返回第一个语言（默认）
func GuessLang(lang string) string {
	tag, _, _ := langMatcher.Match(language.Make(lang))
	if tag == language.Und {
		tag = SupportLangs[0]
	}
	base, _ := tag.Base()
	return base.String()
}
