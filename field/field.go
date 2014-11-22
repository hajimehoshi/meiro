package field

import (
	"math/rand"
	"sort"
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
		nextPosition := position
		nextPosition[i]--
		if nextPosition[i] < 0 {
			continue
		}
		nextIndexes = append(nextIndexes, roomIndex(f.sizes, nextPosition))
	}
	for i := 0; i < maxDimension; i++ {
		nextPosition := position
		nextPosition[i]++
		if f.sizes[i] <= nextPosition[i] {
			continue
		}
		nextIndexes = append(nextIndexes, roomIndex(f.sizes, nextPosition))
	}
	return nextIndexes
}

func (f *Field) nextConnectedRooms(index int) []int {
	// TODO: Unify with nextRooms
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

func (f *Field) parentRoom(index int) int {
	for _, nextIndex := range f.nextConnectedRooms(index) {
		if f.costs[nextIndex] != f.costs[index]-1 {
			continue
		}
		return nextIndex
	}
	return -1
}

func (f *Field) calcCosts() {
	currentPositions := map[Position]struct{}{f.startPosition: struct{}{}}
	for i := 0; 0 < len(currentPositions); i++ {
		nextPositions := map[Position]struct{}{}
		for position, _ := range currentPositions {
			index := roomIndex(f.sizes, position)
			f.costs[index] = i
			for _, nextIndex := range f.nextConnectedRooms(index) {
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

func (f *Field) shortestPath() []int {
	shortestPath := []int{}
	position := f.endPosition
	for {
		index := roomIndex(f.sizes, position)
		shortestPath = append(shortestPath, index)
		nextIndex := f.parentRoom(index)
		if nextIndex == -1 {
			break
		}
		position = roomPosition(f.sizes, nextIndex)
	}
	return shortestPath
}

func (f *Field) deadEnds() []int {
	deadEnds := []int{}
	for i, _ := range f.rooms {
		if len(f.nextConnectedRooms(i)) == 1 {
			deadEnds = append(deadEnds, i)
		}
	}
	return deadEnds
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

type sortingInts struct {
	values []int
	less   func(i, j int) bool
}

func (s *sortingInts) Len() int {
	return len(s.values)
}

func (s *sortingInts) Less(i, j int) bool {
	return s.less(i, j)
}

func (s *sortingInts) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func (f *Field) removeDeadEnds(random *rand.Rand) {
	inShortestPath := map[int]struct{}{}
	for _, index := range f.shortestPath() {
		inShortestPath[index] = struct{}{}
	}

	deadEnds := f.deadEnds()
	inDeadEnds := map[int]struct{}{}
	for _, index := range deadEnds {
		inDeadEnds[index] = struct{}{}
	}

	costToShortestPath := map[int]int{}
	nearestRoomInShortestPath := map[int]int{}
	startIndex := roomIndex(f.sizes, f.startPosition)
	for index, _ := range f.rooms {
		if index == startIndex {
			continue
		}
		cost := 0
		for parent := f.parentRoom(index); ; parent = f.parentRoom(parent) {
			cost++
			if _, ok := inShortestPath[parent]; ok {
				nearestRoomInShortestPath[index] = parent
				break
			}
		}
		costToShortestPath[index] = cost
	}

	for _, deadEnd := range deadEnds {
		if len(f.nextConnectedRooms(deadEnd)) != 1 {
			continue
		}
		nextRooms := f.nextRooms(deadEnd)

		sort.Sort(&sortingInts{nextRooms, func(i, j int) bool {
			nextRoom1 := nextRooms[i]
			nextRoom2 := nextRooms[j]

			b1 := costToShortestPath[nextRoom1]
			b2 := costToShortestPath[nextRoom2]
			c1 := abs(nearestRoomInShortestPath[nextRoom1] - nearestRoomInShortestPath[deadEnd])
			c2 := abs(nearestRoomInShortestPath[nextRoom2] - nearestRoomInShortestPath[deadEnd])
			return (b1 + c1) < (b2 + c2)
		}})

		for i := 0; i < len(nextRooms); i++ {
			nextRoom := nextRooms[i]
			a := costToShortestPath[deadEnd]
			b := costToShortestPath[nextRoom]
			c := abs(nearestRoomInShortestPath[nextRoom] - nearestRoomInShortestPath[deadEnd])
			if c <= abs(b - a) && 10 <= a && 10 <= b && 20 < abs(a - b) {
				f.connectRooms(deadEnd, nextRoom);
				break
			}
		}
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
	//f.removeDeadEnds(random)

	return f
}
