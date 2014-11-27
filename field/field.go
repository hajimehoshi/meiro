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

func (r *Room) OpenWall(dim int) bool {
	return r.openWalls[dim]
}

func (r *Room) SetOpenWall(dim int, open bool) {
	r.openWalls[dim] = open
}

func (r *Room) Block() {
	r.openWalls = [maxDimension]bool{}
}

type Position [maxDimension]int

type Field struct {
	rooms       []Room
	sizes       [maxDimension]int
	offsets     [maxDimension]int
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

func (f *Field) create(random *rand.Rand) {
	denoms := [maxDimension]int{}
	for dim := 0; dim < maxDimension; dim++ {
		denom := 1
		for i := 0; i < dim; i++ {
			denom *= f.sizes[i]
		}
		denoms[dim] = denom
	}

	roomClusters := newClusters(int32(len(f.rooms)))

	type wall struct {
		roomIndex int
		dimension int
	}
	walls := make([]wall, 0, len(f.rooms)*maxDimension)
	for i := 0; i < cap(walls); i++ {
		index := i / maxDimension
		dim := i % maxDimension
		// Instead of roomPosition(f.sizes, index)[dim] == 0
		if (index / denoms[dim]) % f.sizes[dim] == 0 {
			continue
		}
		walls = append(walls, wall{index, dim})
	}
	walls = walls[:len(walls):len(walls)]

	for !roomClusters.AllSame() {
		dim := 0
		index := int32(0)
		cluster := int32(0)
		nextRoomCluster := int32(0)

		wallIndex := random.Intn(len(walls))
		for {
			w := walls[wallIndex]
			dim = w.dimension
			index = int32(w.roomIndex)

			nextRoomIndex := index - int32(f.offsets[dim])
			cluster = roomClusters.Get(index)
			nextRoomCluster = roomClusters.Get(nextRoomIndex)

			l := len(walls) - 1
			walls[wallIndex] = walls[l]
			walls = walls[:l:l]
			if l == 0 {
				panic("too many walls are broken")
			}
			if cluster == nextRoomCluster {
				wallIndex++
				wallIndex %= l
				continue
			}
			break
		}

		f.rooms[index].SetOpenWall(dim, true)
		if cluster < nextRoomCluster {
			roomClusters.Set(nextRoomCluster, cluster)
		} else {
			roomClusters.Set(cluster, nextRoomCluster)
		}
	}
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
	nextIndexes := [maxDimension * 2]int{}
	nextIndexesLen := 0
	roomsLen := len(f.rooms)
	for i := 0; i < maxDimension; i++ {
		if f.rooms[index].OpenWall(i) {
			nextIndexes[nextIndexesLen] = index - f.offsets[i]
			nextIndexesLen++
		}
		nextIndex := index + f.offsets[i]
		if roomsLen <= nextIndex {
			continue
		}
		if !f.rooms[nextIndex].OpenWall(i) {
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
		nextIndexes = nextIndexes[:0]
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
	_, roomsLen = f.nextConnectedRooms(rooms[0])
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
			f.rooms[deadEndToRemove].Block()
			position := roomPosition(f.sizes, deadEndToRemove)
			for i := 0; i < maxDimension; i++ {
				position := position
				position[i]++
				if f.sizes[i] <= position[i] {
					continue
				}
				f.rooms[roomIndex(f.sizes, position)].SetOpenWall(i, false)
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
		f.rooms[index1].SetOpenWall(i, true)
		return true
	}
	for i := 0; i < maxDimension; i++ {
		position := position1
		position[i]++
		if position != position2 {
			continue
		}
		f.rooms[index2].SetOpenWall(i, true)
		return true
	}
	return false
}

func (f *Field) oppositeRoomOfDeadEnd(index int) int {
	position := roomPosition(f.sizes, index)
	for i := 0; i < maxDimension; i++ {
		if f.rooms[index].OpenWall(i) {
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
		if !f.rooms[connectedRoomIndex].OpenWall(i) {
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
			nextIndexes = nextIndexes[:0]
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

func nextRoomOffsets(sizes [maxDimension]int) [maxDimension]int {
	offsets := [maxDimension]int{1, 1, 1, 1}
	for i := 1; i < maxDimension; i++ {
		offsets[i] = offsets[i-1] * sizes[i-1]
	}
	return offsets
}

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	f := &Field{
		rooms:       make([]Room, size1*size2*size3*size4),
		sizes:       [maxDimension]int{size1, size2, size3, size4},
		costs:       make([]int, size1*size2*size3*size4),
		parentRooms: make([]int, size1*size2*size3*size4),
	}
	f.offsets = nextRoomOffsets(f.sizes)
	f.endIndex = roomIndex(f.sizes, Position{size1 - 1, size2 - 1, size3 - 1, size4 - 1})

	f.create(random)

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
