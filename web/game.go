package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/meiro/field"
	"math"
	"math/rand"
	"time"
)

type GameState int

const (
	GameStateInit GameState = iota
	GameStateMap
)

const (
	keyLeft  = 37
	keyUp    = 38
	keyRight = 39
	keyDown  = 40
	keyA     = 65
	keyS     = 83
	keyD     = 68
)

type Game struct {
	state           GameState
	nextState       GameState
	field           *field.Field
	fieldWorker     js.Object
	currentPosition []int
	pressedKey      int
	shiftPressed    bool
}

func (g *Game) onKeydown(event js.Object) {
	switch g.state {
	case GameStateMap:
		keyCode := event.Get("keyCode").Int()
		g.pressedKey = keyCode
		g.shiftPressed = event.Get("shiftKey").Bool()
	}
}

func (g *Game) Update() {
	defer func() {
		g.pressedKey = 0
		g.shiftPressed = false
	}()
	if g.state != g.nextState {
		g.state = g.nextState
	}
	switch g.state {
	case GameStateInit:
		if g.field == nil {
			random := rand.New(rand.NewSource(time.Now().UnixNano()))
			g.field = field.Create(random, 10, 10, 10, 2)
			return
		}
		g.currentPosition = g.field.StartPosition()
		js.Global.Get("window").Set("onkeydown", g.onKeydown)
		g.nextState = GameStateMap
	case GameStateMap:
		openWall_0_0, openWall_0_1 := g.field.IsWallOpen(g.currentPosition, 0)
		openWall_1_0, openWall_1_1 := g.field.IsWallOpen(g.currentPosition, 1)
		openWall_2_0, openWall_2_1 := g.field.IsWallOpen(g.currentPosition, 2)
		openWall_3_0, openWall_3_1 := g.field.IsWallOpen(g.currentPosition, 3)

		switch g.pressedKey {
		case keyLeft:
			if openWall_0_0 {
				g.currentPosition[0]--
				break
			}
		case keyRight:
			if openWall_0_1 {
				g.currentPosition[0]++
				break
			}
		case keyUp:
			if openWall_1_0 {
				g.currentPosition[1]--
				break
			}
		case keyDown:
			if openWall_1_1 {
				g.currentPosition[1]++
				break
			}
		case keyA:
			if openWall_2_0 && g.currentPosition[2]%2 == 1 {
				g.currentPosition[2]--
				break
			}
			if openWall_2_1 && g.currentPosition[2]%2 == 0 {
				g.currentPosition[2]++
				break
			}
		case keyD:
			if openWall_2_0 && g.currentPosition[2]%2 == 0 {
				g.currentPosition[2]--
				break
			}
			if openWall_2_1 && g.currentPosition[2]%2 == 1 {
				g.currentPosition[2]++
				break
			}
		case keyS:
			if openWall_3_0 || openWall_3_1 {
				g.currentPosition[3] = 1 - g.currentPosition[3]
			}
		}
	default:
		panic("Game.Update: invalid state")
	}
}

func (g *Game) hasDoor(dim int, dir int) bool {
	position := g.currentPosition
	openWall_0, openWall_1 := g.field.IsWallOpen(position, dim)
	if openWall_0 && dir == 0 {
		return true
	}
	if openWall_1 && dir == 1 {
		return true
	}
	nextPosition := make([]int, 4)
	copy(nextPosition, position)
	nextPosition[3] = 1 - nextPosition[3]
	nextOpenWall_0, nextOpenWall_1 := g.field.IsWallOpen(nextPosition, dim)
	if nextOpenWall_0 && dir == 0 {
		return true
	}
	if nextOpenWall_1 && dir == 1 {
		return true
	}
	return false
}

func (g *Game) blockColor(dim int, dir int) int {
	position := g.currentPosition
	openWall_0, openWall_1 := g.field.IsWallOpen(position, dim)
	nextPosition := make([]int, 4)
	copy(nextPosition, position)
	nextPosition[3] = 1 - nextPosition[3]
	nextOpenWall_0, nextOpenWall_1 := g.field.IsWallOpen(nextPosition, dim)
	if dir == 0 {
		if openWall_0 == nextOpenWall_0 {
			return -1
		}
		if openWall_0 {
			return g.currentPosition[3]
		}
		if nextOpenWall_0 {
			return 1 - g.currentPosition[3]
		}
	}
	if dir == 1 {
		if openWall_1 == nextOpenWall_1 {
			return -1
		}
		if openWall_1 {
			return g.currentPosition[3]
		}
		if nextOpenWall_1 {
			return 1 - g.currentPosition[3]
		}
	}
	panic("Game.hasBlock: invalid dir")
}

const grid = 16

func drawBlock(canvas js.Object, x, y int, color string, stroke bool) {
	const blockWidth = grid - 2
	const blockHeight = grid - 2

	if stroke {
		canvas.Set("strokeStyle", color)
		canvas.Call("strokeRect", x+1, y+1, blockWidth, blockHeight)
		return
	}
	canvas.Set("fillStyle", color)
	canvas.Call("fillRect", x+1, y+1, blockWidth, blockHeight)
}

