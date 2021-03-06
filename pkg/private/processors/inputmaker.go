package processors

import (
	"os"
	"strings"

	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Inputmaker make input according to "option": "raw" means "source" provides
// input content, "generator" means execute "generator" with arguments in
// "source", separated by space.
type Inputmaker struct {
	// source option generator
	// output: result stderr judgerlog
}

func (r Inputmaker) Label() (inputlabel []string, outputlabel []string) {
	return []string{"source", "option", "generator"}, []string{"result", "stderr", "judgerlog"}
}

func (r Inputmaker) Run(input []string, output []string) *Result {
	os.Chmod(input[2], 0744)

	option, err := os.ReadFile(input[1])
	if err != nil {
		return RtErrRes(err)
	}
	if strings.Contains(string(option), "raw") {
		if _, err := utils.CopyFile(input[0], output[0]); err != nil {
			return RtErrRes(err)
		}
		return &Result{Code: processor.Ok}
	} else { // testlib
		runner := GeneratorTestlib{}
		return runner.Run([]string{input[2], input[0]}, output)
	}
}

var _ Processor = Inputmaker{}
