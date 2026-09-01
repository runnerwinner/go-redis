package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashFunc defines function to generate hash code
type HashFunc func(data []byte) uint32

// NodeMap stores nodes and you can pick node from NodeMap
type NodeMap struct {
	hashFunc    HashFunc
	replicas    int
	nodeHashs   []int // sorted
	nodehashMap map[int]string
}

const defaultReplicas = 100

// NewNodeMap creates a new NodeMap
func NewNodeMap(fn HashFunc) *NodeMap {
	return NewNodeMapWithReplicas(defaultReplicas, fn)
}

// NewNodeMapWithReplicas creates a new NodeMap with virtual nodes
func NewNodeMapWithReplicas(replicas int, fn HashFunc) *NodeMap {
	if replicas <= 0 {
		replicas = defaultReplicas
	}
	m := &NodeMap{
		hashFunc:    fn,
		replicas:    replicas,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty returns if there is no node in NodeMap
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// AddNode add the given nodes into consistent hash circle
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ {
			virtualKey := strconv.Itoa(i) + "-" + key
			hash := int(m.hashFunc([]byte(virtualKey)))
			m.nodeHashs = append(m.nodeHashs, hash)
			m.nodehashMap[hash] = key
		}
	}
	sort.Ints(m.nodeHashs)
}

// PickNode gets the closest item in the hash to the provided key.
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hashFunc([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})

	// Means we have cycled back to the first replica.
	if idx == len(m.nodeHashs) {
		idx = 0
	}

	return m.nodehashMap[m.nodeHashs[idx]]
}
