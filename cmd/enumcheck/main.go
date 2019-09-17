package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/loov/enumcheck"
)

func main() { singlechecker.Main(enumcheck.Analyzer) }
