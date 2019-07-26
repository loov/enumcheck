package checkenum_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/loov/checkenum"
)

func TestFromFileSystem(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, checkenum.Analyzer, "def", "use", "defs")
}
