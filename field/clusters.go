package field

type clusters struct {
	clusters       []int32
	path           []int32
	allSameChecked int
}

func newClusters(num int32) *clusters {
	c := &clusters{
		clusters: make([]int32, num),
		path:     make([]int32, 0, 8),
	}
	for i := int32(1); i < num; i++ {
		c.clusters[i] = i
	}
	return c
}

func (c *clusters) Get(i int32) int32 {
	if c.clusters[i] == 0 {
		return 0
	}
	cluster := c.clusters[i]
	path := c.path[:0]
	for {
		cluster := c.clusters[i]
		if i == cluster {
			break
		}
		path = append(path, i)
		i = cluster
	}
	cluster = i
	for _, i := range path {
		c.clusters[i] = cluster
	}
	return cluster
}

func (c *clusters) Set(oldCluster, newCluster int32) {
	c.clusters[oldCluster] = newCluster
}

func (c *clusters) AllSame() bool {
	for i := c.allSameChecked; i < len(c.clusters); i++ {
		if c.Get(int32(i)) != 0 {
			return false
		}
		c.allSameChecked++
	}
	return true
}
