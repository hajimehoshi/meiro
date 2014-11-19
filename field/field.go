package field

import (
	"math/rand"
)

const maxDimension = 4

type Room struct {
	openWalls [maxDimension]bool
}

type Field struct {
	rooms []Room
	sizes [maxDimension]int
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

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	f := &Field{
		rooms: make([]Room, size1*size2*size3*size4),
		sizes: [maxDimension]int{size1, size2, size3, size4},
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
			r := random.Intn(len(f.rooms) * maxDimension)
			dim = r % maxDimension
			rIndex = r / maxDimension

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
