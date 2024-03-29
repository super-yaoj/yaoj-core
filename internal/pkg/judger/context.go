package judger

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

//go:generate go version
//go:generate make -C yaoj-judger

//#cgo CFLAGS: -I./yaoj-judger/include
//#cgo LDFLAGS: -L./yaoj-judger -lyjudger -lpthread
//#include "./yaoj-judger/include/judger.h"
//#include <stdlib.h>
import "C"

type LimitType int

const (
	realTime LimitType = C.REAL_TIME
	cpuTime  LimitType = C.CPU_TIME
	// virtual memory
	virtMem  LimitType = C.VIRTUAL_MEMORY
	realMem  LimitType = C.ACTUAL_MEMORY
	stackMem LimitType = C.STACK_MEMORY
	// output size
	outputSize LimitType = C.OUTPUT_SIZE
	filenoLim  LimitType = C.FILENO
)

// Set logging options.
// MUST be executed before creating context.
//
// filename set perform log file.
// log_level determine minimum log level (DEBUG, INFO, WARN, ERROR = 0, 1, 2, 3)
// with_color whether use ASCII color controller character
func logSet(filename string, level int) error {
	var cfilename *C.char = C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	res := C.log_set(cfilename, C.int(level), C.int(0))
	if res != 0 {
		return ErrLogSet
	}
	return nil
}

// Close log file.
func logClose() {
	C.log_close()
}

type context struct {
	ctxt C.yjudger_ctxt_t
}

func newContext() context {
	return context{ctxt: C.yjudger_ctxt_create()}
}

/* func (r context) Result() Result {
	result := C.yjudger_result(r.ctxt)
	signal := int(result.signal)
	exitCode := int(result.exit_code)
	realTime := time.Duration(int(result.real_time) * int(time.Millisecond))
	cpuTime := time.Duration(int(result.cpu_time) * int(time.Millisecond))
	memory := ByteValue(result.real_memory)

	return Result{
		Code:     processor.Code(result.code),
		Signal:   &signal,
		Msg:      fmt.Sprintf("Exit with code %d", exitCode),
		RealTime: &realTime,
		CpuTime:  &cpuTime,
		Memory:   &memory,
	}
}*/

func (r context) Free() {
	C.yjudger_ctxt_free(r.ctxt)
}

func (r context) SetPolicy(dirname string, policy string) error {
	var cdirname, cpolicy *C.char = C.CString(dirname), C.CString(policy)
	defer C.free(unsafe.Pointer(cdirname))
	defer C.free(unsafe.Pointer(cpolicy))

	flag := C.yjudger_set_policy(r.ctxt, cdirname, cpolicy)
	if flag != 0 {
		return ErrSetPolicy
	}
	return nil
}

/*func (r context) SetBuiltinPolicy(policy string) error {
	return r.SetPolicy(".", "builtin:"+policy)
}*/

func cCharArray(a []string) []*C.char {
	var ca []*C.char = make([]*C.char, len(a)+1)
	for i := range a {
		ca[i] = C.CString(a[i])
	}
	ca[len(ca)-1] = nil
	return ca
}

func cFreeCharArray(ca []*C.char) {
	for _, val := range ca {
		if val != nil {
			C.free(unsafe.Pointer(val))
		}
	}
}

func (r context) SetRunner(argv []string, env []string) error {
	cargv, cenv := cCharArray(argv), cCharArray(env)
	defer cFreeCharArray(cargv)
	defer cFreeCharArray(cenv)

	flag := C.yjudger_set_runner(r.ctxt, C.int(len(argv)), &cargv[0], &cenv[0])
	if flag != 0 {
		return ErrSetRunner
	}
	return nil
}

type Runner int

// Runner type
const (
	General     Runner = 0
	Interactive Runner = 1
)

/*func (r context) Run(runner Runner) error {
	var flag C.int
	switch runner {
	case General:
		flag = C.yjudger_general(r.ctxt)
	case Interactive:
		flag = C.yjudger_interactive(r.ctxt)
	default:
		return yerrors.Annotated("runner", runner, ErrUnknownRunner)
	}
	if flag != 0 {
		return ErrRun
	}
	return nil
}*/

func (r context) RunForkGeneral() Result {
	result := C.yjudger_general_fork(r.ctxt)
	signal := int(result.signal)
	exitCode := int(result.exit_code)
	realTime := time.Duration(int(result.real_time) * int(time.Millisecond))
	cpuTime := time.Duration(int(result.cpu_time) * int(time.Millisecond))
	memory := ByteValue(result.real_memory)

	return Result{
		Code:     processor.Code(result.code),
		Signal:   &signal,
		Msg:      fmt.Sprintf("Exit with code %d", exitCode),
		RealTime: &realTime,
		CpuTime:  &cpuTime,
		Memory:   &memory,
	}
}

func (r context) RunForkInteractive() Result {
	result := C.yjudger_interactive_fork(r.ctxt)
	signal := int(result.signal)
	exitCode := int(result.exit_code)
	realTime := time.Duration(int(result.real_time) * int(time.Millisecond))
	cpuTime := time.Duration(int(result.cpu_time) * int(time.Millisecond))
	memory := ByteValue(result.real_memory)

	return Result{
		Code:     processor.Code(result.code),
		Signal:   &signal,
		Msg:      fmt.Sprintf("Exit with code %d", exitCode),
		RealTime: &realTime,
		CpuTime:  &cpuTime,
		Memory:   &memory,
	}
}

// short cut for Limitation
type L map[LimitType]int64

// it does nothing for invalid limit type
func (r context) SetLimit(options L) error {
	for key, val := range options {
		switch key {
		case realTime:
			C.yjudger_set_limit(r.ctxt, C.REAL_TIME, C.int(val))
		case cpuTime:
			C.yjudger_set_limit(r.ctxt, C.CPU_TIME, C.int(val))
		case virtMem:
			C.yjudger_set_limit(r.ctxt, C.VIRTUAL_MEMORY, C.int(val))
		case realMem:
			C.yjudger_set_limit(r.ctxt, C.ACTUAL_MEMORY, C.int(val))
		case stackMem:
			C.yjudger_set_limit(r.ctxt, C.STACK_MEMORY, C.int(val))
		case outputSize:
			C.yjudger_set_limit(r.ctxt, C.OUTPUT_SIZE, C.int(val))
		case filenoLim:
			C.yjudger_set_limit(r.ctxt, C.FILENO, C.int(val))
		}
	}
	return nil
}

func init() {
	if code := C.log_init(); code != 0 {
		panic(fmt.Sprint("init log failed: ", code))
	}
}
