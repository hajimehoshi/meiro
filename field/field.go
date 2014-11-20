package field

import (
	"math/rand"
)

const maxDimension = 4

type Room struct {
	openWalls [maxDimension]bool
}

type Position [maxDimension]int

type Field struct {
	rooms         []Room
	sizes         [maxDimension]int
	startPosition Position
	endPosition   Position
	costs         []int
}

func roomPosition(sizes [maxDimension]int, index int) Position {
	coord := Position{}
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

func roomIndex(sizes [maxDimension]int, coord Position) int {
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

func (f *Field) nextRooms(index int) []int {
	nextIndexes := []int{}
	position := roomPosition(f.sizes, index)
	for i := 0; i < maxDimension; i++ {
		if !f.rooms[index].openWalls[i] {
			continue
		}
		nextPosition := position
		nextPosition[i]--
		nextIndexes = append(nextIndexes, roomIndex(f.sizes, nextPosition))
	}
	for i := 0; i < maxDimension; i++ {
		nextPosition := position
		nextPosition[i]++
		if f.sizes[i] <= nextPosition[i] {
			continue
		}
		nextIndex := roomIndex(f.sizes, nextPosition)
		if !f.rooms[nextIndex].openWalls[i] {
			continue
		}
		nextIndexes = append(nextIndexes, nextIndex)
	}
	return nextIndexes
}

func (f *Field) calcCosts() {
	currentPositions := map[Position]struct{}{f.startPosition: struct{}{}}
	for i := 0; 0 < len(currentPositions); i++ {
		nextPositions := map[Position]struct{}{}
		for position, _ := range currentPositions {
			index := roomIndex(f.sizes, position) 
			f.costs[index] = i
			for _, nextIndex := range f.nextRooms(index) {
				nextPosition := roomPosition(f.sizes, nextIndex)
				if nextPosition == f.startPosition {
					continue
				}
				if 0 < f.costs[nextIndex] {
					continue
				}
				nextPositions[nextPosition] = struct{}{}
			}
		}
		currentPositions = nextPositions
	}
}

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	f := &Field{
		rooms:       make([]Room, size1*size2*size3*size4),
		sizes:       [maxDimension]int{size1, size2, size3, size4},
		endPosition: Position{size1 - 1, size2 - 1, size3 - 1, size4 - 1},
		costs:       make([]int, size1*size2*size3*size4),
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
			roomPosition := roomPosition(f.sizes, rIndex)
			nextRoomPosition := Position{}
			copy(nextRoomPosition[:], roomPosition[:])
			nextRoomPosition[dim] -= 1
			if nextRoomPosition[dim] < 0 {
				continue
			}
			nextRoomIndex := roomIndex(f.sizes, nextRoomPosition)
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

	f.calcCosts()

	return f
}
