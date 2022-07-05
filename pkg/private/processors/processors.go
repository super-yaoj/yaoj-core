package processors

import (
	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

type Processor = processor.Processor

type Result = processor.Result

var logger = buflog.New("[processors] ")
