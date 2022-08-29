// Error handling utilities for Yaoj packages.
package yerrors

import (
	"errors"
	"fmt"
)

// 具有外部情况说明的错误（即错误链条）
//
// 类似 “workflow.Run: open file: file not exist” 这种效果
type situatedError struct {
	// situation can be
	//   - name of a function, e. g. "LoadData:", "workflow.Run:"
	//   - an operation, e. g. "open file:"
	situation string
	// inner error
	err error
}

func (r *situatedError) Error() string {
	return fmt.Sprintf("%s: %v", r.situation, r.err)
}

func (r *situatedError) Unwrap() error {
	return r.err
}

// 具有注解的错误，这里的注解是指附加了一些运行时的数据
//
// 注解会被依次换行打印
type annotatedError struct {
	key   string
	value any
	err   error
}

func (r *annotatedError) Error() string {
	return fmt.Sprintf("%v\n(%s=%v)", r.err, r.key, r.value)
}

func (r *annotatedError) Unwrap() error {
	return r.err
}

// wrapper function of [errors.New]
func New(err string) error {
	return errors.New(err)
}

// 给一个错误包裹一个外部情况说明
//
// 约定：
//   - situation 应当描述出错时调用的函数/做的操作，而不是当前语句所在的函数
//   - err 应当是利用 [New] 里提前定义的错误，类似 ErrInvalidInput，这样可以方面错误处理
func Situated(situation string, err error) error {
	return &situatedError{
		situation: situation,
		err:       err,
	}
}

// 具有注解的错误，这里的注解是指附加了一些运行时的数据
//
// 注解会被依次换行打印
func Annotated(key string, value any, err error) error {
	return &annotatedError{
		key:   key,
		value: value,
		err:   err,
	}
}

// check if err's chain contains target
//
// wrapper function of [errors.Is]
func Is(err, target error) bool {
	return errors.Is(err, target)
}
