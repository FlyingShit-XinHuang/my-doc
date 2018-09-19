package lru

import (
	"strconv"
	"testing"
)

func TestOperations(t *testing.T) {
	max := 3
	s := NewStore(max)
	k := "test"

	// Insert max+1 numbers
	for i := 0; i < max; i++ {
		k := strconv.Itoa(i)
		s.Set(k, k)
	}
	s.Delete("1")
	v, ok := s.Get("1")
	if v != "" || ok {
		t.Fatal("Failed to delete value")
	}
	// s.PrintLinkList()

	for i := 0; i < max; i++ {
		k := strconv.Itoa(i)
		s.Delete(k)
	}
	// s.PrintLinkList()

	v, ok = s.Get(k)
	if v != "" || ok {
		t.Fatal("Value is not empty")
	}
	// s.PrintLinkList()

	s.Set(k, k)
	v, ok = s.Get(k)
	if v != k || !ok {
		t.Fatal("Failed to get value")
	}

	s.Set(k, "foo")
	v, ok = s.Get(k)
	if v != "foo" || !ok {
		t.Fatal("Failed to update value")
	}
	// s.PrintLinkList()

}

func TestLRU(t *testing.T) {
	max := 3
	s := NewStore(max)

	i := 0
	// Insert max+1 numbers
	for ; i < max+1; i++ {
		k := strconv.Itoa(i)
		s.Set(k, k)
	}

	v, ok := s.Get("0") // The first should be the LRU data
	if v != "" || ok {
		t.Fatal("LRU data did not be removed")
	}

	s.Get("2")
	for j := i; j < i+2; j++ {
		k := strconv.Itoa(j)
		s.Set(k, k)
	}

	k := strconv.Itoa(max + 1)
	s.Set(k, k)
	v, ok = s.Get("2")
	if v != "2" || !ok {
		t.Fatal("Data is removed by mistake")
	}
}
