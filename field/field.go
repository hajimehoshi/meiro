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
	rooms       []Room
	sizes       [maxDimension]int
	startIndex  int
	endIndex    int
	costs       []int
	parentRooms []int
}

func roomPosition(sizes [maxDimension]int, index int) Position {
	coord := Position{}
	coord[0] = index
	for i := 1; i < len(sizes); i++ {
		coord[i] = coord[i-1] / sizes[i-1]
	}
	for i, size := range sizes {
		coord[i] %= size
	}
	return coord
}

func roomIndex(sizes [maxDimension]int, coord Position) int {
	index := coord[maxDimension-1]
	for i := len(sizes) - 2; 0 <= i; i-- {
		index *= sizes[i]
		index += coord[i]
	}
	return index
}

func roomCluster(roomClusters []int, i int) int {
	path := make([]int, 1, 4)
	path[0] = i
	for ; i != roomClusters[i]; i = roomClusters[i] {
		path = append(path, i)
	}
	cluster := i
	for _, i := range path {
		roomClusters[i] = cluster
	}
	return cluster
}

func allRoomsConnected(roomClusters []int) bool {
	for i, _ := range roomClusters {
		cluster := roomCluster(roomClusters, i)
		if cluster != 0 {
			roomClusters[i] = cluster
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
		if nextPosition := position; nextPosition[i] != 0 {
			nextPosition[i]--
			nextIndexes[len] = roomIndex(f.sizes, nextPosition)
			len++
		}
		if nextPosition := position; nextPosition[i] != f.sizes[i]-1 {
			nextPosition[i]++
			nextIndexes[len] = roomIndex(f.sizes, nextPosition)
			len++
		}
	}
	return nextIndexes, len
}

func (f *Field) nextConnectedRooms(index int) ([maxDimension * 2]int, int) {
	offsets := [maxDimension]int{}
	for i := 0; i < maxDimension; i++ {
		offset := 1
		for j := 0; j < i; j++ {
			offset *= f.sizes[j]
		}
		offsets[i] = offset
	}

	nextIndexes := [maxDimension * 2]int{}
	nextIndexesLen := 0
	roomsLen := len(f.rooms)
	for i := 0; i < maxDimension; i++ {
		if f.rooms[index].openWalls[i] {
			nextIndexes[nextIndexesLen] = index - offsets[i]
			nextIndexesLen++
		}
		nextIndex := index + offsets[i]
		if roomsLen <= nextIndex {
			continue
		}
		if !f.rooms[nextIndex].openWalls[i] {
			continue
		}
		nextIndexes[nextIndexesLen] = nextIndex
		nextIndexesLen++
	}
	return nextIndexes, nextIndexesLen
}

func (f *Field) calcCosts() {
	startIndex := f.startIndex
	currentIndexes := []int{startIndex}
	nextIndexes := []int{}
	f.parentRooms[startIndex] = -1
	for cost := 0; 0 < len(currentIndexes); cost++ {
		nextIndexes = nextIndexes[:0]
		for _, index := range currentIndexes {
			f.costs[index] = cost
			rooms, len := f.nextConnectedRooms(index)
			for _, nextIndex := range rooms[:len] {
				if nextIndex == startIndex {
					continue
				}
				if 0 < f.costs[nextIndex] {
					continue
				}
				nextIndexes = append(nextIndexes, nextIndex)
				f.parentRooms[nextIndex] = index
			}
		}
		diff := len(nextIndexes) - len(currentIndexes)
		if 0 < diff {
			currentIndexes = append(currentIndexes, make([]int, diff)...)
		}
		copy(currentIndexes, nextIndexes)
		currentIndexes = currentIndexes[:len(nextIndexes)]
	}
}

func (f *Field) deadEnds() []int {
	deadEnds := []int{}
	for i, _ := range f.rooms {
		if _, len := f.nextConnectedRooms(i); len == 1 {
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
		if _, roomsLen := f.nextConnectedRooms(deadEnd); roomsLen != -1 {
			continue
		}
		smallEnd := f.isSmallDeadEnd(deadEnd)
		nextRooms, nextRoomsLen := f.nextRooms(deadEnd)
		for _, nextRoom := range nextRooms[:nextRoomsLen] {
			nextSmallEnd := f.isSmallDeadEnd(nextRoom)
			if !nextSmallEnd {
				if !smallEnd {
					continue
				}
				if _, roomsLen := f.nextConnectedRooms(nextRoom); roomsLen != -1 {
					continue
				}
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
	index := f.endIndex
	for {
		shortestPath = append(shortestPath, index)
		nextIndex := f.parentRooms[index]
		if nextIndex == -1 {
			break
		}
		index = nextIndex
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
	position := roomPosition(f.sizes, index)
	for i := 0; i < maxDimension; i++ {
		if f.rooms[index].openWalls[i] {
			nextRoomPosition := position
			nextRoomPosition[i]++
			if f.sizes[i] <= nextRoomPosition[i] {
				return -1
			}
			return roomIndex(f.sizes, nextRoomPosition)
		}
		connectedRoomPosition := position
		connectedRoomPosition[i]++
		connectedRoomIndex := roomIndex(f.sizes, connectedRoomPosition)
		if len(f.rooms) <= connectedRoomIndex {
			continue
		}
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

func (f *Field) costToShortestPath() ([]int, []int) {
	inShortestPath := make([]bool, len(f.rooms))
	for _, index := range f.shortestPath() {
		inShortestPath[index] = true
	}

	costToShortestPath := make([]int, len(f.rooms))
	copy(costToShortestPath, f.costs)
	nearestRoomInShortestPath := make([]int, len(f.rooms))

	for _, shortestPathIndex := range f.shortestPath() {
		currentIndexes := []int{shortestPathIndex}
		nextIndexes := []int{}
		for cost := 0; 0 < len(currentIndexes); cost++ {
			nextIndexes = nextIndexes[:0]
			for _, index := range currentIndexes {
				costToShortestPath[index] = cost
				nearestRoomInShortestPath[index] = shortestPathIndex
				rooms, len := f.nextConnectedRooms(index)
				for _, nextIndex := range rooms[:len] {
					if inShortestPath[nextIndex] {
						continue
					}
					if costToShortestPath[nextIndex] <= cost {
						continue
					}
					nextIndexes = append(nextIndexes, nextIndex)
				}
			}
			diff := len(nextIndexes) - len(currentIndexes)
			if 0 < diff {
				currentIndexes = append(currentIndexes, make([]int, diff)...)
			}
			copy(currentIndexes, nextIndexes)
			currentIndexes = currentIndexes[:len(nextIndexes)]
		}
	}
	return costToShortestPath, nearestRoomInShortestPath
}

func (f *Field) createLoops(random *rand.Rand) {
	costToShortestPath, nearestRoomInShortestPath := f.costToShortestPath()

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
		if c <= (a+b)/4 && 0 < (a+b)%3 {
			f.connectRooms(deadEnd, nextRoom)
		}
	}
}

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	f := &Field{
		rooms:       make([]Room, size1*size2*size3*size4),
		sizes:       [maxDimension]int{size1, size2, size3, size4},
		costs:       make([]int, size1*size2*size3*size4),
		parentRooms: make([]int, size1*size2*size3*size4),
	}
	f.endIndex = roomIndex(f.sizes, Position{size1 - 1, size2 - 1, size3 - 1, size4 - 1})

	roomClusters := make([]int, len(f.rooms))
	for i, _ := range roomClusters {
		roomClusters[i] = i
	}

	for !allRoomsConnected(roomClusters) {
		dim := 0
		index := 0
		cluster := 0
		nextRoomCluster := 0
		for {
			r := random.Intn(len(f.rooms) * maxDimension)
			dim = r % maxDimension
			index = r / maxDimension
			if f.rooms[index].openWalls[dim] {
				continue
			}
			nextRoomPosition := roomPosition(f.sizes, index)
			nextRoomPosition[dim]--
			if nextRoomPosition[dim] < 0 {
				continue
			}
			nextRoomIndex := roomIndex(f.sizes, nextRoomPosition)
			cluster = roomCluster(roomClusters, index)
			nextRoomCluster = roomCluster(roomClusters, nextRoomIndex)
			if cluster == nextRoomCluster {
				continue
			}
			break
		}

		f.rooms[index].openWalls[dim] = true
		if cluster < nextRoomCluster {
			roomClusters[nextRoomCluster] = cluster
		} else {
			roomClusters[cluster] = nextRoomCluster
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
