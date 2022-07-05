package problem

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 存储文件的路径
type Submission map[workflow.Groupname]*map[string]string

// 加入提交文件
func (r Submission) Set(field string, name string) {
	r.SetFile(workflow.Gsubm, field, name)
}

// 加入文件（例如custom test就可以手动加test）
func (r Submission) SetFile(group workflow.Groupname, field string, name string) {
	if r[group] == nil {
		r[group] = &map[string]string{}
	}
	(*r[group])[field] = name
}

// 打包
func (r Submission) DumpFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	var pathmap = map[workflow.Groupname]*map[string]string{}

	for group, data := range r {
		if data == nil {
			continue
		}
		if pathmap[group] == nil {
			pathmap[group] = &map[string]string{}
		}
		for field, name := range *data {
			file, err := os.Open(name)
			if err != nil {
				return err
			}

			filename := string(group) + "-" + field + "-" + path.Base(name)
			f, err := w.Create(filename)
			if err != nil {
				return err
			}

			_, err = io.Copy(f, file)
			if err != nil {
				return err
			}

			file.Close()
			(*pathmap[group])[field] = filename
		}
	}

	conf, err := w.Create("_config.json")
	if err != nil {
		return err
	}

	jsondata, err := json.Marshal(pathmap)
	if err != nil {
		return err
	}

	conf.Write(jsondata)
	return nil
}

// 解压
func LoadSubm(name string, dir string) (Submission, error) {
	err := unzipSource(name, dir)
	if err != nil {
		return nil, err
	}
	bconf, err := os.ReadFile(path.Join(dir, "_config.json"))
	if err != nil {
		return nil, err
	}
	var pathmap map[workflow.Groupname]*map[string]string
	if err := json.Unmarshal(bconf, &pathmap); err != nil {
		return nil, err
	}
	var res = Submission{}
	for group, data := range pathmap {
		res[group] = &map[string]string{}
		for field, name := range *data {
			(*res[group])[field] = path.Join(dir, name)
		}
	}
	return res, nil
}

// 提交文件配置
type SubmConf map[string]SubmLimit

// limitation for any file submitted
type SubmLimit struct {
	// 接受的语言，nil 表示所有语言
	Langs []utils.LangTag
	// 接受哪些类型的文件，必须设置值
	Accepted utils.CtntType
	// 文件大小，单位 byte
	Length uint32
}
