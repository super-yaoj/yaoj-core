package problem

import (
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/buflog"
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

type TestdataInfo struct {
	IsSubtask  bool
	CalcMethod CalcMethod //计分方式
	Subtasks   []SubtaskInfo
}
type DataInfo struct {
	Fullscore float64
	TestdataInfo
	Pretest TestdataInfo
	Extra   TestdataInfo
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
		TestdataInfo: r.data.ProbTestdata.Info(),
		Pretest:      r.data.Pretest.Info(),
		Extra:        r.data.Extra.Info(),
		Fullscore:    r.data.Fullscore,
		Static:       r.data.Static,
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

var logger = buflog.New("[problem] ")
