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

func svgArrow() string {
	lines := []string{}
	cx := svgRoomSize / 2
	cy := svgRoomSize / 2
	lines = append(lines, svgLine(cx, cy, cx, svgRoomSize - 1))
	lines = append(lines, svgLine(3, cy + 2, cx, svgRoomSize - 1))
	lines = append(lines, svgLine(svgRoomSize - 3, cy + 2, cx, svgRoomSize - 1))
	return strings.Join(lines, "\n")
}

func (f *Field) svgRoomWidth() int {
	return f.sizes[0]*svgRoomSize + 2*paddingX
}

func (f *Field) svgRoomHeight() int {
	return f.sizes[1]*svgRoomSize + 2*paddingY
}

func (f *Field) svgRoom(dim3, dim4 int) string {
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
			coord := [maxDimension]int{dim1, dim2, dim3, dim4}
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

	offsetX := dim4*f.svgRoomWidth() + paddingX
	offsetY := dim3*f.svgRoomHeight() + paddingY

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
    {{rooms}}
  </g>
</svg>
`[1:]

	width := f.svgRoomWidth() * f.sizes[3]
	height := f.svgRoomHeight() * f.sizes[2]

	svg := tmpl
	svg = strings.Replace(svg, "{{width}}", strconv.Itoa(width), -1)
	svg = strings.Replace(svg, "{{height}}", strconv.Itoa(height), -1)
	svg = strings.Replace(svg, "{{arrow}}", svgArrow(), -1)
	rooms := []string{}
	for dim4 := 0; dim4 < f.sizes[3]; dim4++ {
		for dim3 := 0; dim3 < f.sizes[2]; dim3++ {
			rooms = append(rooms, f.svgRoom(dim3, dim4))
		}
	}
	svg = strings.Replace(svg, "{{rooms}}", strings.Join(rooms, "\n"), -1)
	io.WriteString(writer, svg)
}
