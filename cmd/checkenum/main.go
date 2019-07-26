package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/loov/checkenum"
)

func main() { singlechecker.Main(checkenum.Analyzer) }
