package problem

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"golang.org/x/text/language"
)

var SupportLangs = []language.Tag{
	language.Chinese,
	language.English,
	language.Und,
}

var langMatcher = language.NewMatcher(SupportLangs)

// 猜测 locale 与支持的语言中匹配的语言。如果是 Und 那么返回第一个语言（默认）
func GuessLang(lang string) string {
	tag, _, _ := langMatcher.Match(language.Make(lang))
	if tag == language.Und {
		tag = SupportLangs[0]
	}
	base, _ := tag.Base()
	return base.String()
}

// 测试点得分的汇总方式
//
// 对于不开子任务的题目同样有效
//
// 通常的子任务模式：外 Msum 内 Mmin
//
// 传统模式：Msum 不开子任务
type CalcMethod int

const (
	// default
	Mmin CalcMethod = iota
	Mmax
	Msum
)

// Problem result
type Result struct {
	IsSubtask  bool
	CalcMethod CalcMethod
	Subtask    []SubtResult
}

func (r Result) Byte() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

var briefTpl = template.Must(template.New("brief").Parse(`
subtask: {{ .IsSubtask }}
{{if .IsSubtask}}{{range .Subtask}}{{ .Subtaskid }} ({{ .Fullscore }}pts)
{{range .Testcase}}{{ .Title }} {{ .Score }}pts {{ .Time }} {{ .Memory }}
{{end}}{{end}}
{{else}}{{range .Subtask}}{{range .Testcase}}{{ .Title }} {{ .Score }}pts {{ .Time }} {{ .Memory }}
{{end}}{{end}}
{{end}}
`))

func (r Result) Brief() string {
	var b bytes.Buffer
	if err := briefTpl.Execute(&b, r); err != nil {
		panic(err)
	}
	return b.String()
}

// Subtask result
type SubtResult struct {
	Subtaskid string
	Fullscore float64
	Score     float64
	Testcase  []workflow.Result
}

func (r SubtResult) IsFull() bool {
	return r.Fullscore-r.Score < 1e-5
}

var logger = buflog.New("[problem] ")
