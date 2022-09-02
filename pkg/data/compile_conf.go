package data

import (
	"encoding/json"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// 编译配置
type CompileConf struct {
	// 编译语言
	Lang utils.LangTag
	// 额外的命令行参数（对于 python 来说没用）
	ExtraArgs []string
}

func (r *CompileConf) Serialize() (res []byte) {
	res, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return
}

func (r *CompileConf) Deserialize(data []byte) error {
	return json.Unmarshal(data, r)
}
