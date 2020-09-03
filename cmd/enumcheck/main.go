package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"loov.dev/enumcheck"
)

func main() { singlechecker.Main(enumcheck.Analyzer) }
