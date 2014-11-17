package main

import (
	"github.com/hajimehoshi/meiro/field"
	"math/rand"
	"os"
	"time"
)

func main() {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	f := field.Create(random, 150, 100)
	f.WriteSVG(os.Stdout)
}
