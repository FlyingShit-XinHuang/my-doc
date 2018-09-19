// Package lru provide a LRU key-value store implemented by map and doubly linked list
package lru

import (
	"log"
	"sync"
)

// KVStore is a LRU kv store
type KVStore struct {
	m     *sync.Mutex
	head  *linkedList
	tail  *linkedList
	store map[string]*linkedList
	max   int // Capacity of store
}

type linkedList struct {
	key        string
	value      string
	last, next *linkedList
}

// NewStore returns a LRU kv store with max capacity
func NewStore(max int) *KVStore {
	return &KVStore{
		m:     new(sync.Mutex),
		store: make(map[string]*linkedList, max),
		max:   max,
	}
}

// Set saves k-v pair into store
func (s *KVStore) Set(k, v string) {
	s.m.Lock()
	defer s.m.Unlock()

	l := s.store[k]
	if l == nil {
		// Create new data node if not exists
		l = &linkedList{
			key: k,
		}
		s.store[k] = l

		if len(s.store) > s.max {
			// Remove the tail
			delete(s.store, s.tail.key)
			s.tail = s.tail.last
			s.tail.next = nil
		}
	}

	// Save value and update pointer of key
	l.value = v

	s.moveToTheHeadLocked(l)
}

// Get reads the value associated with specified key
func (s *KVStore) Get(k string) (string, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	l := s.store[k]
	if nil == l {
		return "", false
	}

	s.moveToTheHeadLocked(l)
	return l.value, true
}

// Delete removes kv pair according to specified key
func (s *KVStore) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	l := s.store[k]
	if nil != l {
		delete(s.store, k)
		log.Println("deleted node:", l)
		if s.tail == l {
			// Set tail the last of data node if it was the tail
			s.tail = l.last
		}
		if s.head == l {
			// Set head the next of data node if it was the head
			s.head = l.next
		}

		if l.next != nil {
			// Link the next to the last
			l.next.last = l.last
		}
		if l.last != nil {
			// Link the last to the next
			l.last.next = l.next
		}
		l.last = nil
		l.next = nil
	}
}

// PrintLinkList is used to debug
func (s *KVStore) PrintLinkList() {
	log.Println("========== Printing linked list ==========")
	log.Printf("head pointer: %p node: %[1]v", s.head)
	log.Printf("tail pointer: %p node: %[1]v", s.tail)

	log.Print("nodes in ascending order of last used time : ")
	for l := s.tail; l != nil; l = l.last {
		log.Printf("  pointer: %p  node: %[1]v", l)
	}
	log.Println("================ Printed =================")
}

// Mutex should be locked before invoke this function
func (s *KVStore) moveToTheHeadLocked(l *linkedList) {
	// Remove data node at current position
	if l == s.tail && l.last != nil {
		// Set tail the last of data node if it was the tail
		s.tail = l.last
	}
	if l.next != nil {
		// Link the next to the last
		l.next.last = l.last
	}
	if l.last != nil {
		// Link the last to the next
		l.last.next = l.next
		l.last = nil
	}

	// Push data node into link list as head
	if nil == s.head {
		s.head, s.tail = l, l
	} else if l != s.head {
		s.head.last, s.head, l.next = l, l, s.head
	}
}
