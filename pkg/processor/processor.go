// Package processor provides builtin processors and processor plugin loader
package processor

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Processor takes a series of input (files) and generates a series of outputs.
type Processor interface {
	// Report human-readable label for each input and output
	Label() (inputlabel []string, outputlabel []string)
	// Given a fixed number of input files, generate output to  corresponding files
	// with execution result. It's ok if result == nil, which means success.
	Run(input []string, output []string) (result *Result)
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

// Code is required, others are optional
type Result struct {
	// Result status：OK/RE/MLE/...
	Code              Code
	RealTime, CpuTime *time.Duration
	Memory            *utils.ByteValue
	Msg               string
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

// runner config
type RunConf struct {
	RealTime, CpuTime, VirMem, RealMem, StkMem, Output, Fileno uint   // limitation
	Inf, Ouf                                                   string // input file name, output file name (not data)
	Interpreter                                                string
}

func (r *RunConf) Serialize() (res []byte) {
	res, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return
}

func (r *RunConf) Deserialize(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r *RunConf) IsFileIO() bool {
	return r.Inf != "" && r.Ouf != ""
}
