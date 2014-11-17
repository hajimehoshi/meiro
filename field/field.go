package field

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
)

const maxDimension = 4

type Room struct {
	openWalls [maxDimension]bool
}

type Field struct {
	rooms []Room
	sizes [maxDimension]int
}

func (f *Field) Write(writer io.Writer) {
	for j := 0; j < f.sizes[1]; j++ {
		line1 := ""
		line2 := ""
		for i := 0; i < f.sizes[0]; i++ {
			room := f.rooms[roomIndex(f.sizes, [maxDimension]int{i, j})]
			line1 += "+"
			if room.openWalls[1] {
				line1 += "  "
			} else {
				line1 += "--"
			}
			if room.openWalls[0] {
				line2 += " "
			} else {
				line2 += "|"
			}
			line2 += "  "
		}
		line1 += "+\n"
		line2 += "|\n"
		io.WriteString(writer, line1)
		io.WriteString(writer, line2)
	}
	line := strings.Repeat("+--", f.sizes[0]) + "+\n"
	io.WriteString(writer, line)
}

const svgRoomSize = 8

var svgTemplate = `
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns='http://www.w3.org/2000/svg' xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" viewBox="0 0 {{width}} {{height}}" background-color="#fff">
<g transform="translate({{offsetX}}, {{offsetY}})" stroke="black" stroke-width="1" stroke-linecap="round">
{{lines}}
</g>
</svg>
`[1:]

func (f *Field) WriteSVG(writer io.Writer) {
	paddingX := svgRoomSize
	paddingY := svgRoomSize
	svg := svgTemplate
	svg = strings.Replace(svg, "{{width}}", strconv.Itoa(f.sizes[0]*svgRoomSize+paddingX*2), -1)
	svg = strings.Replace(svg, "{{height}}", strconv.Itoa(f.sizes[1]*svgRoomSize+paddingY*2), -1)
	svg = strings.Replace(svg, "{{offsetX}}", strconv.Itoa(paddingX), -1)
	svg = strings.Replace(svg, "{{offsetY}}", strconv.Itoa(paddingY), -1)
	lines := []string{}
	for i, room := range f.rooms {
		coord := roomCoord(f.sizes, i)
		x1 := coord[0] * svgRoomSize
		y1 := coord[1] * svgRoomSize
		if !room.openWalls[0] {
			x2 := coord[0] * svgRoomSize
			y2 := (coord[1] + 1) * svgRoomSize
			lines = append(lines,
				fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2))
		}
		if !room.openWalls[1] {
			x2 := (coord[0] + 1) * svgRoomSize
			y2 := coord[1] * svgRoomSize
			lines = append(lines,
				fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2))
		}
	}
	width := f.sizes[0] * svgRoomSize
	height := f.sizes[1] * svgRoomSize
	lines = append(lines, fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`,
		0, height, width, height))
	lines = append(lines, fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`,
		width, 0, width, height))
	svg = strings.Replace(svg, "{{lines}}", strings.Join(lines, "\n"), -1)
	io.WriteString(writer, svg)
}

func roomCoord(sizes [maxDimension]int, index int) [maxDimension]int {
	coord := [maxDimension]int{}
	for i := 0; i < len(sizes); i++ {
		c := index
		for j := i - 1; 0 <= j; j-- {
			c /= sizes[j]
		}
		c %= sizes[i]
		coord[i] = c
	}
	return coord
}

func roomIndex(sizes [maxDimension]int, coord [maxDimension]int) int {
	index := 0
	for i := len(sizes) - 1; 0 <= i; i-- {
		index += coord[i]
		if 0 <= i-1 {
			index *= sizes[i-1]
		}
	}
	return index
}

func cluster(roomClusters []int, i int) int {
	for ; i != roomClusters[i]; i = roomClusters[i] {
	}
	return i
}

func allRoomsConnected(roomClusters []int) bool {
	for i := 0; i < len(roomClusters); i++ {
		if cluster(roomClusters, i) != 0 {
			return false
		}
	}
	return true
}

func Create(random *rand.Rand, width, height int) *Field {
	const dimNum = 2

	f := &Field{
		rooms: make([]Room, width*height),
		sizes: [maxDimension]int{width, height, 1, 1},
	}

	roomClusters := make([]int, len(f.rooms))
	for i := 0; i < len(roomClusters); i++ {
		roomClusters[i] = i
	}

	for !allRoomsConnected(roomClusters) {
		dim := 0
		rIndex := 0
		rCluster := 0
		nextRoomCluster := 0
		for {
			r := random.Intn(len(f.rooms) * dimNum)
			dim = r % dimNum
			rIndex = r / dimNum

			room := f.rooms[rIndex]
			if room.openWalls[dim] {
				continue
			}
			roomCoord := roomCoord(f.sizes, rIndex)
			nextRoomCoord := [maxDimension]int{}
			copy(nextRoomCoord[:], roomCoord[:])
			nextRoomCoord[dim] -= 1
			if nextRoomCoord[dim] < 0 {
				continue
			}
			nextRoomIndex := roomIndex(f.sizes, nextRoomCoord)
			rCluster = cluster(roomClusters, rIndex)
			nextRoomCluster = cluster(roomClusters, nextRoomIndex)
			if rCluster == nextRoomCluster {
				continue
			}
			break
		}

		room := &f.rooms[rIndex]
		room.openWalls[dim] = true
		if rCluster < nextRoomCluster {
			roomClusters[nextRoomCluster] = rCluster
		} else {
			roomClusters[rCluster] = nextRoomCluster
		}
	}

	return f
}
