package problem

import (
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type Problem interface {
	// 将题目打包为一个文件（压缩包）
	DumpFile(filename string) error
	// 获取题面，lang 见 http://www.lingoes.net/zh/translator/langcode.htm
	Stmt(lang string) []byte
	// 题解
	Tutr(lang string) []byte
	// 附加文件
	Assert(filename string) (*os.File, error)
	// 获取提交格式的数据表格
	SubmConf() SubmConf
	// 评测用的
	Data() *ProbData
	// 展示数据
	DataInfo() DataInfo
}

type DataInfo struct {
	IsSubtask  bool
	Fullscore  float64
	CalcMethod CalcMethod //计分方式
	Subtasks   []SubtaskInfo
	// 静态文件
	Static map[string]string //other properties of data
}

type SubtaskInfo struct {
	Id        int
	Fullscore float64
	Depend    []int
	Field     map[string]string //other properties of subtasks
	Tests     []TestInfo
}

type TestInfo struct {
	Id    int
	Field map[string]string //other properties of tests, i.e. in/output file path
}

type prob struct {
	data *ProbData
}

// 将题目打包为一个文件（压缩包）
func (r *prob) DumpFile(filename string) error {
	return zipDir(r.data.dir, filename)
}

func (r *prob) tryReadFile(filename string) []byte {
	ctnt, _ := os.ReadFile(path.Join(r.data.dir, filename))
	return ctnt
}

func (r *prob) Stmt(lang string) []byte {
	lang = GuessLang(lang)
	logger.Printf("Get statement lang=%s", lang)
	filename := r.data.Statement["s."+lang]
	return r.tryReadFile(filename)
}

func (r *prob) Tutr(lang string) []byte {
	lang = GuessLang(lang)
	logger.Printf("Get tutorial lang=%s", lang)
	filename := r.data.Statement["t."+lang]
	return r.tryReadFile(filename)
}

func (r *prob) Assert(filename string) (*os.File, error) {
	return os.Open(path.Join(r.data.dir, r.data.Statement[filename]))
}

// 获取提交格式的数据表格
func (r *prob) SubmConf() SubmConf {
	return r.data.Submission
}

func (r *prob) DataInfo() DataInfo {
	var res = DataInfo{
		IsSubtask:  r.data.IsSubtask(),
		Fullscore:  r.data.Fullscore,
		CalcMethod: r.data.CalcMethod,
		Static:     r.data.Static,
		Subtasks:   []SubtaskInfo{},
	}
	if res.IsSubtask {
		for i, task := range r.data.Subtasks.Record {
			var tests = []TestInfo{}
			for j, test := range r.data.Tests.Record {
				if test["_subtaskid"] != task["_subtaskid"] {
					continue
				}
				tests = append(tests, TestInfo{
					Id:    j,
					Field: copyRecord(test),
				})
			}

			depend := []int{}
			if task["_depend"] != "" {
				deps := strings.Split(task["_depend"], ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					for id, subt := range r.data.Subtasks.Record {
						if subt["_subtaskid"] == dep {
							depend = append(depend, id)
						}
					}
				}
			}
			score, _ := strconv.ParseFloat(task["_score"], 64)
			res.Subtasks = append(res.Subtasks, SubtaskInfo{
				Id:        i,
				Fullscore: score,
				Field:     task,
				Tests:     tests,
				Depend:    depend,
			})
		}
	} else {
		var tests = []TestInfo{}
		for j, test := range r.data.Tests.Record {
			tests = append(tests, TestInfo{
				Id:    j,
				Field: copyRecord(test),
			})
		}

		res.Subtasks = append(res.Subtasks, SubtaskInfo{
			Fullscore: r.data.Fullscore,
			Tests:     tests,
		})
	}
	return res
}

var _ Problem = (*prob)(nil)

// 加载一个题目文件夹
func LoadDir(dir string) (Problem, error) {
	data, err := LoadProbData(dir)
	if err != nil {
		return nil, err
	}
	return &prob{data: data}, nil
}

// 将打包的题目在空的文件夹下加载
func LoadDump(filename string, dir string) (Problem, error) {
	err := unzipSource(filename, dir)
	if err != nil {
		return nil, err
	}
	return LoadDir(dir)
}

var SupportLangs = []language.Tag{
	language.Chinese,
	language.English,
	language.Und,
}

var langMatcher = language.NewMatcher(SupportLangs)

var logger = log.New(os.Stderr, "[problem] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

func GuessLang(lang string) string {
	tag, _, _ := langMatcher.Match(language.Make(lang))
	if tag == language.Und {
		tag = SupportLangs[0]
	}
	base, _ := tag.Base()
	return base.String()
}

func (r *prob) Data() *ProbData {
	return r.data
}
