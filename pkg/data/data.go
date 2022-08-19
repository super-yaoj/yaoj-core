package data

import "os"

// Setter 是指可以修改内容的数据
type Setter interface {
	Set(data []byte) error
}

// Getter 是指可以读取所有内容的数据
type Getter interface {
	// idempotent
	Get() (data []byte, err error)
}

// Filer 是指具有文件形式的数据
type Filer interface {
	// idempotent
	File() (*os.File, error)
	// 获取该文件的路径
	Path() string
}

// ModeSetter 是指可以设置读写模式的数据
//
// RFC: ModeSetter 应当可以直接影响 Setter/Getter 的行为（如果有）
//
// 对于 Filer，它可以影响文件的可执行性
type ModeSetter interface {
	SetMode(mode os.FileMode) error
}

// Store 可以存取数据
type Store interface {
	Setter
	Getter
}

// 可以存取数据，导出为文件的数据
type FileStore interface {
	Store
	Filer
	ModeSetter
}
