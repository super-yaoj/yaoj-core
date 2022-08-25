package worker

import "os"

// 提供题目评测的服务
//
// 原则上全局只有一个 Service 实例
type Service struct {
	// working directory
	dir string
}

// create a new worker in a dir
//
// create the dir if necessary
func New(dir string) (*Service, error) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}
	return &Service{dir: dir}, nil
}
