package grpclb

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type HashFunc func(data []byte) uint32

const (
	DefaultReplicas = 10
)

type Ketama struct {
	sync.Mutex
	hash     HashFunc
	replicas int
	keys     []int //  Sorted keys
	hashMap  map[int]string
}

func NewKetama(replicas int, fn HashFunc) *Ketama {
	h := &Ketama{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if h.replicas <= 0 {
		h.replicas = DefaultReplicas
	}
	if h.hash == nil {
		h.hash = crc32.ChecksumIEEE
	}
	return h
}

func (h *Ketama) IsEmpty() bool {
	h.Lock()
	defer h.Unlock()

	return len(h.keys) == 0
}

func (h *Ketama) Add(nodes ...string) {
	h.Lock()
	defer h.Unlock()

	for _, node := range nodes {
		fmt.Println("kk======add node :", node)
		for i := 0; i < h.replicas; i++ {
			key := int(h.hash([]byte(strconv.Itoa(i) + node + "goodluck")))

			if _, ok := h.hashMap[key]; !ok {
				h.keys = append(h.keys, key)
			}
			h.hashMap[key] = node
		}
	}
	sort.Ints(h.keys)
}

func (h *Ketama) Remove(nodes ...string) {
	h.Lock()
	defer h.Unlock()

	deletedKey := make([]int, 0)
	for _, node := range nodes {
		for i := 0; i < h.replicas; i++ {
			key := int(h.hash([]byte(strconv.Itoa(i) + node)))

			if _, ok := h.hashMap[key]; ok {
				deletedKey = append(deletedKey, key)
				delete(h.hashMap, key)
			}
		}
	}
	if len(deletedKey) > 0 {
		h.deleteKeys(deletedKey)
	}
}

func (h *Ketama) deleteKeys(deletedKeys []int) {
	sort.Ints(deletedKeys)

	index := 0
	count := 0
	for _, key := range deletedKeys {
		for ; index < len(h.keys); index++ {
			h.keys[index-count] = h.keys[index]

			if key == h.keys[index] {
				count++
				index++
				break
			}
		}
	}

	for ; index < len(h.keys); index++ {
		h.keys[index-count] = h.keys[index]
	}

	h.keys = h.keys[:len(h.keys)-count]
}

func (h *Ketama) Get(key string) (string, bool) {
	if h.IsEmpty() {
		return "", false
	}

	hash := int(h.hash([]byte(key)))

	h.Lock()
	defer h.Unlock()

	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	fmt.Println("iii=======hashkey index:", idx)
	if idx == len(h.keys) {
		idx = 0
	}
	str, ok := h.hashMap[h.keys[idx]]

	fmt.Println("iii=======hashkey:", str)
	return str, ok
}
