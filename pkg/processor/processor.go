// Package processor provides builtin processors and processor plugin loader
package processor

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

type Bounds map[string]data.FileStore
type Inbounds Bounds
type Outbouds Bounds

// Processor takes a series of input (files) and generates a series of outputs.
type Processor interface {
	// Report human-readable label for each input and output.
	// Labels are used in workflow (especially builder).
	Label() (inputlabel []string, outputlabel []string)
	// Given a fixed number of input files, generate output to  corresponding files
	// with execution result. It's ok if result == nil, which means success.
	// Run(inputs []string, outputs []string) (result *Result)

	// Given a fixed number of input files, generate output to  corresponding files
	// with execution result. It's ok if result == nil, which means success.
	// Inputs are considered unordered.
	Process(inputs Inbounds, outputs Outbouds) (result *Result)
}

type Code int

const (
	Ok Code = iota
	RuntimeError
	MemoryExceed
	TimeExceed
	OutputExceed
	SystemError
	DangerousSyscall
	ExitError
)

// Result of processor' execution
//
// Code is required, others are optional
type Result struct {
	// Result status：OK/RE/MLE/...
	Code              Code
	RealTime, CpuTime *time.Duration
	Memory            *utils.ByteValue
	// describing the status
	Msg string
}

func (r Result) Serialize() []byte {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	err := encoder.Encode(r)
	if err != nil {
		panic(err)
	}
	return b.Bytes()[:]
}

func (r *Result) Unserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(r)
	return err
}

func init() {
	gob.Register(Result{})
}
