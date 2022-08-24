package preset_test

import (
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
)

func TestAll(t *testing.T) {
	pp.Print(preset.Traditional.Inbound)
}
