package data

import (
	"io"
	"os"
)

type flexStoreMode int

const (
	mCtnt flexStoreMode = 0
	mFile flexStoreMode = 1
)

// 根据调用的方法灵活变化存储方式
//
// 可直接声明无需初始化（最好不要）
//
// 在文件模式中我们存储的是文件路径而非 os.File 指针，因此人为地改动文件内容是可以的
//
// TODO: 规范的错误处理、test
type Flex struct {
	// default mCtnt
	mode flexStoreMode
	// if mFile
	filepath string
	// if mCtnt
	content []byte
}

// turn to file mode (if not)
func (r *Flex) ToFile() error {
	if r.mode == mFile {
		return nil
	}
	err := os.WriteFile(r.filepath, r.content, os.ModePerm)
	if err != nil {
		return err
	}
	r.mode = mFile
	r.content = nil
	return nil
}

func (r *Flex) Path() string {
	if err := r.ToFile(); err != nil {
		panic(err)
	}
	return r.filepath
}

func (r *Flex) File() (*os.File, error) {
	if err := r.ToFile(); err != nil {
		panic(err)
	}
	return os.Open(r.filepath)
}

func (r *Flex) SetMode(mode os.FileMode) error {
	if err := r.ToFile(); err != nil {
		panic(err)
	}
	return os.Chmod(r.filepath, mode)
}

func (r *Flex) Get() (data []byte, err error) {
	if r.mode == mFile {
		return os.ReadFile(r.filepath)
	} else {
		return r.content, nil
	}
}

func (r *Flex) Set(data []byte) error {
	if r.mode == mFile {
		return os.WriteFile(r.filepath, data, os.ModePerm)
	} else {
		r.content = data
		return nil
	}
}

// change the filepath of store
//
// this does not remove the origin file
func (r *Flex) ChangePath(name string) error {
	if r.mode == mCtnt {
		r.filepath = name
	} else if r.filepath != name {
		src, err := r.File()
		if err != nil {
			return err
		}
		defer src.Close()

		dest, err := os.Create(name)
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
		stat, err := src.Stat()
		if err != nil {
			return err
		}
		err = os.Chmod(name, stat.Mode())
		if err != nil {
			return err
		}
		r.filepath = name
	}
	return nil
}

func (r *Flex) DupFile(name string, mode os.FileMode) error {
	data, err := r.Get()
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, mode)
}

var _ FileStore = (*Flex)(nil)

// flex store with filepath initialized (empty content)
func FlexWithPath(name string) *Flex {
	return &Flex{filepath: name}
}

// flex store with file initialized
func FlexWithFile(name string) *Flex {
	res := &Flex{
		mode:     mFile,
		filepath: name,
	}
	return res
}

func FlexWithData(data []byte) *Flex {
	res := &Flex{}
	res.Set(data)
	return res
}

func NewFlex(name string, data []byte) *Flex {
	return &Flex{
		mode:     mCtnt,
		filepath: name,
		content:  data,
	}
}

func FlexFromStore(store Store) (*Flex, error) {
	res := &Flex{}
	data, err := store.Get()
	if err != nil {
		return nil, err
	}
	err = res.Set(data)
	if err != nil {
		return nil, err
	}
	return res, nil
}
