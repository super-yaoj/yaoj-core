package problem

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 一个提交记录由入口组下若干个文件构成。
//
// 考虑到对 hack 的支持，一个提交可以不止包含 Gsubm 域的内容
type Submission map[workflow.Groupname]map[string]data.Store

// 根据文件路径名加入提交文件
func (r Submission) Set(field string, filename string) {
	file, _ := os.Open(filename)
	r.SetSource(workflow.Gsubm, field, filename, file)
	file.Close()
}

// 加入文件（例如custom test就可以手动加test）
//
//	group: 所属数据组，一般是 workflow.Gsubm 表示提交数据。
//	field: 字段名
//	name: 文件名（一般不带路径）
//	reader：文件内容
func (r Submission) SetSource(group workflow.Groupname, field string, name string, reader io.Reader) {
	logger.Printf("SetSource in group %s's %q naming %q", group, field, name)
	if r[group] == nil {
		r[group] = make(map[string]data.Store)
	}
	ctnt, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	r[group][field] = data.NewInMemory(ctnt)
}

func (r Submission) DumpTo(writer io.Writer) error {
	w := zip.NewWriter(writer)
	defer w.Close()

	var pathmap = map[workflow.Groupname]map[string]string{}

	for group, data := range r {
		if data == nil {
			continue
		}
		if pathmap[group] == nil {
			pathmap[group] = map[string]string{}
		}
		for field, store := range data {
			filename := string(group) + "-" + field
			fileInzip, err := w.Create(filename)
			if err != nil {
				return err
			}
			ctnt, err := store.Get()
			if err != nil {
				return err
			}
			_, err = fileInzip.Write(ctnt)
			if err != nil {
				return err
			}
			pathmap[group][field] = filename
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

// 打包
func (r Submission) DumpFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return r.DumpTo(file)
}

// to inbound groups
func (r Submission) Download(dir string) (res workflow.InboundGroups) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	prefix := utils.RandomString(10)
	res = workflow.InboundGroups{}
	for group, gdata := range r {
		if gdata == nil {
			continue
		}
		res[group] = make(map[string]data.FileStore)
		for field, store := range gdata {
			filename := path.Join(dir, prefix+"-"+string(group)+"-"+field)
			flex, err := data.FlexFromStore(store)
			if err != nil {
				panic(err)
			}
			flex.ChangePath(filename)
			res[group][field] = flex
		}
	}
	return res
}

// 针对某个域的提交文件配置
//
// json mashalable
type SubmConf map[string]SubmLimit

// Limitation for submitted files
type SubmLimit struct {
	// 接受的语言，nil 表示所有语言
	Langs []utils.LangTag
	// 接受哪些类型的文件，必须设置值
	Accepted utils.CtntType
	// 文件大小，单位 byte
	Length uint32
}

// 事实上只检查长度
// TODO: 完善提交检查
func (r SubmLimit) Validate(data []byte) error {
	if len(data) > int(r.Length) {
		return fmt.Errorf("file size limit exceed")
	}
	return nil
}

func loadSubmOpener(zipfile interface {
	Open(name string) (fs.File, error)
}) (Submission, error) {
	file, _ := zipfile.Open("_config.json")
	confdata, _ := io.ReadAll(file)
	var pathmap map[workflow.Groupname]*map[string]string
	if err := json.Unmarshal(confdata, &pathmap); err != nil {
		return nil, err
	}

	var res = Submission{}
	for group, data := range pathmap {
		for field, name := range *data {
			file, _ := zipfile.Open(name)
			res.SetSource(group, field, name, file)
			file.Close()
		}
	}
	return res, nil
}

// 解压
func LoadSubm(name string) (Submission, error) {
	zipfile, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	defer zipfile.Close()

	return loadSubmOpener(zipfile)
}

func LoadSubmData(data []byte) (Submission, error) {
	reader := bytes.NewReader(data)
	zipfile, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, err
	}

	return loadSubmOpener(zipfile)
}
