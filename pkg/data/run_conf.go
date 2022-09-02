package data

import "encoding/json"

// runner config
type RunConf struct {
	RealTime, CpuTime, VirMem, RealMem, StkMem, Output, Fileno uint // limitation
	// 如果 Inf 和 Ouf 都非空那么识别为文件 IO
	Inf, Ouf    string // input file name, output file name (not data)
	Interpreter string
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
