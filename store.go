package liftca

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/jeanfric/liftca/idsource"
)

type Store struct {
	rw        sync.RWMutex
	idsource  *idsource.IDSource
	m         map[int64]*Parcel
	parent    map[int64]int64
	children  map[int64][]int64
	topLevel  map[int64]bool
	revoked   map[int64]bool
	listeners []chan<- struct{}
}

type gobStore struct {
	SpentIDs []int64
	M        map[int64]*Parcel
	Parent   map[int64]int64
	Children map[int64][]int64
	TopLevel map[int64]bool
	Revoked  map[int64]bool
}

func (s *Store) Updates(c chan<- struct{}) {
	s.listeners = append(s.listeners, c)
}

func (s *Store) signalUpdates() {
	for _, c := range s.listeners {
		c <- struct{}{}
	}
}

func NewStore() *Store {
	s := &Store{
		rw:        sync.RWMutex{},
		idsource:  idsource.New(make([]int64, 0)),
		m:         make(map[int64]*Parcel),
		children:  make(map[int64][]int64),
		parent:    make(map[int64]int64),
		topLevel:  make(map[int64]bool),
		revoked:   make(map[int64]bool),
		listeners: make([]chan<- struct{}, 0),
	}
	return s
}

func LoadStore(source io.Reader) *Store {
	d := &gobStore{}
	dec := gob.NewDecoder(source)
	err := dec.Decode(&d)
	if err != nil {
		// Create an empty default store if we were unable to load from the backing file.
		d.SpentIDs = make([]int64, 0)
		d.M = make(map[int64]*Parcel)
		d.Children = make(map[int64][]int64)
		d.Parent = make(map[int64]int64)
		d.TopLevel = make(map[int64]bool)
		d.Revoked = make(map[int64]bool)
	}
	s := &Store{
		rw:        sync.RWMutex{},
		idsource:  idsource.New(d.SpentIDs),
		m:         d.M,
		children:  d.Children,
		parent:    d.Parent,
		topLevel:  d.TopLevel,
		revoked:   d.Revoked,
		listeners: make([]chan<- struct{}, 0),
	}
	return s
}

func (s *Store) DumpStore(dest io.Writer) {
	s.withRLocked(func() {
		d := gobStore{
			SpentIDs: s.idsource.SpentIDs(),
			M:        s.m,
			Parent:   s.parent,
			Children: s.children,
			TopLevel: s.topLevel,
			Revoked:  s.revoked,
		}
		enc := gob.NewEncoder(dest)
		err := enc.Encode(d)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	})
}

func (s *Store) IsRevoked(id int64) bool {
	var revoked bool
	s.withRLocked(func() {
		revoked = s.revoked[id]
	})
	return revoked
}

func (s *Store) SetRevoked(id int64, revoked bool) {
	s.withLocked(func() {
		s.revoked[id] = revoked
	})
	return
}

func (s *Store) AddCA(visible bool, name string) (int64, error) {
	serial := s.idsource.Int63()
	p, err := makeCAParcel(visible, serial, name)
	if err != nil {
		return 0, err
	}
	s.withLocked(func() {
		s.m[serial] = p
		s.topLevel[serial] = true
		s.children[serial] = make([]int64, 0)
	})
	return serial, nil
}

func (s *Store) AddExistingCA(visible bool, pemCertificate []byte, pemPrivateKey []byte, pemPassword []byte) (int64, error) {
	serial := s.idsource.Int63()
	p, err := importCAFromPEM(visible, serial, pemCertificate, pemPrivateKey, pemPassword)
	if err != nil {
		return 0, err
	}

	s.withLocked(func() {
		s.m[serial] = p
		s.topLevel[serial] = true
		s.children[serial] = make([]int64, 0)
	})
	return serial, nil
}

func (s *Store) Add(visible bool, parentId int64, host string) (int64, error) {
	serial := s.idsource.Int63()
	parent, found := s.Get(parentId)
	if !found {
		return 0, fmt.Errorf("parent not found")
	}

	p, err := makeParcel(visible, serial, parent, host)
	if err != nil {
		return 0, err
	}

	s.withLocked(func() {
		s.m[serial] = p
		s.parent[serial] = parentId
		s.children[parentId] = append(s.children[parentId], serial)
	})
	return serial, nil
}

func (s *Store) Get(id int64) (*Parcel, bool) {
	var ret *Parcel = nil
	var found bool
	s.withRLocked(func() {
		var element *Parcel
		element, found = s.m[id]
		if found {
			ret = element
		}
	})
	return ret, found
}

func (s *Store) GetParent(id int64) (int64, bool) {
	var ret int64
	var found bool
	s.withRLocked(func() {
		ret, found = s.parent[id]
	})
	return ret, found
}

func (s *Store) GetChildren(id int64) ([]int64, bool) {
	var ret []int64
	var found bool
	s.withRLocked(func() {
		ret, found = s.children[id]
	})
	return ret, found
}

func (s *Store) GetRevokedChildren(id int64) []int64 {
	var revokedChildren []int64
	s.withRLocked(func() {
		rrr := make([]int64, 0)
		children, found := s.children[id]
		if found {
			for _, c := range children {
				revoked, found := s.revoked[c]
				if found && revoked {
					rrr = append(rrr, c)
				}
			}
		}
		revokedChildren = rrr
	})
	return revokedChildren
}

func (s *Store) GetCAs() []int64 {
	var ret []int64 = make([]int64, 0)
	s.withRLocked(func() {
		for k, _ := range s.topLevel {
			ret = append(ret, k)
		}
	})
	return ret
}

func (s *Store) withLocked(f func()) {
	s.rw.Lock()
	defer s.rw.Unlock()
	f()
	s.signalUpdates()
}

func (s *Store) withRLocked(f func()) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	f()
}
