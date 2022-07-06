package problem

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 存储文件的路径
type InmemoryFile struct {
	Name string
	Ctnt []byte
}

type Submission map[workflow.Groupname]*map[string]InmemoryFile

// 根据文件路径名加入提交文件
func (r Submission) Set(field string, filename string) {
	file, _ := os.Open(filename)
	r.SetSource(workflow.Gsubm, field, filename, file)
	file.Close()
}

// 加入文件（例如custom test就可以手动加test）
// group: 所属数据组，一般是 workflow.Gsubm 表示提交数据。
// field: 字段名
// name: 文件名（一般不带路径）
// reader：文件内容
func (r Submission) SetSource(group workflow.Groupname, field string, name string, reader io.Reader) {
	logger.Printf("SetSource in group %s's %q naming %q", group, field, name)
	if r[group] == nil {
		r[group] = &map[string]InmemoryFile{}
	}
	var buf bytes.Buffer
	io.Copy(&buf, reader)
	imfile := InmemoryFile{
		Name: path.Base(name),
		Ctnt: buf.Bytes()[:],
	}
	(*r[group])[field] = imfile
}

func (r Submission) DumpTo(writer io.Writer) error {
	w := zip.NewWriter(writer)
	defer w.Close()

	var pathmap = map[workflow.Groupname]*map[string]string{}

	for group, data := range r {
		if data == nil {
			continue
		}
		if pathmap[group] == nil {
			pathmap[group] = &map[string]string{}
		}
		for field, imfile := range *data {
			filename := string(group) + "-" + field + "-" + path.Base(imfile.Name)
			fileInzip, err := w.Create(filename)
			if err != nil {
				return err
			}
			_, err = fileInzip.Write(imfile.Ctnt)
			if err != nil {
				return err
			}
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

// 打包
func (r Submission) DumpFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return r.DumpTo(file)
}

// to path map
func (r Submission) Download(dir string) (res map[workflow.Groupname]*map[string]string) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	prefix := utils.RandomString(10)
	res = map[workflow.Groupname]*map[string]string{}
	for group, data := range r {
		if data == nil {
			continue
		}
		res[group] = &map[string]string{}
		for field, imfile := range *data {
			filename := path.Join(dir, prefix+"-"+string(group)+"-"+field+"-"+imfile.Name)
			err := os.WriteFile(filename, imfile.Ctnt, os.ModePerm)
			if err != nil {
				panic(err)
			}
			(*res[group])[field] = filename
		}
	}
	return res
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

// 解压
func LoadSubm(name string) (Submission, error) {
	// Open the zip file
	zipfile, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	defer zipfile.Close()

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
