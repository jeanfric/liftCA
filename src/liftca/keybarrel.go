package liftca

import (
	"crypto/rand"
	"crypto/rsa"
)

type barrel struct {
	workRequest chan bool
	ready       chan *rsa.PrivateKey
	keyBits     int
}

// NewBarrel returns a barrel of barrelSize size and generates keyBits-sized keys
func NewBarrel(barrelSize, keyBits int) *barrel {
	b := &barrel{
		workRequest: make(chan bool, barrelSize),
		ready:       make(chan *rsa.PrivateKey, barrelSize),
		keyBits:     keyBits,
	}

	if barrelSize == 0 {
		go processRequests(b)
		return b
	}

	for i := 0; i < barrelSize; i++ {
		go processRequests(b)
		b.workRequest <- true
	}
	return b
}

// GetKey returns a key from the barrel
func (b *barrel) GetKey() *rsa.PrivateKey {
	b.workRequest <- true
	k := <-b.ready
	return k
}

func processRequests(b *barrel) {
	for {
		<-b.workRequest
		key, err := rsa.GenerateKey(rand.Reader, b.keyBits)
		if err != nil {
			panic(err)
		}
		b.ready <- key
	}
}
