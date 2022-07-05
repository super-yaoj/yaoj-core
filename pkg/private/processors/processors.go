package processors

import (
	"github.com/sshwy/yaoj-core/pkg/buflog"
	"github.com/sshwy/yaoj-core/pkg/processor"
)

type Processor = processor.Processor

type Result = processor.Result

var logger = buflog.New("[processors] ")
