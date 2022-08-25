package data

import (
	"io"
	"os"
)

// 存储的是文件路径而非 os.File 指针，因此人为地改动文件内容是可以的
//
// TODO: 规范的错误处理、test
type File struct {
	// if mFile
	filepath string
}

func (r *File) Path() string {
	return r.filepath
}

func (r *File) File() (*os.File, error) {
	return os.Open(r.filepath)
}

func (r *File) SetMode(mode os.FileMode) error {
	return os.Chmod(r.filepath, mode)
}

func (r *File) Get() (data []byte, err error) {
	return os.ReadFile(r.filepath)
}

func (r *File) Set(data []byte) error {
	return os.WriteFile(r.filepath, data, os.ModePerm)
}

// change the filepath of store
//
// this does not remove the origin file
func (r *File) ChangePath(name string) error {
	if r.filepath != name {
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

func (r *File) DupFile(name string, mode os.FileMode) error {
	data, err := r.Get()
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, mode)
}

var _ FileStore = (*File)(nil)

// 必须指定一个有效的文件路径，创建一个 File
func NewFile(name string, data []byte) *File {
	res := &File{
		filepath: name,
	}
	err := res.Set(data)
	if err != nil {
		panic(err)
	}
	return res
}

// 必须指定一个有效的文件路径，创建一个 File
func NewFileStore(name string, store Store) (*File, error) {
	data, err := store.Get()
	if err != nil {
		return nil, err
	}
	return NewFile(name, data), nil
}

// File store with file initialized
func NewFileFile(name string) *File {
	res := &File{
		filepath: name,
	}
	return res
}

/*
func FileFromStore(store Store) (*File, error) {
	res := &File{}
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
*/
