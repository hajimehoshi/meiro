package field

import (
	"math/rand"
)

func abs(i int32) int32 {
	if i < 0 {
		return -i
	}
	return i
}

const MaxDimension = 4

// TODO: Make it unexported
type Room struct {
	openWalls [MaxDimension]bool
}

func (r *Room) OpenWall(dim int32) bool {
	return r.openWalls[dim]
}

func (r *Room) SetOpenWall(dim int32, open bool) {
	r.openWalls[dim] = open
}

func (r *Room) Block() {
	r.openWalls = [MaxDimension]bool{}
}

// TODO: Make it unexported
type Position [MaxDimension]int32

type Field struct {
	rooms       []Room
	sizes       [MaxDimension]int32
	offsets     [MaxDimension]int32
	startIndex  int32
	endIndex    int32
	costs       []int32
	parentRooms []int32
}

func roomPosition(sizes [MaxDimension]int32, index int32) Position {
	position := Position{}
	position[0] = index
	for i := 1; i < len(sizes); i++ {
		position[i] = position[i-1] / sizes[i-1]
	}
	for i, size := range sizes {
		position[i] %= size
	}
	return position
}

func roomIndex(sizes [MaxDimension]int32, position Position) int32 {
	index := position[MaxDimension-1]
	for i := len(sizes) - 2; 0 <= i; i-- {
		index *= sizes[i]
		index += position[i]
	}
	return index
}

