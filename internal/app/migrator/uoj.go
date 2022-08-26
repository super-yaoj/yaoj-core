package migrator

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
)

func NewUojTraditional(data_dir string, logger *log.Entry) *UojTraditional {
	return &UojTraditional{
		data_dir: data_dir,
		lg:       logger.WithField("migrator", "uoj_traditional"),
	}
}

type UojTraditional struct {
	// 数据文件夹（就是svn的1文件夹里的内容。也就是说 problem.conf 是配置文件）
	data_dir string
	lg       *log.Entry
}

var _ Migrator = (*UojTraditional)(nil)

func (r *UojTraditional) Migrate(dest string) error {
	tmpdir := path.Join(os.TempDir(), "yaoj-migrator-"+utils.RandomString(10))
	if err := os.MkdirAll(tmpdir, os.ModePerm); err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	// create problem
	prob, err := problem.New(tmpdir, r.lg)
	if err != nil {
		return err
	}

	prob.Fullscore = 100
	prob.AnalyzerName = "traditional"
	prob.Workflow = &preset.Traditional

	// read conf
	fconf, err := os.ReadFile(path.Join(r.data_dir, "problem.conf"))
	if err != nil {
		return err
	}
	conf := r.parseConf(fconf)

	// ensure use builtin judger
	if conf["use_builtin_judger"] != "on" {
		return ErrUnsupportedJudger
	}

	// parse sample
	if conf["n_sample_tests"] != "" {
		prob.Pretest.InitTestcases()
		prob.Pretest.Method = problem.Msum
		prob.Pretest.Fullscore = 100

		nsample := parseInt(conf["n_sample_tests"])
		for i := 1; i <= int(nsample); i++ {
			input := fmt.Sprint("ex_", conf["input_pre"], i, ".", conf["input_suf"])
			output := fmt.Sprint("ex_", conf["output_pre"], i, ".", conf["output_suf"])

			tc := prob.Pretest.NewTestcase()
			err = tc.SetSource("input", path.Join(r.data_dir, input))
			if err != nil {
				return err
			}
			err = tc.SetSource("output", path.Join(r.data_dir, output))
			if err != nil {
				return err
			}
		}
	}
	// parse extra tests
	if conf["n_ex_tests"] != "" {
		prob.Extra.InitTestcases()
		prob.Extra.Method = problem.Msum
		prob.Extra.Fullscore = 100

		nextra := parseInt(conf["n_ex_tests"])
		for i := 1; i <= int(nextra); i++ {
			input := fmt.Sprint("ex_", conf["input_pre"], i, ".", conf["input_suf"])
			output := fmt.Sprint("ex_", conf["output_pre"], i, ".", conf["output_suf"])

			tc := prob.Extra.NewTestcase()
			err = tc.SetSource("input", path.Join(r.data_dir, input))
			if err != nil {
				return err
			}
			err = tc.SetSource("output", path.Join(r.data_dir, output))
			if err != nil {
				return err
			}
		}
	}

	// parse checker
	if _, ok := conf["use_builtin_checker"]; ok {
		r.lg.Infof("use builtin checker: %q", conf["use_builtin_checker"])
		// copy checker
		file, _ := asserts.Open(path.Join("asserts", "checker", conf["use_builtin_checker"]+".cpp"))
		if err != nil {
			return err
		}
		ctnt, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		file.Close()
		prob.Static.SetData("checker", ctnt)
	} else { // custom checker
		err = prob.Static.SetSource("checker", path.Join(r.data_dir, "chk.cpp"))
		if err != nil {
			return err
		}
	}

	// parse limitation
	tl := parseInt(conf["time_limit"])
	ml := parseInt(conf["memory_limit"])
	ol := parseInt(conf["output_limit"])

	err = prob.Static.SetData("runner_config", (&processors.RunConf{
		RealTime: 1000 * 60, // 1min
		CpuTime:  uint(tl) * 1000,
		VirMem:   0,
		RealMem:  uint(ml) * 1024 * 1024,
		StkMem:   uint(ml) * 1024 * 1024,
		Output:   uint(ol) * 1024 * 1024,
		Fileno:   10,
	}).Serialize())
	if err != nil {
		return err
	}
	prob.Attr["time_limit"] = fmt.Sprint(tl * 1000)
	prob.Attr["memory_limit"] = conf["memory_limit"]
	prob.Attr["output_limit"] = conf["output_limit"]

	// parse data
	if sNsubt, ok := conf["n_subtasks"]; ok {
		prob.Data.InitSubtasks()
		prob.Data.Method = problem.Msum
		nsubt, _ := strconv.ParseInt(sNsubt, 10, 32)

		r.lg.Infof("nsubtask = %d", nsubt)

		las := 0
		for i := 1; i <= int(nsubt); i++ {
			endid, _ := strconv.ParseInt(conf[fmt.Sprint("subtask_end_", i)], 10, 32)
			score, _ := strconv.ParseInt(conf[fmt.Sprint("subtask_score_", i)], 10, 32)
			sub := prob.Data.NewSubtask(float64(score), problem.Mmin)

			// if depstr, ok := conf[fmt.Sprint("subtask_dependence_", i)]; ok {
			// 	if depstr == "many" {
			// 		var deps = []string{}
			// 		for j := 1; conf[fmt.Sprint("subtask_dependence_", i, "_", j)] != ""; j++ {
			// 			deps = append(deps, conf[fmt.Sprint("subtask_dependence_", i, "_", j)])
			// 		}
			// 		deps = utils.Map(deps, func(token string) string {
			// 			return "subtask_" + token
			// 		})
			// 		record["_depend"] = strings.Join(deps, ",")
			// 	} else {
			// 		record["_depend"] = "subtask_" + depstr
			// 	}
			// }

			for j := las + 1; j <= int(endid); j++ {
				input := fmt.Sprint(conf["input_pre"], j, ".", conf["input_suf"])
				output := fmt.Sprint(conf["output_pre"], j, ".", conf["output_suf"])

				tc := sub.NewTestcase()
				tc.SetSource("input", path.Join(r.data_dir, input))
				tc.SetSource("output", path.Join(r.data_dir, output))
			}
			las = int(endid)
		}
	} else {
		prob.Data.InitTestcases()
		n_tests := parseInt(conf["n_tests"])

		r.lg.Infof("ntests = %d", n_tests)

		for i := 1; i <= n_tests; i++ {
			input := fmt.Sprint(conf["input_pre"], i, ".", conf["input_suf"])
			output := fmt.Sprint(conf["output_pre"], i, ".", conf["output_suf"])

			tc := prob.Data.NewTestcase()
			tc.SetSource("input", path.Join(r.data_dir, input))
			tc.SetSource("output", path.Join(r.data_dir, output))
		}
	}

	// analyzer
	// panic("not complete")
	prob.Submission = problem.SubmConf{
		"source": {
			Length:   1024 * 64,
			Accepted: utils.Csource,
		},
		"option": {
			Length:   1024 * 64,
			Accepted: utils.Ccompconf,
		},
	}

	err = prob.DumpFile(dest)
	if err != nil {
		return err
	}
	return nil
}

// 对于每一行，第一个 token 作为字段，之后的作为值
func (r *UojTraditional) parseConf(content []byte) (res map[string]string) {
	res = map[string]string{}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens := strings.Split(line, " ")

		finaltokens := []string{}
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token != "" {
				finaltokens = append(finaltokens, token)
			}
		}
		tokens = finaltokens

		var directive, val string
		if len(tokens) == 1 {
			directive = tokens[0]
		} else {
			directive, val = tokens[0], strings.Join(tokens[1:], " ")
		}
		res[directive] = val
	}
	// r.lg.Infof("conf: %+v", res)
	return
}

func parseInt(s string) int {
	res, _ := strconv.ParseInt(s, 10, 32)
	return int(res)
}
