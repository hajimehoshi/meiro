package main

import (
	"github.com/gopherjs/gopherjs/js"
	"time"
)

func main() {
	game := new(Game)
	canvas := js.Global.Get("document").Call("getElementById", "mainCanvas").Call("getContext", "2d")
	for {
		game.Update()
		game.Draw(canvas)
		time.Sleep(0)
	}
}
