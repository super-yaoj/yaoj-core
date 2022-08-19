// 内置处理器
package processors

import (
	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

type (
	Processor = processor.Processor
	Result    = processor.Result
	Inbounds  = processor.Inbounds
	Outbounds = processor.Outbounds
)

var logger = buflog.New("[processors] ")
