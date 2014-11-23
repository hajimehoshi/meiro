package field

import (
	"fmt"
	"io"
)

const svgRoomSize = 8
const paddingX = svgRoomSize
const paddingY = svgRoomSize

func writeSvgLine(writer io.Writer, x1, y1, x2, y2 int) {
	fmt.Fprintf(writer, `<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2)
	io.WriteString(writer, "\n")
}

func writeSvgDashedLine(writer io.Writer, x1, y1, x2, y2 int) {
	fmt.Fprintf(writer, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke-dasharray="2" stroke-opacity="0.3" />`, x1, y1, x2, y2)
	io.WriteString(writer, "\n")
}

func svgLine(x1, y1, x2, y2 int) string {
	return fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" />`, x1, y1, x2, y2)
}

func writeSvgArrows(writer io.Writer) {
	cx := svgRoomSize / 2
	cy := svgRoomSize / 2
	writeSvgLine(writer, cx, cy, cx, svgRoomSize - 1)
	writeSvgLine(writer, 3, cy + 2, cx, svgRoomSize - 1)
	writeSvgLine(writer, svgRoomSize - 3, cy + 2, cx, svgRoomSize - 1)
}

func (f *Field) svgFloorWidth() int {
	return f.sizes[0]*svgRoomSize + 2*paddingX
}

func (f *Field) svgFloorHeight() int {
	return f.sizes[1]*svgRoomSize + 2*paddingY
}

func (f *Field) writeSvgFloor(writer io.Writer, dim3, dim4 int) {
	offsetX := dim4*f.svgFloorWidth() + paddingX
	offsetY := dim3*f.svgFloorHeight() + paddingY

	fmt.Fprintf(writer, `<g transform="translate(%d %d)">`, offsetX, offsetY)
	io.WriteString(writer, "\n")

	for dim2 := 0; dim2 < f.sizes[1]; dim2++ {
		for dim1 := 0; dim1 < f.sizes[0]; dim1++ {
			coord := Position{dim1, dim2, dim3, dim4}
			room := f.rooms[roomIndex(f.sizes, coord)]
			x1 := dim1 * svgRoomSize
			y1 := dim2 * svgRoomSize
			if !room.openWalls[0] {
				x2 := dim1 * svgRoomSize
				y2 := (dim2 + 1) * svgRoomSize
				writeSvgLine(writer, x1, y1, x2, y2)
			}
			if !room.openWalls[1] {
				x2 := (dim1 + 1) * svgRoomSize
				y2 := dim2 * svgRoomSize
				writeSvgLine(writer, x1, y1, x2, y2)
			}
			if room.openWalls[2] {
				cx := svgRoomSize / 2
				cy := svgRoomSize / 2
				fmt.Fprintf(writer, `<use xlink:href="#arrow" transform="translate(%d, %d) rotate(180, %d, %d)" />`,
					x1, y1, cx, cy)
				io.WriteString(writer, "\n")
			}
			if room.openWalls[3] {
				cx := svgRoomSize / 2
				cy := svgRoomSize / 2
				fmt.Fprintf(writer, `<use xlink:href="#arrow" transform="translate(%d, %d) rotate(90, %d, %d)" />`,
					x1, y1, cx, cy)
				io.WriteString(writer, "\n")
			}

			nextCoord := coord
			nextCoord[2]++
			if nextCoord[2] < f.sizes[2] {
				if f.rooms[roomIndex(f.sizes, nextCoord)].openWalls[2] {
					fmt.Fprintf(writer, `<use xlink:href="#arrow" transform="translate(%d, %d)" />`,
						x1, y1)
					io.WriteString(writer, "\n")
				}
			}

			nextCoord = coord
			nextCoord[3]++
			if nextCoord[3] < f.sizes[3] {
				if f.rooms[roomIndex(f.sizes, nextCoord)].openWalls[3] {
					cx := svgRoomSize / 2
					cy := svgRoomSize / 2
					fmt.Fprintf(writer, `<use xlink:href="#arrow" transform="translate(%d, %d) rotate(270, %d, %d)" />`,
						x1, y1, cx, cy)
					io.WriteString(writer, "\n")
				}
			}

		}
	}

	width := f.sizes[0] * svgRoomSize
	height := f.sizes[1] * svgRoomSize
	writeSvgLine(writer, 0, height, width, height)
	writeSvgLine(writer, width, 0, width, height)

	fmt.Fprintln(writer, `</g>`)
}

func (f *Field) WriteSVG(writer io.Writer) {
	width := f.svgFloorWidth() * f.sizes[3]
	height := f.svgFloorHeight() * f.sizes[2]

	fmt.Fprintf(writer, `<?xml version="1.0" encoding="utf-8" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" viewBox="0 0 %d %d" background-color="#fff">
`, width, height);

	fmt.Fprintln(writer, `<defs>`)
	fmt.Fprintln(writer, `<symbol id="arrow" stroke-width="0.5">`)
	writeSvgArrows(writer)
	fmt.Fprintln(writer, `</symbol>`)
	fmt.Fprintln(writer, `</defs>`)

	fmt.Fprintln(writer, `<g stroke="black" stroke-width="1" stroke-linecap="round">`)
	for dim4 := 0; dim4 < f.sizes[3]; dim4++ {
		for dim3 := 0; dim3 < f.sizes[2]; dim3++ {
			f.writeSvgFloor(writer, dim3, dim4)
		}
	}
	fmt.Fprintln(writer, `</g>`)

	fmt.Fprintln(writer, `<g stroke="red" stroke-width="1" stroke-linecap="round">`)
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
		if position[2] == nextPosition[2] && position[3] == nextPosition[3] {
			writeSvgLine(writer, x1, y1, x2, y2)
		} else {
			writeSvgDashedLine(writer, x1, y1, x2, y2)
		}
	}

	fmt.Fprintln(writer, `</g>`)

	fmt.Fprintln(writer, `</svg>`)
}
