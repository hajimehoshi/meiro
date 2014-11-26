package field

type clusters struct {
	clusters []int
	numZero  int
	path     []int
}

func newClusters(num int) *clusters {
	c := &clusters{
		clusters: make([]int, num),
		numZero:  1,
		path:     make([]int, 0, 8),
	}
	for i := 1; i < num; i++ {
		c.clusters[i] = i
	}
	return c
}

func (c *clusters) Get(i int) int {
	if c.clusters[i] == 0 {
		return 0
	}
	cluster := c.clusters[i]
	if i == cluster {
		return cluster
	}
	if cluster == c.clusters[cluster] {
		return cluster
	}
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

func (c *clusters) Set(oldCluster, newCluster int) {
	c.clusters[oldCluster] = newCluster
}

func (c *clusters) AllSame() bool {
	for i, _ := range c.clusters {
		if c.Get(i) != 0 {
			return false
		}
	}
	return true
}
