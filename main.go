package main

import (
	"github.com/hajimehoshi/meiro/field"
	"os"
)

func main() {
	f := field.Create(40, 20)
	f.Write(os.Stdout)
}
