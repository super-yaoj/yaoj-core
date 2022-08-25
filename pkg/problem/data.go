package problem

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 题目数据存储
//
// 对于静态的题目数据，为了方便人类阅读和人为修改，所有的数据都会存储在文件中
type Data struct {
	// 题目满分 Usually 100.
	// Full score can be used to determine the point of testcase
	Fullscore float64 `json:"fullscore"`

	// 该题目所有测试点运行的 workflow
	Workflow *workflow.Workflow `json:"workflow"`

	// pretest 常用于样例评测
	Pretest *TestdataGroup `json:"pretest"`
	// 额外数据例如 hack 数据
	Extra *TestdataGroup `json:"extra"`
	// 题目本身数据
	Data *TestdataGroup `json:"data"`

	// 评测时用到的全局静态文件
	Static *DirRecord `json:"static"`

	// 提交配置 "submission" configuration
	Submission SubmConf `json:"submission_config"`
	// hack 时 Gtests 里需要提交的字段，以及其对应的限制
	// 为 nil 表示不支持 hack
	HackFields SubmConf `json:"hack_config"`
	// 由 tests 的字段映射到 workflow 的中间输出文件
	// 为 nil 表示不支持 hack
	HackIOMap map[string]workflow.Outbound `json:"hack_map"`

	// statement fielded by language
	Statement *DirRecord `json:"statement"`
	// tutorial
	Tutorial *DirRecord `json:"tutorial"`
	// 附加文件
	Attached *DirRecord `json:"attached"`

	// 可以序列化为 json 的元信息
	// "tl" "ml" "ol" denotes cpu time limit (ms), real memory limit (MB)
	// and output limit (MB) respectively.
	Attr map[string]string `json:"attr"`

	// 数据文件存放的文件夹 absolute dir
	dir string
}

// 将整个题目打包
//
// 在题目文件夹下建立的 problem.json 包含所有元信息
func (r *Data) DumpFile(dest string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(r.dir, "problem.json"), data, 0644)
	if err != nil {
		return err
	}
	err = zipDir(r.dir, dest)
	if err != nil {
		return err
	}
	return nil
}

func LoadFileTo(name string, dir string) (*Data, error) {
	err := unzipSource(name, dir)
	if err != nil {
		return nil, err
	}
	conf, err := os.ReadFile(path.Join(dir, "problem.json"))
	if err != nil {
		return nil, err
	}
	res := &Data{
		dir: dir,
	}
	err = json.Unmarshal(conf, res)
	if err != nil {
		return nil, err
	}
	res.initProb()
	return res, nil
}

func (r *Data) initProb() {
	r.Attached.prob = r
	r.Statement.prob = r
	r.Tutorial.prob = r
	r.Static.prob = r
	r.Pretest.initProb(r)
	r.Extra.initProb(r)
	r.Data.initProb(r)
}

// remove problem (dir)
func (r *Data) Finalize() error {
	log.Printf("finalize %q", r.dir)
	return os.RemoveAll(r.dir)
}

func (r *Data) Hackable() bool {
	return r.HackFields != nil && r.HackIOMap != nil
}

// create a new problem in an empty dir. default fullscore: 100
//
// create the dir if necessary
func New(dir string) (*Data, error) {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return nil, err
	}

	res := &Data{
		Fullscore:  100, // default
		Workflow:   workflow.New(),
		Submission: make(SubmConf),
		Attr:       make(map[string]string),
		dir:        dir,
	}
	res.Attached = res.newDirRecord("attached")
	res.Statement = res.newDirRecord("statement")
	res.Tutorial = res.newDirRecord("tutorial")
	res.Static = res.newDirRecord("static")
	res.Pretest = res.newTestdataGroup("pretest")
	res.Extra = res.newTestdataGroup("extra")
	res.Data = res.newTestdataGroup("data")

	return res, nil
}
