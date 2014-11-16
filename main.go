package main

import (
	"github.com/hajimehoshi/meiro/field"
	"os"
)

func main() {
	f := field.Create(100, 100)
	f.Write(os.Stdout)
}