func (g *Game) Draw(canvas js.Object) {
	const roomX = 2 * grid
	const roomY = 2 * grid
	const roomWidth = 7 * grid
	const roomHeight = 5 * grid
	const switchColor0 = "#99f"
	const switchColor1 = "#f99"

	// TODO: Use width / height
	canvas.Call("clearRect", 0, 0, 320, 240)
	switch g.state {
	case GameStateInit:
	case GameStateMap:
		// Walls
		canvas.Call("beginPath")
		canvas.Call("moveTo", roomX, roomY)
		if g.hasDoor(0, 0) {
			canvas.Call("lineTo", roomX, roomY+2*grid)
			canvas.Call("moveTo", roomX, roomY+3*grid)
		}
		canvas.Call("lineTo", roomX, roomY+roomHeight)
		if g.hasDoor(1, 1) {
			canvas.Call("lineTo", roomX+3*grid, roomY+roomHeight)
			canvas.Call("moveTo", roomX+4*grid, roomY+roomHeight)
		}
		canvas.Call("lineTo", roomX+roomWidth, roomY+roomHeight)
		if g.hasDoor(0, 1) {
			canvas.Call("lineTo", roomX+roomWidth, roomY+3*grid)
			canvas.Call("moveTo", roomX+roomWidth, roomY+2*grid)
		}
		canvas.Call("lineTo", roomX+roomWidth, roomY)
		if g.hasDoor(2, 0) && g.currentPosition[2]%2 == 0 ||
			g.hasDoor(2, 1) && g.currentPosition[2]%2 == 1 {
			canvas.Call("lineTo", roomX+6*grid, roomY)
			canvas.Call("moveTo", roomX+5*grid, roomY)
		}
		if g.hasDoor(1, 0) {
			canvas.Call("lineTo", roomX+4*grid, roomY)
			canvas.Call("moveTo", roomX+3*grid, roomY)
		}
		if g.hasDoor(2, 0) && g.currentPosition[2]%2 == 1 ||
			g.hasDoor(2, 1) && g.currentPosition[2]%2 == 0 {
			canvas.Call("lineTo", roomX+2*grid, roomY)
			canvas.Call("moveTo", roomX+1*grid, roomY)
		}
		canvas.Call("lineTo", roomX, roomY)
		canvas.Set("strokeStyle", "#000")
		canvas.Call("stroke")

		// Stairs
		if g.hasDoor(2, 0) {
			canvas.Set("font", "12px Helvetica")
			canvas.Set("textAlign", "center")
			canvas.Set("textBaseline", "top")
			canvas.Set("fillStyle", "#000")
			x := roomX + grid + grid/2
			if g.currentPosition[2]%2 == 0 {
				x = roomX + roomWidth - grid - grid/2
			}
			canvas.Call("fillText", "↑", x, roomY-grid, grid)
		}
		if g.hasDoor(2, 1) {
			canvas.Set("font", "12px Helvetica")
			canvas.Set("textAlign", "center")
			canvas.Set("textBaseline", "top")
			canvas.Set("fillStyle", "#000")
			x := roomX + grid + grid/2
			if g.currentPosition[2]%2 == 1 {
				x = roomX + roomWidth - grid - grid/2
			}
			canvas.Call("fillText", "↓", x, roomY-grid, grid)
		}

		// Switch
		openWall_3_0, openWall_3_1 := g.field.IsWallOpen(g.currentPosition, 3)
		if openWall_3_0 || openWall_3_1 {
			canvas.Call("beginPath")
			cx := roomX + roomWidth/2
			cy := roomY + roomHeight/2
			canvas.Call("arc", cx, cy, grid/2, 2*math.Pi, false)
			color := switchColor0
			if g.currentPosition[3] == 1 {
				color = switchColor1
			}
			canvas.Set("fillStyle", color)
			canvas.Call("fill")
		}

		// Blocks
		if blockColor := g.blockColor(0, 0); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX
			y := roomY + 2*grid
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}
		if blockColor := g.blockColor(0, 1); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX + roomWidth - grid
			y := roomY + 2*grid
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}
		if blockColor := g.blockColor(1, 0); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX + 3*grid
			y := roomY
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}
		if blockColor := g.blockColor(1, 1); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX + 3*grid
			y := roomY + roomHeight - grid
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}
		if blockColor := g.blockColor(2, 0); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX + 1*grid
			if g.currentPosition[2]%2 == 0 {
				x = roomX + roomWidth - 2*grid
			}
			y := roomY
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}
		if blockColor := g.blockColor(2, 1); blockColor != -1 {
			color := switchColor0
			if blockColor == 1 {
				color = switchColor1
			}
			x := roomX + 1*grid
			if g.currentPosition[2]%2 == 1 {
				x = roomX + roomWidth - 2*grid
			}
			y := roomY
			drawBlock(canvas, x, y, color, g.currentPosition[3] == blockColor)
		}

		// Info
		canvas.Set("font", "14px Helvetica")
		canvas.Set("textAlign", "left")
		canvas.Set("textBaseline", "top")
		canvas.Set("fillStyle", "#000")
		floorStr := fmt.Sprintf("B%dF", g.currentPosition[2]+1)
		canvas.Call("fillText", floorStr, 0, 0)
	default:
		panic("Game.Draw: invalid state")
	}
}
