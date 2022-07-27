package processors

import (
	"encoding/json"
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/private/judger"
)

// runner config
type RunConf struct {
	RealTime, CpuTime, VirMem, RealMem, StkMem, Output, Fileno uint   // limitation
	Inf, Ouf                                                   string // file io
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

// `s` contains a series of number seperated by space, denoting
// real time (ms), cpu time (ms), virtual memory (byte), real memory (byte),
// stack memory (byte), output limit (byte), fileno limitation respectively.
func runLimOptions(s RunConf) []judger.OptionProvider {
	var rt, ct, vm, rm, sm, ol, fl uint = s.RealTime, s.CpuTime, s.VirMem, s.RealMem, s.StkMem, s.Output, s.Fileno
	options := []judger.OptionProvider{}
	if rt > 0 {
		options = append(options, judger.WithRealTime(time.Millisecond*time.Duration(rt)))
	}
	if ct > 0 {
		options = append(options, judger.WithCpuTime(time.Millisecond*time.Duration(ct)))
	}
	if vm > 0 {
		options = append(options, judger.WithVirMemory(judger.ByteValue(vm)))
	}
	if rm > 0 {
		options = append(options, judger.WithRealMemory(judger.ByteValue(rm)))
	}
	if sm > 0 {
		options = append(options, judger.WithStack(judger.ByteValue(sm)))
	}
	if ol > 0 {
		options = append(options, judger.WithOutput(judger.ByteValue(ol)))
	}
	if fl > 0 {
		options = append(options, judger.WithFileno(int(fl)))
	}
	return options
}
