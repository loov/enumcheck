package enumcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"loov.dev/enumcheck/enumcheck"
)

func TestFromFileSystem(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, enumcheck.Analyzer,
		"enumbyte",
		"enumstring",
		"enumstruct",
		"enumtype",
		"indirect",
	)
}