func (f *Field) create(random *rand.Rand) {
	denoms := [MaxDimension]int32{}
	for dim := int32(0); dim < MaxDimension; dim++ {
		denom := int32(1)
		for i := int32(0); i < dim; i++ {
			denom *= f.sizes[i]
		}
		denoms[dim] = denom
	}

	roomClusters := newClusters(int32(len(f.rooms)))

	type wall struct {
		roomIndex int32
		dimension int32
	}
	walls := make([]wall, 0, len(f.rooms)*MaxDimension)
	for i := int32(0); i < int32(cap(walls)); i++ {
		index := i / MaxDimension
		dim := i % MaxDimension
		// Instead of roomPosition(f.sizes, index)[dim] == 0
		if (index/denoms[dim])%f.sizes[dim] == 0 {
			continue
		}
		walls = append(walls, wall{index, dim})
	}
	walls = walls[:len(walls):len(walls)]

	for !roomClusters.AllSame() {
		dim := int32(0)
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

func (f *Field) nextRooms(index int32) ([MaxDimension * 2]int32, int32) {
	nextIndexes := [MaxDimension * 2]int32{}
	position := roomPosition(f.sizes, index)
	len := int32(0)
	for i := 0; i < MaxDimension; i++ {
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

func (f *Field) nextConnectedRooms(index int32) ([MaxDimension * 2]int32, int32) {
	nextIndexes := [MaxDimension * 2]int32{}
	nextIndexesLen := int32(0)
	roomsLen := int32(len(f.rooms))
	for i :=int32(0); i < MaxDimension; i++ {
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
	currentIndexes := []int32{startIndex}
	nextIndexes := []int32{}
	f.parentRooms[startIndex] = -1
	for cost := int32(0); 0 < len(currentIndexes); cost++ {
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
			currentIndexes = append(currentIndexes, make([]int32, diff)...)
		}
		copy(currentIndexes, nextIndexes)
		currentIndexes = currentIndexes[:len(nextIndexes)]
		nextIndexes = nextIndexes[:0]
	}
}

func isDeadEndAndSmallEnd(f *Field, index int32) (bool, bool) {
	rooms, roomsLen := f.nextConnectedRooms(index)
	if roomsLen != 1 {
		return false, false
	}
	_, roomsLen = f.nextConnectedRooms(rooms[0])
	return true, 2 < roomsLen
}

func (f *Field) reduceDeadEnds(deadEnds []int32, random *rand.Rand) {
	for _, deadEnd := range deadEnds {
		if _, roomsLen := f.nextConnectedRooms(deadEnd); roomsLen == -1 {
			continue
		}
		_, smallEnd := isDeadEndAndSmallEnd(f, deadEnd)
		if !smallEnd {
			continue
		}
		nextRooms, nextRoomsLen := f.nextRooms(deadEnd)
		for _, nextRoom := range nextRooms[:nextRoomsLen] {
			nextDeadEnd, nextSmallEnd := isDeadEndAndSmallEnd(f, nextRoom)
			if !nextDeadEnd {
				continue
			}
			deadEndToRemove := deadEnd
			if nextSmallEnd {
				if random.Intn(2) == 0 {
					deadEndToRemove = nextRoom
				}
			}

			f.rooms[deadEndToRemove].Block()
			position := roomPosition(f.sizes, deadEndToRemove)
			for i := int32(0); i < MaxDimension; i++ {
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

func (f *Field) shortestPath() []int32 {
	shortestPath := []int32{}
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

func (f *Field) connectRooms(index1, index2 int32) bool {
	position1 := roomPosition(f.sizes, index1)
	position2 := roomPosition(f.sizes, index2)
	for i := int32(0); i < MaxDimension; i++ {
		position := position1
		position[i]--
		if position != position2 {
			continue
		}
		f.rooms[index1].SetOpenWall(i, true)
		return true
	}
	for i := int32(0); i < MaxDimension; i++ {
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

func (f *Field) oppositeRoomOfDeadEnd(index int32) int32 {
	position := roomPosition(f.sizes, index)
	for i := int32(0); i < MaxDimension; i++ {
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
		if int32(len(f.rooms)) <= connectedRoomIndex {
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

func (f *Field) costToShortestPath() ([]int32, []int32) {
	inShortestPath := make([]bool, len(f.rooms))
	for _, index := range f.shortestPath() {
		inShortestPath[index] = true
	}

	costToShortestPath := make([]int32, len(f.rooms))
	copy(costToShortestPath, f.costs)
	nearestRoomInShortestPath := make([]int32, len(f.rooms))

	for _, shortestPathIndex := range f.shortestPath() {
		currentIndexes := []int32{shortestPathIndex}
		nextIndexes := []int32{}
		for cost := int32(0); 0 < len(currentIndexes); cost++ {
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
				currentIndexes = append(currentIndexes, make([]int32, diff)...)
			}
			copy(currentIndexes, nextIndexes)
			currentIndexes = currentIndexes[:len(nextIndexes)]
			nextIndexes = nextIndexes[:0]
		}
	}
	return costToShortestPath, nearestRoomInShortestPath
}

func (f *Field) createLoops(deadEnds []int32, random *rand.Rand) {
	costToShortestPath, nearestRoomInShortestPath := f.costToShortestPath()

	for _, deadEnd := range deadEnds {
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
		if c <= (a+b)/4 && (a+b)%7 <= 2 {
			f.connectRooms(deadEnd, nextRoom)
		}
	}
}

func nextRoomOffsets(sizes [MaxDimension]int32) [MaxDimension]int32 {
	offsets := [MaxDimension]int32{1, 1, 1, 1}
	for i := int32(1); i < MaxDimension; i++ {
		offsets[i] = offsets[i-1] * sizes[i-1]
	}
	return offsets
}

func getDeadEnds(f *Field) []int32 {
	deadEnds := []int32{}
	for i, _ := range f.rooms {
		i := int32(i)
		if _, len := f.nextConnectedRooms(i); len == 1 {
			deadEnds = append(deadEnds, i)
		}
	}
	return deadEnds
}

func Create(random *rand.Rand, size1, size2, size3, size4 int) *Field {
	l := int32(size1 * size2 * size3 * size4)
	f := &Field{
		rooms:       make([]Room, l),
		sizes:       [MaxDimension]int32{int32(size1), int32(size2), int32(size3), int32(size4)},
		costs:       make([]int32, l), // TODO: Make this lazily?
		parentRooms: make([]int32, l),
	}
	f.offsets = nextRoomOffsets(f.sizes)
	f.endIndex = roomIndex(f.sizes, Position{int32(size1 - 1), int32(size2 - 1), int32(size3 - 1), int32(size4 - 1)})

	f.create(random)

	deadEnds := getDeadEnds(f)
	deadEndsNum := len(deadEnds)
	for {
		f.reduceDeadEnds(deadEnds, random)
		deadEnds = getDeadEnds(f)
		currentDeadEndNum := len(deadEnds)
		if deadEndsNum == currentDeadEndNum {
			break
		}
		deadEndsNum = currentDeadEndNum
	}
	f.calcCosts()
	f.createLoops(deadEnds, random)

	return f
}

func (f *Field) IsWallOpen(position []int, dim int) (bool, bool) {
	p := Position{int32(position[0]), int32(position[1]), int32(position[2]), int32(position[3])}
	index := roomIndex(f.sizes, p)
	openWall1 := f.rooms[index].openWalls[dim]
	nextPosition := p
	if nextPosition[dim] == f.sizes[dim] - 1 {
		return openWall1, false
	}
	nextPosition[dim]++
	nextIndex := roomIndex(f.sizes, nextPosition)
	openWall2 := f.rooms[nextIndex].openWalls[dim]
	return openWall1, openWall2
}

func (f *Field) StartPosition() []int {
	position := roomPosition(f.sizes, f.startIndex)
	return []int{int(position[0]), int(position[1]), int(position[2]), int(position[3])}
}

func (f *Field) EndPosition() []int {
	position := roomPosition(f.sizes, f.endIndex)
	return []int{int(position[0]), int(position[1]), int(position[2]), int(position[3])}
}
