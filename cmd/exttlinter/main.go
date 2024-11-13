package main

import (
	"github.com/tomato3713/exttlinter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(exttlinter.Analyzer) }
