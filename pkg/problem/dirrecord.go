package problem

import (
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 依赖文件夹的以文件相对路径的形式存储字段值
//
// field 字段对应的文件为 probDir/Dir/field
type DirRecord struct {
	prob *Data
	// 存放数据的关于 ProbDir 的相对路径形式文件夹
	Dir string `json:"dir"`
	// 记录对应字段的内容的语言（用于 yaoj cook）
	Lang map[string]utils.LangTag `json:"lang"`
}

func (r *DirRecord) makeDir() error {
	err := os.MkdirAll(path.Join(r.prob.dir, r.Dir), 0750)
	if err != nil {
		return err
	}
	return nil
}

// 删除某个字段及其数据
func (r *DirRecord) Delete(field string) error {
	delete(r.Lang, field)
	return os.RemoveAll(path.Join(r.prob.dir, r.Dir, field))
}
func (r *DirRecord) SetData(field string, data []byte) error {
	if err := r.makeDir(); err != nil {
		return err
	}
	return os.WriteFile(path.Join(r.prob.dir, r.Dir, field), data, 0644)
}
func (r *DirRecord) GetData(field string) ([]byte, error) {
	data, err := os.ReadFile(path.Join(r.prob.dir, r.Dir, field))
	return data, err
}
func (r *DirRecord) SetSource(field string, source string) error {
	if err := r.makeDir(); err != nil {
		return err
	}
	_, err := utils.CopyFile(source, path.Join(r.prob.dir, r.Dir, field))
	return err
}

// 设置某个字段的内容的语言标签
func (r *DirRecord) SetLang(field string, lang utils.LangTag) {
	r.Lang[field] = lang
}

// 获取某个字段的内容的语言标签, -1 表示没有标签
func (r *DirRecord) GetLang(field string) utils.LangTag {
	if lang, ok := r.Lang[field]; ok {
		return lang
	}
	return -1
}

// 遍历文件夹中的文件（name 是完整的文件名）
func (r *DirRecord) Range(visitor func(field string, name string)) {
	entries, err := os.ReadDir(path.Join(r.prob.dir, r.Dir))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		visitor(name, path.Join(r.prob.dir, r.Dir, name))
	}
}

// 转化为读入数据
func (r *DirRecord) InboundGroup() workflow.InboundGroup {
	res := workflow.InboundGroup{}
	r.Range(func(field, name string) {
		res[field] = data.NewFileFile(name)
	})
	return res
}
