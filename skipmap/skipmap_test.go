package skipmap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestSkipMap_PutAndGet(t *testing.T) {
	s := New[int, string]()

	// Test Put and Get
	s.Put(1, "one")
	s.Put(2, "two")
	s.Put(3, "three")

	if val, ok := s.Get(1); !ok || val != "one" {
		t.Errorf("Expected to get 'one' for key 1, but got '%s'", val)
	}
	if val, ok := s.Get(2); !ok || val != "two" {
		t.Errorf("Expected to get 'two' for key 2, but got '%s'", val)
	}
	if val, ok := s.Get(3); !ok || val != "three" {
		t.Errorf("Expected to get 'three' for key 3, but got '%s'", val)
	}

	// Test overwrite
	s.Put(1, "new_one")
	if val, ok := s.Get(1); !ok || val != "new_one" {
		t.Errorf("Expected to get 'new_one' for key 1, but got '%s'", val)
	}

	// Test Get on non-existent key
	if _, ok := s.Get(4); ok {
		t.Error("Expected to get no value for key 4, but got one")
	}
}

func TestSkipMap_Remove(t *testing.T) {
	s := New[int, string]()
	s.Put(1, "one")
	s.Put(2, "two")

	// Test Remove on existing key
	if !s.Remove(1) {
		t.Error("Expected to remove key 1, but it failed")
	}
	if _, ok := s.Get(1); ok {
		t.Error("Expected key 1 to be removed, but it still exists")
	}

	// Test Remove on non-existent key
	if s.Remove(4) {
		t.Error("Expected to fail removing non-existent key 4, but it succeeded")
	}
}

func TestSkipMap_Range(t *testing.T) {
	s := New[int, string]()
	for i := 0; i < 10; i++ {
		s.Put(i, fmt.Sprintf("val%d", i))
	}

	// Test Range
	values := s.Range(2, 5)
	expected := []string{"val2", "val3", "val4", "val5"}
	if len(values) != len(expected) {
		t.Errorf("Expected %d values, but got %d", len(expected), len(values))
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Expected value '%s' at index %d, but got '%s'", expected[i], i, v)
		}
	}
}

func TestSkipMap_RangeFunc(t *testing.T) {
	s := New[int, string]()
	for i := 0; i < 10; i++ {
		s.Put(i, fmt.Sprintf("val%d", i))
	}

	// Test RangeFunc
	var result []string
	s.RangeFunc(2, 5, func(key int, value string) bool {
		result = append(result, value)
		return true
	})

	expected := []string{"val2", "val3", "val4", "val5"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values, but got %d", len(expected), len(result))
	}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected value '%s' at index %d, but got '%s'", expected[i], i, v)
		}
	}

	// Test stopping iteration
	result = nil
	s.RangeFunc(2, 8, func(key int, value string) bool {
		if key == 5 {
			return false
		}
		result = append(result, value)
		return true
	})
	expected = []string{"val2", "val3", "val4"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values, but got %d", len(expected), len(result))
	}
}

func TestSkipMap_MinMax(t *testing.T) {
	s := New[int, string]()

	// Test on empty map
	if _, _, ok := s.Min(); ok {
		t.Error("Expected Min to fail on empty map")
	}
	if _, _, ok := s.Max(); ok {
		t.Error("Expected Max to fail on empty map")
	}

	s.Put(5, "five")
	s.Put(1, "one")
	s.Put(10, "ten")

	// Test Min
	if k, v, ok := s.Min(); !ok || k != 1 || v != "one" {
		t.Errorf("Expected min to be (1, 'one'), but got (%d, '%s')", k, v)
	}

	// Test Max
	if k, v, ok := s.Max(); !ok || k != 10 || v != "ten" {
		t.Errorf("Expected max to be (10, 'ten'), but got (%d, '%s')", k, v)
	}
}

func TestSkipMap_LenAndIsEmpty(t *testing.T) {
	s := New[int, string]()

	if !s.IsEmpty() {
		t.Error("Expected IsEmpty to be true for a new map")
	}
	if s.Len() != 0 {
		t.Errorf("Expected Len to be 0, but got %d", s.Len())
	}

	s.Put(1, "one")
	s.Put(2, "two")

	if s.IsEmpty() {
		t.Error("Expected IsEmpty to be false after adding elements")
	}
	if s.Len() != 2 {
		t.Errorf("Expected Len to be 2, but got %d", s.Len())
	}

	s.Remove(1)
	if s.Len() != 1 {
		t.Errorf("Expected Len to be 1 after removing an element, but got %d", s.Len())
	}
}

func TestSkipMap_ConcurrentAccess(t *testing.T) {
	s := New[int, string]()
	var wg sync.WaitGroup
	numGoroutines := 100
	numWritesPerG := 10

	// Concurrent Puts
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < numWritesPerG; j++ {
				key := gID*numWritesPerG + j
				s.Put(key, "val"+strconv.Itoa(key))
			}
		}(i)
	}
	wg.Wait()

	if s.Len() != numGoroutines*numWritesPerG {
		t.Errorf("Expected length to be %d, but got %d", numGoroutines*numWritesPerG, s.Len())
	}

	// Concurrent Gets and Removes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < numWritesPerG; j++ {
				key := gID*numWritesPerG + j
				if _, ok := s.Get(key); !ok {
					t.Errorf("Concurrent Get failed for key %d", key)
				}
				if j%2 == 0 {
					s.Remove(key)
				}
			}
		}(i)
	}
	wg.Wait()

	expectedLen := (numGoroutines * numWritesPerG) / 2
	if s.Len() != expectedLen {
		t.Errorf("Expected length to be %d after concurrent removes, but got %d", expectedLen, s.Len())
	}
}
