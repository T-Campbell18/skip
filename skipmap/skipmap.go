package skipmap

import (
	"cmp"
	"math/rand/v2"
	"sync"
)

const (
	DefaultMaxLevel    = 32
	DefaultProbability = 0.5
)

type Node[K any, V any] struct {
	key     K
	value   V
	forward []*Node[K, V]
}

type SkipMap[K any, V any] struct {
	header      *Node[K, V]
	maxLevel    int
	level       int
	length      int
	probability float32
	mu          sync.RWMutex
	comparator  func(a, b K) int

	updateCache []*Node[K, V]
}

func New[K cmp.Ordered, V any]() *SkipMap[K, V] {
	return NewWithComparator[K, V](cmp.Compare[K])
}

func NewWithComparator[K any, V any](comparator func(a, b K) int) *SkipMap[K, V] {
	return &SkipMap[K, V]{
		header:      &Node[K, V]{forward: make([]*Node[K, V], DefaultMaxLevel)},
		maxLevel:    DefaultMaxLevel,
		level:       0,
		length:      0,
		probability: DefaultProbability,
		comparator:  comparator,
		updateCache: make([]*Node[K, V], DefaultMaxLevel),
	}
}

func defaultComparator[K cmp.Ordered](a, b K) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func (s *SkipMap[K, V]) randomLevel() int {
	level := 0
	for rand.Float32() < s.probability && level < s.maxLevel-1 {
		level++
	}
	return level
}

func (s *SkipMap[K, V]) Put(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update := s.updateCache
	current := s.header

	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && s.comparator(current.forward[i].key, key) < 0 {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current != nil && s.comparator(current.key, key) == 0 {
		current.value = value
		return
	}

	newLevel := s.randomLevel()

	if newLevel > s.level {
		for i := s.level + 1; i <= newLevel; i++ {
			update[i] = s.header
		}
		s.level = newLevel
	}

	newNode := &Node[K, V]{
		key:     key,
		value:   value,
		forward: make([]*Node[K, V], newLevel+1),
	}

	for i := 0; i <= newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	s.length++
}

func (s *SkipMap[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	current := s.header

	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && s.comparator(current.forward[i].key, key) < 0 {
			current = current.forward[i]
		}
	}

	current = current.forward[0]

	if current != nil && s.comparator(current.key, key) == 0 {
		return current.value, true
	}
	var zero V
	return zero, false
}

func (s *SkipMap[K, V]) Remove(key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	update := s.updateCache
	current := s.header

	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && s.comparator(current.forward[i].key, key) < 0 {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current != nil && s.comparator(current.key, key) == 0 {
		for i := 0; i <= s.level; i++ {
			if update[i].forward[i] != current {
				break
			}
			update[i].forward[i] = current.forward[i]
		}

		for s.level > 0 && s.header.forward[s.level] == nil {
			s.level--
		}

		s.length--
		return true
	}

	return false
}

func (s *SkipMap[K, V]) Range(start, end K) []V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]V, 0)
	current := s.header

	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && s.comparator(current.forward[i].key, start) < 0 {
			current = current.forward[i]
		}
	}

	current = current.forward[0]

	for current != nil && s.comparator(current.key, end) <= 0 {
		result = append(result, current.value)
		current = current.forward[0]
	}

	return result
}

// RangeFunc iterates over the elements in the range [start, end] and calls f for each key-value pair.
// If f returns false, iteration stops.
func (s *SkipMap[K, V]) RangeFunc(start, end K, f func(key K, value V) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	current := s.header

	// Find the start node
	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && s.comparator(current.forward[i].key, start) < 0 {
			current = current.forward[i]
		}
	}

	current = current.forward[0]

	// Iterate and call the function until the end of the range or the callback returns false
	for current != nil && s.comparator(current.key, end) <= 0 {
		if !f(current.key, current.value) {
			break
		}
		current = current.forward[0]
	}
}

func (s *SkipMap[K, V]) Min() (K, V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var zeroK K
	var zeroV V

	if s.length == 0 {
		return zeroK, zeroV, false
	}

	minNode := s.header.forward[0]
	return minNode.key, minNode.value, true
}

func (s *SkipMap[K, V]) Max() (K, V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var zeroK K
	var zeroV V

	if s.length == 0 {
		return zeroK, zeroV, false
	}

	current := s.header
	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil {
			current = current.forward[i]
		}
	}

	return current.key, current.value, true
}

func (s *SkipMap[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.length
}

func (s *SkipMap[K, V]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.length == 0
}
