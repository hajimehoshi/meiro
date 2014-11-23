package field

import (
	"math/rand"
)

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

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
	parentRooms   []int
}

func roomPosition(sizes [maxDimension]int, index int) Position {
	coord := Position{}
	coord[0] = index
	for i := 1; i < len(sizes); i++ {
		coord[i] = coord[i-1] / sizes[i-1]
	}
	for i := 0; i < len(sizes); i++ {
		coord[i] %= sizes[i]
	}
	return coord
}

func roomIndex(sizes [maxDimension]int, coord Position) int {
	index := coord[maxDimension - 1]
	for i := len(sizes) - 2; 0 <= i; i-- {
		index *= sizes[i]
		index += coord[i]
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

func (f *Field) nextRooms(index int) ([maxDimension * 2]int, int) {
	nextIndexes := [maxDimension * 2]int{}
	position := roomPosition(f.sizes, index)
	len := 0
	for i := 0; i < maxDimension; i++ {
		nextPosition := position
		nextPosition[i]--
		if nextPosition[i] < 0 {
			continue
		}
		nextIndexes[len] = roomIndex(f.sizes, nextPosition)
		len++
	}
	for i := 0; i < maxDimension; i++ {
		nextPosition := position
		nextPosition[i]++
		if f.sizes[i] <= nextPosition[i] {
			continue
		}
		nextIndexes[len] = roomIndex(f.sizes, nextPosition)
		len++
	}
	return nextIndexes, len
}

func (f *Field) nextConnectedRooms(index int) ([maxDimension * 2]int, int) {
	// TODO: Unify with nextRooms
	nextIndexes := [maxDimension * 2]int{}
	position := roomPosition(f.sizes, index)
	len := 0
	for i := 0; i < maxDimension; i++ {
		if !f.rooms[index].openWalls[i] {
			continue
		}
		nextPosition := position
		nextPosition[i]--
		nextIndexes[len] = roomIndex(f.sizes, nextPosition)
		len++
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
		nextIndexes[len] = nextIndex
		len++
	}
	return nextIndexes, len
}

func (f *Field) calcCosts() {
	currentPositions := map[Position]struct{}{f.startPosition: struct{}{}}
	f.parentRooms[roomIndex(f.sizes, f.startPosition)] = -1
	for i := 0; 0 < len(currentPositions); i++ {
		nextPositions := map[Position]struct{}{}
		for position, _ := range currentPositions {
			index := roomIndex(f.sizes, position)
			f.costs[index] = i
			rooms, len := f.nextConnectedRooms(index)
			for _, nextIndex := range rooms[:len] {
				nextPosition := roomPosition(f.sizes, nextIndex)
				if nextPosition == f.startPosition {
					continue
				}
				if 0 < f.costs[nextIndex] {
					continue
				}
				nextPositions[nextPosition] = struct{}{}
				f.parentRooms[nextIndex] = index
			}
		}
		currentPositions = nextPositions
	}
}

func (f *Field) deadEnds() []int {
	deadEnds := []int{}
	for i, _ := range f.rooms {
		_, len := f.nextConnectedRooms(i)
		if len == 1 {
			deadEnds = append(deadEnds, i)
		}
	}
	return deadEnds
}

func (f *Field) isSmallDeadEnd(index int) bool {
	rooms, roomsLen := f.nextConnectedRooms(index)
	if roomsLen != 1 {
		return false
	}
	rooms, roomsLen = f.nextConnectedRooms(rooms[0])
	if roomsLen == 2 {
		return false
	}
	return true
}

func (f *Field) reduceDeadEnds(random *rand.Rand) {
	for _, deadEnd := range f.deadEnds() {
		_, roomsLen := f.nextConnectedRooms(deadEnd)
		if roomsLen != 1 {
			continue
		}
		smallEnd := f.isSmallDeadEnd(deadEnd)
		nextRooms, nextRoomsLen := f.nextRooms(deadEnd)
		for _, nextRoom := range nextRooms[:nextRoomsLen] {
			_, roomsLen := f.nextConnectedRooms(nextRoom)
			if roomsLen != 1 {
				continue
			}
			nextSmallEnd := f.isSmallDeadEnd(nextRoom)
			if !smallEnd && !nextSmallEnd {
				continue
			}
			deadEndToRemove := deadEnd
			if smallEnd && nextSmallEnd {
				if random.Intn(2) == 0 {
					deadEndToRemove = nextRoom
				}
			} else if nextSmallEnd {
				deadEndToRemove = nextRoom
			}

			// Block all walls
			f.rooms[deadEndToRemove].openWalls = [maxDimension]bool{}
			position := roomPosition(f.sizes, deadEndToRemove)
			for i := 0; i < maxDimension; i++ {
				position := position
				position[i]++
				if f.sizes[i] <= position[i] {
					continue
				}
				f.rooms[roomIndex(f.sizes, position)].openWalls[i] = false
			}

			deadEndToExtend := deadEnd
			if deadEndToRemove == deadEnd {
				deadEndToExtend = nextRoom
			}
			f.connectRooms(deadEndToExtend, deadEndToRemove)
			break
		}
	}
}

func (f *Field) shortestPath() []int {
	shortestPath := []int{}
	position := f.endPosition
	for {
		index := roomIndex(f.sizes, position)
		shortestPath = append(shortestPath, index)
		nextIndex := f.parentRooms[index]
		if nextIndex == -1 {
			break
		}
		position = roomPosition(f.sizes, nextIndex)
	}
	return shortestPath
}

func (f *Field) connectRooms(index1, index2 int) bool {
	position1 := roomPosition(f.sizes, index1)
	position2 := roomPosition(f.sizes, index2)
	for i := 0; i < maxDimension; i++ {
		position := position1
		position[i]--
		if position != position2 {
			continue
		}
		f.rooms[index1].openWalls[i] = true
		return true
	}
	for i := 0; i < maxDimension; i++ {
		position := position1
		position[i]++
		if position != position2 {
			continue
		}
		f.rooms[index2].openWalls[i] = true
		return true
	}
	return false
}

func (f *Field) oppositeRoomOfDeadEnd(index int) int {
	room := f.rooms[index]
	position := roomPosition(f.sizes, index)
	for i := 0; i < maxDimension; i++ {
		if !room.openWalls[i] {
			continue
		}
		nextRoomPosition := position
		nextRoomPosition[i]++
		if f.sizes[i] <= nextRoomPosition[i] {
			return -1
		}
		return roomIndex(f.sizes, nextRoomPosition)
	}
	for i := 0; i < maxDimension; i++ {
		connectedRoomPosition := position
		connectedRoomPosition[i]++
		connectedRoomIndex := roomIndex(f.sizes, connectedRoomPosition)
		if !f.rooms[connectedRoomIndex].openWalls[i] {
			continue
		}
		nextRoomPosition := position
		nextRoomPosition[i]--
		if nextRoomPosition[i] < 0 {
			return -1
		}
		return roomIndex(f.sizes, nextRoomPosition)
	}
	return -1
}

func (f *Field) createLoops(random *rand.Rand) {
	inShortestPath := make([]bool, len(f.rooms))
	for _, index := range f.shortestPath() {
		inShortestPath[index] = true
	}

	costToShortestPath := make([]int, len(f.rooms))
	nearestRoomInShortestPath := make([]int, len(f.rooms))
	startIndex := roomIndex(f.sizes, f.startPosition)
	for index, _ := range f.rooms {
		if index == startIndex {
			continue
		}
		cost := 0
		for parent := f.parentRooms[index]; ; parent = f.parentRooms[parent] {
			cost++
			if inShortestPath[parent] {
				nearestRoomInShortestPath[index] = parent
				break
			}
		}
		costToShortestPath[index] = cost
	}

	for _, deadEnd := range f.deadEnds() {
		if _, roomsLen := f.nextConnectedRooms(deadEnd); roomsLen != 1 {
			continue
		}
		nextRoom := f.oppositeRoomOfDeadEnd(deadEnd)
		if nextRoom == -1 {
			continue
		}

		a := costToShortestPath[deadEnd]
		b := costToShortestPath[nextRoom]
		c := abs(nearestRoomInShortestPath[nextRoom] - nearestRoomInShortestPath[deadEnd])
		if c <= (a+b)/4 && 0 < (a + b) % 3 {
			f.connectRooms(deadEnd, nextRoom)
		}
	}
}

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	f := &Field{
		rooms:       make([]Room, size1*size2*size3*size4),
		sizes:       [maxDimension]int{size1, size2, size3, size4},
		endPosition: Position{size1 - 1, size2 - 1, size3 - 1, size4 - 1},
		costs:       make([]int, size1*size2*size3*size4),
		parentRooms: make([]int, size1*size2*size3*size4),
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

	deadEndsNum := len(f.deadEnds())
	for {
		f.reduceDeadEnds(random)
		currentDeadEndNum := len(f.deadEnds())
		if deadEndsNum == currentDeadEndNum {
			break
		}
		deadEndsNum = currentDeadEndNum
	}
	f.calcCosts()
	f.createLoops(random)

	return f
}
