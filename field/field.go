package field

import (
	"math/rand"
	"io"
	"strings"
	"time"
)

const maxDimension = 4

type Room struct {
	openWalls [maxDimension]bool
}

type Field struct {
	rooms []Room
	sizes []int
}

func (f *Field) roomCoord(index int) [maxDimension]int {
	coord := [maxDimension]int{}
	for i := 0; i < len(f.sizes); i++ {
		c := index
		for j := i - 1; 0 <= j; j-- {
			c /= f.sizes[j]
		}
		c %= f.sizes[i]
		coord[i] = c
	}
	return coord
}

func (f *Field) roomIndex(coord [maxDimension]int) int {
	index := 0
	for i := len(f.sizes) - 1; 0 <= i; i-- {
		index += coord[i]
		if 0 <= i-1 {
			index *= f.sizes[i-1]
		}
	}
	return index
}

func (f *Field) Write(writer io.Writer) {
	for j := 0; j < f.sizes[1]; j++ {
		line1 := ""
		line2 := ""
		for i := 0; i < f.sizes[0]; i++ {
			room := f.rooms[f.roomIndex([maxDimension]int{i, j})]
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
		// TODO: Error handling
		io.WriteString(writer, line1)
		io.WriteString(writer, line2)
	}
	line := strings.Repeat("+--", f.sizes[0]) + "+\n"
	// TODO: Error handling
	io.WriteString(writer, line)
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

func Create(width, height int) *Field {
	const dimNum = 2

	rand.Seed(time.Now().UnixNano())

	f := &Field{
		rooms: make([]Room, width*height),
		sizes: []int{width, height},
	}

	roomClusters := make([]int, len(f.rooms))
	for i := 0; i < len(roomClusters); i++ {
		roomClusters[i] = i
	}

	for !allRoomsConnected(roomClusters) {
		dim := 0
		roomIndex := 0
		roomCluster := 0
		nextRoomCluster := 0
		for {
			dim = rand.Intn(dimNum)
			roomIndex = rand.Intn(len(f.rooms))

			room := f.rooms[roomIndex]
			if room.openWalls[dim] {
				continue
			}
			roomCoord := f.roomCoord(roomIndex)
			nextRoomCoord := [maxDimension]int{}
			copy(nextRoomCoord[:], roomCoord[:])
			nextRoomCoord[dim] -= 1
			if nextRoomCoord[dim] < 0 {
				continue
			}
			nextRoomIndex := f.roomIndex(nextRoomCoord)
			roomCluster = cluster(roomClusters, roomIndex)
			nextRoomCluster = cluster(roomClusters, nextRoomIndex)
			if roomCluster == nextRoomCluster {
				continue
			}
			break
		}

		room := &f.rooms[roomIndex]
		room.openWalls[dim] = true
		if roomCluster < nextRoomCluster {
			roomClusters[nextRoomCluster] = roomCluster
		} else {
			roomClusters[roomCluster] = nextRoomCluster
		}
	}

	return f
}
