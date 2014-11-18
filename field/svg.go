package field

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

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

func svgLine(x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2)
}

func (f *Field) WriteSVG(writer io.Writer) {
	// TODO: Expand this to 3D and 4D
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
			lines = append(lines, svgLine(x1, y1, x2, y2))
		}
		if !room.openWalls[1] {
			x2 := (coord[0] + 1) * svgRoomSize
			y2 := coord[1] * svgRoomSize
			lines = append(lines, svgLine(x1, y1, x2, y2))
		}
	}
	width := f.sizes[0] * svgRoomSize
	height := f.sizes[1] * svgRoomSize
	lines = append(lines, svgLine(0, height, width, height))
	lines = append(lines, svgLine(width, 0, width, height))
	svg = strings.Replace(svg, "{{lines}}", strings.Join(lines, "\n"), -1)
	io.WriteString(writer, svg)
}
