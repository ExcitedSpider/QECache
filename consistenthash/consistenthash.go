package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// We assume 2^32 possible key
type Hash func(data []byte) uint32

// KeyHashInfo constains the hash information used for
// mapping keys to distributed instances
type KeyHashInfo struct {
	hash Hash
	// controls how many virtual node are there in consistent hash
	vnodeScalar int
	// the keys that being hashed. keep it sorted.
	keys []int
	// records of vnode to physical node
	vnodeDict map[int]string
}

func New(vnodeScalar int, fn Hash) *KeyHashInfo {
	m := &KeyHashInfo{
		vnodeScalar: vnodeScalar,
		hash:        fn,
		vnodeDict:   make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *KeyHashInfo) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.vnodeScalar; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.vnodeDict[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *KeyHashInfo) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.vnodeDict[m.keys[idx%len(m.keys)]]
}
