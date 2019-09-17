package enumcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/loov/enumcheck"
)

func TestFromFileSystem(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, enumcheck.Analyzer,
		"enumbyte", "enumstring", "enumstruct", "indirect")
}
