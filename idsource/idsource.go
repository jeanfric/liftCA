package idsource

import (
	"math/rand"
	"sync"
)

type nothing struct{}

type IDSource struct {
	spent map[int64]nothing
	mutex *sync.Mutex
}

func New(spent []int64) *IDSource {
	s := &IDSource{
		spent: make(map[int64]nothing),
		mutex: &sync.Mutex{},
	}
	for _, v := range spent {
		s.spent[v] = nothing{}
	}
	return s
}

// Int63 returns a random number, but never the same one (it keeps track of numbers it has already given out).
func (s *IDSource) Int63() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for {
		candidate := rand.Int63()
		if _, found := s.spent[candidate]; found {
			continue
		}
		s.spent[candidate] = nothing{}
		return candidate
	}
}

func (s *IDSource) SpentIDs() []int64 {
	spent := make([]int64, 0, len(s.spent))
	for k, _ := range s.spent {
		spent = append(spent, k)
	}
	return spent
}
