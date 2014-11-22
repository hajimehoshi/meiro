package field

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const svgRoomSize = 8
const paddingX = svgRoomSize
const paddingY = svgRoomSize

func svgLine(x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2)
}

func svgDashedLine(x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke-dasharray="2" stroke-opacity="0.3" />`, x1, y1, x2, y2)
}

func svgArrow() string {
	lines := []string{}
	cx := svgRoomSize / 2
	cy := svgRoomSize / 2
	lines = append(lines, svgLine(cx, cy, cx, svgRoomSize - 1))
	lines = append(lines, svgLine(3, cy + 2, cx, svgRoomSize - 1))
	lines = append(lines, svgLine(svgRoomSize - 3, cy + 2, cx, svgRoomSize - 1))
	return strings.Join(lines, "\n")
}

func (f *Field) svgFloorWidth() int {
	return f.sizes[0]*svgRoomSize + 2*paddingX
}

func (f *Field) svgFloorHeight() int {
	return f.sizes[1]*svgRoomSize + 2*paddingY
}

func (f *Field) svgFloor(dim3, dim4 int) string {
	var tmpl = `
<g transform="translate({{offsetX}}, {{offsetY}})">
  {{lines}}
  {{arrows}}
</g>
`[1:]

	lines := []string{}
	arrows := []string{}

	for dim2 := 0; dim2 < f.sizes[1]; dim2++ {
		for dim1 := 0; dim1 < f.sizes[0]; dim1++ {
			coord := Position{dim1, dim2, dim3, dim4}
			room := f.rooms[roomIndex(f.sizes, coord)]
			x1 := dim1 * svgRoomSize
			y1 := dim2 * svgRoomSize
			if !room.openWalls[0] {
				x2 := dim1 * svgRoomSize
				y2 := (dim2 + 1) * svgRoomSize
				lines = append(lines, svgLine(x1, y1, x2, y2))
			}
			if !room.openWalls[1] {
				x2 := (dim1 + 1) * svgRoomSize
				y2 := dim2 * svgRoomSize
				lines = append(lines, svgLine(x1, y1, x2, y2))
			}
			if room.openWalls[2] {
				cx := svgRoomSize / 2
				cy := svgRoomSize / 2
				arrow := fmt.Sprintf(`<use xlink:href="#arrow" transform="translate(%d, %d) rotate(180, %d, %d)" />`,
					x1, y1, cx, cy)
				arrows = append(arrows, arrow)
			}
			if room.openWalls[3] {
				cx := svgRoomSize / 2
				cy := svgRoomSize / 2
				arrow := fmt.Sprintf(`<use xlink:href="#arrow" transform="translate(%d, %d) rotate(90, %d, %d)" />`,
					x1, y1, cx, cy)
				arrows = append(arrows, arrow)
			}

			nextCoord := coord
			nextCoord[2]++
			if nextCoord[2] < f.sizes[2] {
				if f.rooms[roomIndex(f.sizes, nextCoord)].openWalls[2] {
					arrow := fmt.Sprintf(`<use xlink:href="#arrow" transform="translate(%d, %d)" />`,
						x1, y1)
					arrows = append(arrows, arrow)
				}
			}

			nextCoord = coord
			nextCoord[3]++
			if nextCoord[3] < f.sizes[3] {
				if f.rooms[roomIndex(f.sizes, nextCoord)].openWalls[3] {
					cx := svgRoomSize / 2
					cy := svgRoomSize / 2
					arrow := fmt.Sprintf(`<use xlink:href="#arrow" transform="translate(%d, %d) rotate(270, %d, %d)" />`,
						x1, y1, cx, cy)
					arrows = append(arrows, arrow)
				}
			}

		}
	}

	width := f.sizes[0] * svgRoomSize
	height := f.sizes[1] * svgRoomSize
	lines = append(lines, svgLine(0, height, width, height))
	lines = append(lines, svgLine(width, 0, width, height))

	offsetX := dim4*f.svgFloorWidth() + paddingX
	offsetY := dim3*f.svgFloorHeight() + paddingY

	svg := tmpl
	svg = strings.Replace(svg, "{{offsetX}}", strconv.Itoa(offsetX), -1)
	svg = strings.Replace(svg, "{{offsetY}}", strconv.Itoa(offsetY), -1)
	svg = strings.Replace(svg, "{{lines}}", strings.Join(lines, "\n"), -1)
	svg = strings.Replace(svg, "{{arrows}}", strings.Join(arrows, "\n"), -1)
	return svg
}

func (f *Field) WriteSVG(writer io.Writer) {
	var tmpl = `
<?xml version="1.0" encoding="utf-8" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" viewBox="0 0 {{width}} {{height}}" background-color="#fff">
  <defs>
    <symbol id="arrow" stroke-width="0.5">
      {{arrow}}
    </symbol>
  </defs>
  <g stroke="black" stroke-width="1" stroke-linecap="round">
    {{floors}}
  </g>
  <g stroke="red" stroke-width="1" stroke-linecap="round">
    {{shortestPath}}
  </g>
</svg>
`[1:]

	width := f.svgFloorWidth() * f.sizes[3]
	height := f.svgFloorHeight() * f.sizes[2]

	svg := tmpl
	svg = strings.Replace(svg, "{{width}}", strconv.Itoa(width), -1)
	svg = strings.Replace(svg, "{{height}}", strconv.Itoa(height), -1)
	svg = strings.Replace(svg, "{{arrow}}", svgArrow(), -1)
	floors := []string{}
	for dim4 := 0; dim4 < f.sizes[3]; dim4++ {
		for dim3 := 0; dim3 < f.sizes[2]; dim3++ {
			floors = append(floors, f.svgFloor(dim3, dim4))
		}
	}
	svg = strings.Replace(svg, "{{floors}}", strings.Join(floors, "\n"), -1)

	shortestPathLines := []string{}
	shortestPath := f.shortestPath()
	for i := 0; i < len(shortestPath) - 1; i++ {
		index := shortestPath[i]
		nextIndex := shortestPath[i+1]
		position := roomPosition(f.sizes, index)
		nextPosition := roomPosition(f.sizes, nextIndex)
		x1 := position[3] * f.svgFloorWidth() + position[0] * svgRoomSize +
			svgRoomSize / 2 + paddingX
		y1 := position[2] * f.svgFloorHeight() + position[1] * svgRoomSize +
			svgRoomSize / 2 + paddingY
		x2 := nextPosition[3] * f.svgFloorWidth() + nextPosition[0] * svgRoomSize +
			svgRoomSize / 2 + paddingX
		y2 := nextPosition[2] * f.svgFloorHeight() + nextPosition[1] * svgRoomSize +
			svgRoomSize / 2 + paddingY
		line := ""
		if position[2] == nextPosition[2] && position[3] == nextPosition[3] {
			line = svgLine(x1, y1, x2, y2)
		} else {
			line = svgDashedLine(x1, y1, x2, y2)
		}
		shortestPathLines = append(shortestPathLines, line)
	}

	svg = strings.Replace(svg, "{{shortestPath}}", strings.Join(shortestPathLines, "\n"), -1)

	io.WriteString(writer, svg)
}
